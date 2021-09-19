package domain

import (
	"fmt"
	"net/http"
)

type ErrorCode int

type RestError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
}

func (e *RestError) Error() string {
	return e.Message
}

const (
	MissingParams ErrorCode = iota + 64000
	NotFound
	Internal
	InvalidParams
	BadRequest
	Timeout
)

var ErrInternalError = RestError{
	Code:       Internal,
	Message:    "internal server error",
	HTTPStatus: http.StatusInternalServerError,
}

var ErrTimeout = RestError{
	Code:       Timeout,
	Message:    "operation timeout",
	HTTPStatus: http.StatusRequestTimeout,
}

func ErrMissingParam(params ...string) *RestError {
	return &RestError{
		Code:       MissingParams,
		Message:    fmt.Sprintf("missing parameters: %v", params),
		HTTPStatus: http.StatusBadRequest,
	}
}

func InvalidParam(params ...string) *RestError {
	return &RestError{
		Code:       MissingParams,
		Message:    fmt.Sprintf("ivalid parameter value: %v", params),
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrNotFound(resource string) *RestError {
	return &RestError{
		Code:       NotFound,
		Message:    fmt.Sprintf("resource not found: %q", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

func ErrBadRequest(message string, code ErrorCode) *RestError {
	return &RestError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}
