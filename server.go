package jsonrpc

import (
	"context"
	"encoding/json"
	"sync"
)

type ErrorHandler func(error)
type ServerRequestHandler func(context.Context, Port, RequestId, json.RawMessage) (json.RawMessage, error)
type ServerNotificationHandler func(context.Context, Port, json.RawMessage) error

type ServerRequest struct {
	ctx    context.Context
	cancel context.CancelFunc
	hdrs   *HeaderSet
	id     RequestId
}

func (r *ServerRequest) Cancelled() bool {
	return (r.ctx.Err() == context.Canceled)
}

func (r *ServerRequest) Timeout() bool {
	return (r.ctx.Err() == context.DeadlineExceeded)
}

func (r *ServerRequest) Cancel() {
	r.cancel()
}

type Server struct {
	ctx          context.Context
	lock         sync.Mutex
	requests     map[RequestId]*ServerRequest
	errorHandler ErrorHandler
}

func NewServer(ctx context.Context, errorHandler ErrorHandler) *Server {
	return &Server{
		ctx:          ctx,
		requests:     make(map[RequestId]*ServerRequest),
		errorHandler: errorHandler,
	}
}

func (s *Server) Done() bool {
	return (s.ctx.Err() != nil)
}

func (s *Server) GetRequest(id RequestId) *ServerRequest {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.requests[id]
}

func (s *Server) ProcessRequest(port Port, hdrs *HeaderSet, id RequestId, params json.RawMessage, handler ServerRequestHandler) error {
	ctx, cancel := context.WithCancel(s.ctx)

	req := &ServerRequest{
		ctx:    ctx,
		cancel: cancel,
		hdrs:   hdrs,
		id:     id,
	}

	s.lock.Lock()
	s.requests[id] = req
	s.lock.Unlock()

	go func() {
		var res json.RawMessage
		var err error

		if res, err = handler(ctx, port, id, params); err != nil {
			var rpcerr Error
			var ok bool

			if rpcerr, ok = err.(Error); !ok {
				rpcerr = NewInternalError(nil)
			}

			err = SendError(port, hdrs, id, rpcerr)
		} else {
			err = sendMessage(port, hdrs, newResultMessage(id, res))
		}

		req.cancel()

		if err != nil {
			if s.errorHandler != nil {
				s.errorHandler(err)
			}
		}

		s.lock.Lock()
		delete(s.requests, id)
		s.lock.Unlock()
	}()

	return nil
}
