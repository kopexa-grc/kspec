package core

import (
	"errors"
	"fmt"
)

// SafeGo runs a function in a goroutine with panic recovery.
// Panics are silently recovered and the goroutine exits gracefully.
// For panic handling, use SafeGoWithHandler instead.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			recover() //nolint:errcheck // intentionally ignoring panic value
		}()
		fn()
	}()
}

// PanicHandler is a function that handles recovered panics.
type PanicHandler func(recovered any)

// SafeGoWithHandler runs a function in a goroutine with panic recovery.
// If a panic occurs, the handler is called with the recovered value.
func SafeGoWithHandler(fn func(), handler PanicHandler) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if handler != nil {
					handler(r)
				}
			}
		}()
		fn()
	}()
}

// Result holds the result of a SafeGoResult operation.
type Result[T any] struct {
	Value T
	Err   error
}

// SafeGoResult runs a function in a goroutine and returns its result via a channel.
// If the function panics, the error field will contain a PanicError.
// The channel is buffered with size 1, so it's safe to not read from it.
func SafeGoResult[T any](fn func() (T, error)) <-chan Result[T] {
	resultChan := make(chan Result[T], 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				var zero T
				resultChan <- Result[T]{
					Value: zero,
					Err:   RecoverError(r),
				}
			}
		}()

		value, err := fn()
		resultChan <- Result[T]{
			Value: value,
			Err:   err,
		}
	}()

	return resultChan
}

// PanicError represents an error that was recovered from a panic.
type PanicError struct {
	Value any
}

// Error implements the error interface.
func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

// RecoverError creates an error from a recovered panic value.
func RecoverError(recovered any) error {
	return &PanicError{Value: recovered}
}

// IsPanicError checks if an error is a PanicError.
func IsPanicError(err error) bool {
	var pe *PanicError
	return errors.As(err, &pe)
}
