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
	MissingParams ErrorCode = 64000
	NotFound      ErrorCode = 64001
	Internal      ErrorCode = 64002
	InvalidParams ErrorCode = 64003
	BadRequest    ErrorCode = 64003
)

var ErrInternalError = RestError{
	Code:       Internal,
	Message:    "internal server error",
	HTTPStatus: http.StatusInternalServerError,
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
