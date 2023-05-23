package apistatus

import (
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// ErrUnrecognizedStatusV2 is an instance of UnrecognizedStatusV2 error status. It's expected to be used for [errors.Is]
// and MUST NOT be changed.
var ErrUnrecognizedStatusV2 UnrecognizedStatusV2

// UnrecognizedStatusV2 describes status of the uncertain failure.
// Instances provide [StatusV2] and error interfaces.
type UnrecognizedStatusV2 struct {
	v2 status.Status
}

func (x UnrecognizedStatusV2) Error() string {
	return errMessageStatusV2("unrecognized", x.v2.Message())
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x UnrecognizedStatusV2) Is(target error) bool {
	switch target.(type) {
	default:
		return false
	case UnrecognizedStatusV2, *UnrecognizedStatusV2:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *UnrecognizedStatusV2) fromStatusV2(st *status.Status) {
	x.v2 = *st
}
