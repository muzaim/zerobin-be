package apperror

import (
	"fmt"
)

type Type string

const (
	TypeNotFound     Type = "NOT_FOUND"
	TypeUnauthorized Type = "UNAUTHORIZED"
	TypeForbidden    Type = "FORBIDDEN"
	TypeBadRequest   Type = "BAD_REQUEST"
	TypeConflict     Type = "CONFLICT"
	TypeInternal     Type = "INTERNAL_SERVER_ERROR"
	TypeValidation   Type = "VALIDATION_ERROR"
)

type AppError struct {
	Type    Type        `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
	Err     error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(errType Type, message string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

func NewWithDetail(errType Type, message string, detail interface{}, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Detail:  detail,
		Err:     err,
	}
}
