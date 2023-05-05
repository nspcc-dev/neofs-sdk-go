package client

import (
	"fmt"
)

// returns error describing missing field with the given name.
func newErrMissingResponseField(name string) error {
	return fmt.Errorf("missing %s field in the response", name)
}

// returns error describing invalid field (according to the NeoFS protocol)
// with the given name and format violation err.
func newErrInvalidResponseField(name string, err error) error {
	return fmt.Errorf("invalid %s field in the response: %w", name, err)
}
