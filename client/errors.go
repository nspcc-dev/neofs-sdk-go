package client

import (
	"errors"
	"fmt"
)

var (
	ErrMissingServer      = errors.New("server address is unset or empty")
	ErrNonPositiveTimeout = errors.New("non-positive timeout")

	ErrMissingContainer     = errors.New("missing container")
	ErrMissingObject        = errors.New("missing object")
	ErrMissingAccount       = errors.New("missing account")
	ErrMissingEACL          = errors.New("missing eACL table")
	ErrMissingEACLContainer = errors.New("missing container in eACL table")
	ErrMissingAnnouncements = errors.New("missing announcements")
	ErrZeroRangeLength      = errors.New("zero range length")
	ErrMissingRanges        = errors.New("missing ranges")
	ErrZeroEpoch            = errors.New("zero epoch")
	ErrMissingTrusts        = errors.New("missing trusts")
	ErrMissingTrust         = errors.New("missing trust")

	ErrUnexpectedReadCall = errors.New("unexpected call to `Read`")

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

type SignError struct {
	err error
}

func NewSignError(err error) SignError {
	return SignError{err: err}
}

func (e SignError) Error() string {
	return fmt.Sprintf("sign: %v", e.err)
}

func (e SignError) Unwrap() error {
	return e.err
}

func (e SignError) Is(target error) bool {
	switch target.(type) {
	default:
		return false
	case SignError, *SignError:
		return true
	}
}
