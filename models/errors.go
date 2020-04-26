package models

import "fmt"

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
	_, ok := err.(*ErrNotFound)

	return ok
}
