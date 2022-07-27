package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
)

type Listener interface {
	Close()
	Accept() (Port, error)
}

type Port interface {
	Close()
	Receive() (*HeaderSet, []byte, error)
	Send(hdrs *HeaderSet, bytes []byte) error
}

func SendRequest[PA any](port Port, hdrs *HeaderSet, id RequestId, method string, params PA) error {
	var bytes []byte
	var err error

	if hdrs == nil {
		hdrs = NewHeaderSet()
	}

	if bytes, err = json.Marshal(&params); err != nil {
		return err
	}

	return sendMessage(port, hdrs, newRequestMessage(id, method, bytes))
}

func SendNotification[PA any](port Port, hdrs *HeaderSet, method string, params PA) error {
	var bytes []byte
	var err error

	if hdrs == nil {
		hdrs = NewHeaderSet()
	}

	if bytes, err = json.Marshal(&params); err != nil {
		return err
	}

	return sendMessage(port, hdrs, newNotificationMessage(method, bytes))
}

func SendResult[RE any](port Port, hdrs *HeaderSet, id RequestId, result *RE) error {
	var bytes []byte = nil
	var err error

	if hdrs == nil {
		hdrs = NewHeaderSet()
	}

	if result != nil {
		if bytes, err = json.Marshal(result); err != nil {
			return err
		}
	}

	return sendMessage(port, hdrs, newResultMessage(id, bytes))
}

func SendError(port Port, hdrs *HeaderSet, id RequestId, value Error) error {
	if hdrs == nil {
		hdrs = NewHeaderSet()
	}

	return sendMessage(port, hdrs, newErrorMessage(id, value))
}

func Receive(port Port, handler MessageHandler) (bool, error) {
	if hdrs, msg, err := receiveMessage(port); err != nil {
		return true, err
	} else {
		if hdrs == nil {
			return false, nil
		}

		if msg.isRequest() {
			return true, handler.HandleRequest(hdrs, msg.Id, msg.Method, msg.Params)
		} else if msg.isNotification() {
			return true, handler.HandleNotification(hdrs, msg.Method, msg.Params)
		} else if msg.isResponse() {
			return true, handler.HandleResponse(hdrs, msg.Id, msg.Result, msg.Error)
		}

		if msg.Id != nil {
			if err = SendError(port, hdrs, msg.Id, NewInvalidRequestError(nil)); err != nil {
				return true, err
			}
		}

		return true, fmt.Errorf("invalid message received: %s", msg.String())
	}
}

func Listen(port Port, handler MessageHandler, errorHandler ErrorHandler) {
	var cont bool = true
	var err error

	for cont {
		if cont, err = Receive(port, handler); err != nil {
			if errorHandler != nil {
				errorHandler(err)
			}
		}
	}
}

func ReceiveFromBuffered(rd *bufio.Reader) (*HeaderSet, []byte, error) {
	var hdrs *HeaderSet
	var clen int
	var bytes []byte
	var err error

	if hdrs, err = ReadHeaderSet(rd); err != nil {
		return nil, nil, err
	}

	if clen, err = hdrs.ContentLength(); err != nil {
		return hdrs, nil, err
	}

	bytes = make([]byte, clen)

	if _, err = rd.Read(bytes); err != nil {
		return hdrs, nil, err
	}

	return hdrs, bytes, nil
}

func SendToBuffered(wr *bufio.Writer, hdrs *HeaderSet, bytes []byte) error {
	hdrs.WriteTo(wr)

	if _, err := wr.Write(bytes); err != nil {
		return err
	}

	return nil
}

func receiveMessage(port Port) (*HeaderSet, message, error) {
	var msg message

	if hdrs, bytes, err := port.Receive(); err != nil {
		return hdrs, msg, err
	} else {
		if (hdrs == nil) && (bytes == nil) {
			return nil, msg, nil
		}

		if err := json.Unmarshal(bytes, &msg); err != nil {
			return hdrs, msg, err
		}

		return hdrs, msg, nil
	}
}

func sendMessage(port Port, hdrs *HeaderSet, msg message) error {
	if bytes, err := json.Marshal(&msg); err != nil {
		return err
	} else {
		hdrs.SetContentLength(len(bytes))
		return port.Send(hdrs, bytes)
	}
}
