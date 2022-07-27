package jsonrpc

import (
	"encoding/json"
	"fmt"
	"sync"
)

type ClientResult struct {
	Result json.RawMessage
	Error  *Error
}

type ClientRequest struct {
	client *Client
	id     RequestId
	result chan ClientResult
}

func (r *ClientRequest) Close() {
	close(r.result)
	r.client.remove(r)
}

func (r *ClientRequest) Done() bool {
	select {
	case <-r.result:
		return true
	default:
		return false
	}
}

func (r *ClientRequest) GetResult() ClientResult {
	return <-r.result
}

type Client struct {
	lock     sync.Mutex
	nextId   int
	requests map[RequestId]*ClientRequest
}

func NewClient() *Client {
	return &Client{
		nextId:   0,
		requests: make(map[RequestId]*ClientRequest),
	}
}

func (c *Client) NewRequestId() RequestId {
	c.lock.Lock()
	defer c.lock.Unlock()

	res := c.nextId
	c.nextId += 1
	return res
}

func SendClientRequest[PA any](client *Client, port Port, hdrs *HeaderSet, id RequestId, method string, params PA) (*ClientRequest, error) {
	var req *ClientRequest = &ClientRequest{
		id:     id,
		result: make(chan ClientResult),
	}

	client.lock.Lock()
	client.requests[id] = req
	client.lock.Unlock()

	if err := SendRequest(port, hdrs, id, method, params); err != nil {
		client.lock.Lock()
		delete(client.requests, id)
		client.lock.Unlock()

		return nil, err
	}

	return req, nil
}

func (c *Client) ProcessResponse(id RequestId, result json.RawMessage, err *Error) error {
	var req *ClientRequest
	var res ClientResult

	c.lock.Lock()
	req = c.requests[id]
	c.lock.Unlock()

	if req == nil {
		return fmt.Errorf("request '%v' not found or already closed", id)
	}

	if err != nil {
		res = ClientResult{
			Error: err,
		}
	} else {
		res = ClientResult{
			Result: result,
		}
	}

	req.result <- res
	return nil
}

func (c *Client) remove(r *ClientRequest) {
	c.lock.Lock()
	delete(c.requests, r.id)
	c.lock.Unlock()
}
