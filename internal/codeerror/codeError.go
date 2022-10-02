package codeerror

import (
	"net/http"

	"github.com/jackc/pgx"
)

// CodeError type
type CodeError struct {
	message   string
	status    int
	code      string
	qualifier string
}

func (e CodeError) Error() string {
	return e.message
}

// Code function
func (e CodeError) Code() string {
	return e.code
}

// New function
func New(message string, status int, code string, qualifier string) *CodeError {
	return &CodeError{message: message, status: status, code: code, qualifier: qualifier}
}

// NewInternalServerError function
func NewInternalServerError(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusInternalServerError}
}

// NewDatabaseError function
func NewDatabaseError(x pgx.PgError) *CodeError {
	return &CodeError{message: x.Message, status: http.StatusBadRequest, code: x.Code, qualifier: x.ConstraintName}
}

// NewBadRequest function
func NewBadRequest(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusBadRequest}
}

// NewNotFound function
func NewNotFound(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusNotFound}
}

// NewForbidden function
func NewForbidden(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusForbidden}
}

// NewUnauthorized function
func NewUnauthorized(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusUnauthorized}
}

// NewUnauthorizedJWTExpired function
func NewUnauthorizedJWTExpired(message string) *CodeError {
	return &CodeError{message: message, status: http.StatusUnauthorized}
}
