package client

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingServer is returned when server endpoint is empty in parameters.
	ErrMissingServer = errors.New("server address is unset or empty")
	// ErrNonPositiveTimeout is returned when any timeout is below zero in parameters.
	ErrNonPositiveTimeout = errors.New("non-positive timeout")

	// ErrMissingContainer is returned when container is not provided.
	ErrMissingContainer = errors.New("missing container")
	// ErrMissingObject is returned when object is not provided.
	ErrMissingObject = errors.New("missing object")
	// ErrMissingAccount is returned when account/owner is not provided.
	ErrMissingAccount = errors.New("missing account")
	// ErrMissingEACL is returned when eACL table is not provided.
	ErrMissingEACL = errors.New("missing eACL table")
	// ErrMissingEACLContainer is returned when container info is not provided in eACL table.
	ErrMissingEACLContainer = errors.New("missing container in eACL table")
	// ErrMissingAnnouncements is returned when announcements are not provided.
	ErrMissingAnnouncements = errors.New("missing announcements")
	// ErrZeroRangeLength is returned when range parameter has zero length.
	ErrZeroRangeLength = errors.New("zero range length")
	// ErrMissingRanges is returned when empty ranges list is provided.
	ErrMissingRanges = errors.New("missing ranges")
	// ErrZeroEpoch is returned when zero epoch is provided.
	ErrZeroEpoch = errors.New("zero epoch")
	// ErrMissingTrusts is returned when empty slice of trusts is provided.
	ErrMissingTrusts = errors.New("missing trusts")
	// ErrMissingTrust is returned when empty trust is not provided.
	ErrMissingTrust = errors.New("missing trust")

	// ErrUnexpectedReadCall is returned when we already got all data but truing to get more.
	ErrUnexpectedReadCall = errors.New("unexpected call to `Read`")

	// ErrSign is returned when unable to sign service message.
	ErrSign SignError

	errMissingResponseField missingResponseFieldErr
)

type missingResponseFieldErr struct {
	name string
}

func (e missingResponseFieldErr) Error() string {
	return fmt.Sprintf("missing %s field in the response", e.name)
}

func (e missingResponseFieldErr) Is(target error) bool {
	switch target.(type) {
	default:
		return false
	case missingResponseFieldErr, *missingResponseFieldErr:
		return true
	}
}

// returns error describing missing field with the given name.
func newErrMissingResponseField(name string) error {
	return missingResponseFieldErr{name: name}
}

// returns error describing invalid field (according to the NeoFS protocol)
// with the given name and format violation err.
func newErrInvalidResponseField(name string, err error) error {
	return fmt.Errorf("invalid %s field in the response: %w", name, err)
}

// SignError wraps another error with reason why sign process was failed.
type SignError struct {
	err error
}

// NewSignError is a constructor for [SignError].
func NewSignError(err error) SignError {
	return SignError{err: err}
}

// Error implements the error interface.
func (e SignError) Error() string {
	return fmt.Sprintf("sign: %v", e.err)
}

// Unwrap implements the error interface.
func (e SignError) Unwrap() error {
	return e.err
}

// Is implements interface for correct checking current error type with [errors.Is].
func (e SignError) Is(target error) bool {
	switch target.(type) {
	default:
		return false
	case SignError, *SignError:
		return true
	}
}
