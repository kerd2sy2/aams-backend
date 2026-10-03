package errors

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden access")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrConflict          = errors.New("resource conflict or duplicate")
	ErrInternalServer    = errors.New("internal server error")
	ErrInsufficientStock = errors.New("insufficient stock")
)
