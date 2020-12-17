package core

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidName is returned by Env when the bind name is
	// invalid.
	ErrInvalidName = errors.New("invalid bind name")

	// ErrIncomparableTypes is returned if two types cannot be meaningfully
	// compared to each other.
	ErrIncomparableTypes = errors.New("incomparable types")

	// ErrIndexOutOfBounds is returned when a sequence's index is out of range.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// ErrNotFound is returned by Env when a binding is not found
	// for a given symbol/name.
	ErrNotFound = errors.New("not found")

	// ErrArity is returned when an Invokable is invoked with wrong number
	// of arguments.
	ErrArity = errors.New("wrong number of arguments")

	// ErrNotInvokable is returned by InvokeExpr when the target is not invokable.
	ErrNotInvokable = errors.New("non-invokable type")

	// ErrIllegalState is returned when an operation is attempted against a datatype
	// that has the right type but an inappropriate value.
	ErrIllegalState = errors.New("illegal state")

	// ErrMemory is returned when an operation is attempted against an illegally
	// formatted datatype.
	ErrMemory = errors.New("memory error")
)

// Error is returned by all core operations. Cause indicates the underlying
// error type. Use errors.Is() with Cause to check for specific errors.
type Error struct {
	Message    string
	Cause      error
	Form       string
	Begin, End Position
}

// With returns a clone of the error with message set to given value.
func (e Error) With(msg string) Error {
	return Error{
		Cause:   e.Cause,
		Message: msg,
	}
}

// Is returns true if the other error is same as the cause of this error.
func (e Error) Is(other error) bool { return errors.Is(e.Cause, other) }

// Unwrap returns the underlying cause of the error.
func (e Error) Unwrap() error { return e.Cause }

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%v: %s", e.Cause, e.Message)
	}
	return e.Message
}

// Position represents the positional information about a value read
// by reader.
type Position struct {
	File string
	Ln   int
	Col  int
}

func (p Position) String() string {
	if p.File == "" {
		p.File = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Ln, p.Col)
}
