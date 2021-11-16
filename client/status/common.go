package apistatus

import (
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// ServerInternal describes failure statuses related to internal server errors.
// Instances provide Status and StatusV2 interfaces.
//
// The status is purely informative, the client should not go into details of the error except for debugging needs.
type ServerInternal struct {
	v2 status.Status
}

func (x ServerInternal) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(status.Internal, status.GlobalizeCommonFail),
		x.v2.Message(),
	)
}

// implements method of the FromStatusV2 local interface.
func (x *ServerInternal) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: INTERNAL;
//  * string message: empty;
//  * details: empty.
func (x ServerInternal) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(status.Internal, status.GlobalizeCommonFail))
	return &x.v2
}

// SetMessage sets message describing internal error.
//
// Message should be used for debug purposes only.
func (x *ServerInternal) SetMessage(msg string) {
	x.v2.SetMessage(msg)
}

// Message returns message describing internal server error.
//
// Message should be used for debug purposes only. By default, it is empty.
func (x ServerInternal) Message() string {
	return x.v2.Message()
}

// WriteInternalServerErr writes err message to ServerInternal instance.
func WriteInternalServerErr(x *ServerInternal, err error) {
	x.SetMessage(err.Error())
}
