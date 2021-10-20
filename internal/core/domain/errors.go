package domain

import (
	"fmt"
	"net/http"
)

type ErrorCode int

// RestError is used to normalize errors return to consumers using REST API
type RestError struct {
	// Internal error code
	Code ErrorCode `json:"code"`
	// A human friendly error message
	Message string `json:"message"`
	// The error HTTP status code
	HTTPStatus int `json:"-"`
}

func (e *RestError) Error() string {
	return e.Message
}

// Internal error codes
const (
	MissingParams ErrorCode = iota + 64000
	NotFound
	Internal
	InvalidParams
	BadRequest
	Timeout
)

// Return this for any unknown/unhandled error
var ErrInternalError = RestError{
	Code:       Internal,
	Message:    "internal server error",
	HTTPStatus: http.StatusInternalServerError,
}

// Return this for any request that can't be proccessed on time
var ErrTimeout = RestError{
	Code:       Timeout,
	Message:    "operation timeout",
	HTTPStatus: http.StatusRequestTimeout,
}

// Creates a new error for a list of missing params
func ErrMissingParam(params ...string) *RestError {
	return &RestError{
		Code:       MissingParams,
		Message:    fmt.Sprintf("missing parameters: %v", params),
		HTTPStatus: http.StatusBadRequest,
	}
}

// Creates a new error for a list of invalid params
func InvalidParam(params ...string) *RestError {
	return &RestError{
		Code:       MissingParams,
		Message:    fmt.Sprintf("ivalid parameter value: %v", params),
		HTTPStatus: http.StatusBadRequest,
	}
}

// Creates a new error for a resource not present in server
func ErrNotFound(resource string) *RestError {
	return &RestError{
		Code:       NotFound,
		Message:    fmt.Sprintf("resource not found: %q", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// Create a new error with the given message and BadRequest internal code
func ErrBadRequest(message string) *RestError {
	return &RestError{
		Code:       BadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}
