package client

import (
	"fmt"
)

var (
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
