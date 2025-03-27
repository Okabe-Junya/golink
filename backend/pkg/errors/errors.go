package errors

import (
	"errors"
	"fmt"
)

// Standard errors
var (
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInternalServer = errors.New("internal server error")
	ErrAlreadyExists  = errors.New("already exists")
)

// Error is a custom error type with status code
type Error struct {
	Err     error
	Message string
	Code    int
}

// Error returns the error message
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// NewNotFound creates a new not found error
func NewNotFound(message string) *Error {
	return &Error{
		Code:    404,
		Message: message,
		Err:     ErrNotFound,
	}
}

// NewBadRequest creates a new bad request error
func NewBadRequest(message string) *Error {
	return &Error{
		Code:    400,
		Message: message,
		Err:     ErrBadRequest,
	}
}

// NewUnauthorized creates a new unauthorized error
func NewUnauthorized(message string) *Error {
	return &Error{
		Code:    401,
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewForbidden creates a new forbidden error
func NewForbidden(message string) *Error {
	return &Error{
		Code:    403,
		Message: message,
		Err:     ErrForbidden,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(err error) *Error {
	return &Error{
		Code:    500,
		Message: "An internal server error occurred",
		Err:     err,
	}
}

// NewAlreadyExists creates a new already exists error
func NewAlreadyExists(message string) *Error {
	return &Error{
		Code:    409,
		Message: message,
		Err:     ErrAlreadyExists,
	}
}

// Wrap wraps an error with additional message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is reports whether any error in err's chain matches target
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetStatusCode returns the status code for an error
func GetStatusCode(err error) int {
	var e *Error
	if As(err, &e) {
		return e.Code
	}
	return 500
}
