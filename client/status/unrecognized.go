package apistatus

import (
	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

// ErrUnrecognizedStatus allows to check whether some error is a NeoFS status
// unknown to the current lib version.
var ErrUnrecognizedStatus UnrecognizedStatus

// UnrecognizedStatus describes status unknown to the current lib version.
type UnrecognizedStatus struct {
	code uint32
	msg  string
	dts  []*protostatus.Status_Detail
}

// ErrUnrecognizedStatusV2 is an instance of UnrecognizedStatusV2 error status. It's expected to be used for [errors.Is]
// and MUST NOT be changed.
// Deprecated: use ErrUnrecognizedStatus instead.
var ErrUnrecognizedStatusV2 UnrecognizedStatusV2

// UnrecognizedStatusV2 describes status of the uncertain failure.
// Instances provide [StatusV2] and error interfaces.
// Deprecated: use UnrecognizedStatus instead.
type UnrecognizedStatusV2 = UnrecognizedStatus

func (x UnrecognizedStatus) Error() string {
	return errMessageStatus("unrecognized", x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x UnrecognizedStatus) Is(target error) bool {
	switch target.(type) {
	default:
		return false
	case UnrecognizedStatus, *UnrecognizedStatus:
		return true
	}
}

// implements local interface defined in [FromError] func.
func (x *UnrecognizedStatus) fromProtoMessage(st *protostatus.Status) {
	x.code = st.Code
	x.msg = st.Message
	x.dts = st.Details
}
