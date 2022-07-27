package jsonrpc

import "encoding/json"

type ErrorCode int

const (
	ErrorCode_ParseError     ErrorCode = -32700
	ErrorCode_InvalidRequest ErrorCode = -32600
	ErrorCode_MethodNotFound ErrorCode = -32601
	ErrorCode_InvalidParams  ErrorCode = -32602
	ErrorCode_InternalError  ErrorCode = -32603
)

type Error struct {
	Code    ErrorCode       `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

func NewError(code ErrorCode, message string, data json.RawMessage) Error {
	return Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func NewErrorWithData[DA any](code ErrorCode, message string, data DA) Error {
	var bytes json.RawMessage

	bytes, _ = json.Marshal(&data)

	return NewError(code, message, bytes)
}

func NewParseError(data json.RawMessage) Error {
	return Error{
		Code:    ErrorCode_ParseError,
		Message: "Parse error",
		Data:    data,
	}
}

func NewInvalidRequestError(data json.RawMessage) Error {
	return Error{
		Code:    ErrorCode_InvalidRequest,
		Message: "Invalid Request",
		Data:    data,
	}
}

func NewMethodNotFoundError(data json.RawMessage) Error {
	return Error{
		Code:    ErrorCode_MethodNotFound,
		Message: "Method not found",
		Data:    data,
	}
}

func NewInvalidParamsError(data json.RawMessage) Error {
	return Error{
		Code:    ErrorCode_InvalidParams,
		Message: "Invalid params",
		Data:    data,
	}
}

func NewInternalError(data json.RawMessage) Error {
	return Error{
		Code:    ErrorCode_InternalError,
		Message: "Internal error",
		Data:    data,
	}
}
