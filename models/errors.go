package models

import (
	"fmt"

	"github.com/olivere/elastic/v7"
)

type RequireNewPasswordError struct {
	Session string
}

func (e *RequireNewPasswordError) Error() string {
	return "User Require to change temporary password"
}

func NewRequireNewPasswordError(session string) *RequireNewPasswordError {
	return &RequireNewPasswordError{Session: session}
}

// ErrNotFound is error type that denotes that everything worked correctly but value was not found
type ErrNotFound struct {
	msg string
}

// Error so that ErrNotFound implements error interface
func (e *ErrNotFound) Error() string {
	return e.msg
}

// NewErrNotFound is constructor for ErrNotFound
func NewErrNotFound(format string, a ...interface{}) *ErrNotFound {
	return &ErrNotFound{
		msg: fmt.Sprintf(format, a...),
	}
}

// IsErrNotFound returns true if error is ErrNotFound
func IsErrNotFound(err error) bool {
	if elastic.IsNotFound(err) {
		return true
	}

	_, ok := err.(*ErrNotFound)

	return ok
}

// ErrInvalidData is error type that denotes that everything worked correctly but value was not found
type ErrInvalidData struct {
	msg string
}

// Error so that ErrInvalidData implements error interface
func (e *ErrInvalidData) Error() string {
	return e.msg
}

// NewErrInvalidData is constructor for ErrNotFound
func NewErrInvalidData(format string, a ...interface{}) *ErrInvalidData {
	return &ErrInvalidData{
		msg: fmt.Sprintf(format, a...),
	}
}

// IsErrInvalidData returns true if error is ErrNotFound
func IsErrInvalidData(err error) bool {
	_, ok := err.(*ErrInvalidData)

	return ok
}
