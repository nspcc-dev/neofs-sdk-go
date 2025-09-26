package apistatus

import (
	"errors"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

var (
	// ErrIncomplete is an instance of Incomplete error status. It's
	// expected to be used for [errors.Is] and MUST NOT be changed.
	ErrIncomplete Incomplete
)

// Incomplete describes partially successful status when some work has been
// done, but the result can't be considered completely fine. Details are
// available in the associated message and client should figure out if
// this status is sufficient for it to consider operation to be successful.
type Incomplete struct {
	msg string
	dts []*protostatus.Status_Detail
}

func (x Incomplete) Error() string {
	return errMessageStatus(protostatus.IncompleteSuccess, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x Incomplete) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case Incomplete, *Incomplete:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *Incomplete) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x Incomplete) protoMessage() *protostatus.Status {
	return &protostatus.Status{Code: protostatus.IncompleteSuccess, Message: x.msg, Details: x.dts}
}

// SetMessage sets message describing incomplete success details.
//
// Message should be used for debug purposes only.
func (x *Incomplete) SetMessage(msg string) {
	x.msg = msg
}

// Message returns message describing incomplete success details.
//
// Message should be used for debug purposes only. By default, it is empty.
func (x Incomplete) Message() string {
	return x.msg
}

// WriteIncompleteErr writes err message to Incomplete instance.
func WriteIncompleteErr(x *Incomplete, err error) {
	x.SetMessage(err.Error())
}
