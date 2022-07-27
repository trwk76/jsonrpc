package jsonrpc

import "encoding/json"

type RequestId interface{}

type MessageHandler interface {
	HandleRequest(hdrs *HeaderSet, id RequestId, method string, params json.RawMessage) error
	HandleNotification(hdrs *HeaderSet, method string, params json.RawMessage) error
	HandleResponse(hdrs *HeaderSet, id RequestId, result json.RawMessage, err *Error) error
}

type message struct {
	JsonRpc string          `json:"jsonrpc"`
	Id      RequestId       `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

const currentVersion string = "2.0"

func newRequestMessage(id RequestId, method string, params json.RawMessage) message {
	return message{
		JsonRpc: currentVersion,
		Id:      id,
		Method:  method,
		Params:  params,
	}
}

func newNotificationMessage(method string, params json.RawMessage) message {
	return message{
		JsonRpc: currentVersion,
		Method:  method,
		Params:  params,
	}
}

func newResultMessage(id RequestId, result json.RawMessage) message {
	return message{
		JsonRpc: currentVersion,
		Id:      id,
		Result:  result,
	}
}

func newErrorMessage(id RequestId, err Error) message {
	return message{
		JsonRpc: currentVersion,
		Id:      id,
		Error:   &err,
	}
}

func (m message) isRequest() bool {
	return (m.Id != nil) && (m.Method != "") && (m.Result == nil) && (m.Error == nil)
}

func (m message) isNotification() bool {
	return (m.Id == nil) && (m.Method != "") && (m.Result == nil) && (m.Error == nil)
}

func (m message) isResponse() bool {
	return (m.Id != nil) && ((m.Result != nil) || (m.Error != nil))
}

func (m message) String() string {
	if bytes, err := json.MarshalIndent(&m, "", "\t"); err != nil {
		return "<serialization error>"
	} else {
		return string(bytes)
	}
}
