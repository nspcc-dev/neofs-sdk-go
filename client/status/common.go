package apistatus

import (
	"encoding/binary"
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// Error describes common error which is a grouping type for any [apistatus] errors. Any [apistatus] error may be checked
// explicitly via it's type of just check the group via errors.Is(err, [apistatus.Error]).
var Error = errors.New("api error")

var (
	// ErrServerInternal is an instance of ServerInternal error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrServerInternal ServerInternal
	// ErrWrongMagicNumber is an instance of WrongMagicNumber error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrWrongMagicNumber WrongMagicNumber
	// ErrSignatureVerification is an instance of SignatureVerification error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrSignatureVerification SignatureVerification
	// ErrNodeUnderMaintenance is an instance of NodeUnderMaintenance error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrNodeUnderMaintenance NodeUnderMaintenance
)

// ServerInternal describes failure statuses related to internal server errors.
// Instances provide [StatusV2] and error interfaces.
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

// Is implements interface for correct checking current error type with [errors.Is].
func (x ServerInternal) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ServerInternal, *ServerInternal:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *ServerInternal) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: INTERNAL;
//   - string message: empty;
//   - details: empty.
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

// WrongMagicNumber describes failure status related to incorrect network magic.
// Instances provide [StatusV2] and error interfaces.
type WrongMagicNumber struct {
	v2 status.Status
}

func (x WrongMagicNumber) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(status.WrongMagicNumber, status.GlobalizeCommonFail),
		x.v2.Message(),
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x WrongMagicNumber) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case WrongMagicNumber, *WrongMagicNumber:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *WrongMagicNumber) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: WRONG_MAGIC_NUMBER;
//   - string message: empty;
//   - details: empty.
func (x WrongMagicNumber) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(status.WrongMagicNumber, status.GlobalizeCommonFail))
	return &x.v2
}

// WriteCorrectMagic writes correct network magic.
func (x *WrongMagicNumber) WriteCorrectMagic(magic uint64) {
	// serialize the number
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, magic)

	// create corresponding detail
	var d status.Detail

	d.SetID(status.DetailIDCorrectMagic)
	d.SetValue(buf)

	// attach the detail
	x.v2.AppendDetails(d)
}

// CorrectMagic returns network magic returned by the server.
// Second value indicates presence status:
//   - -1 if number is presented in incorrect format
//   - 0 if number is not presented
//   - +1 otherwise
func (x WrongMagicNumber) CorrectMagic() (magic uint64, ok int8) {
	x.v2.IterateDetails(func(d *status.Detail) bool {
		if d.ID() == status.DetailIDCorrectMagic {
			if val := d.Value(); len(val) == 8 {
				magic = binary.BigEndian.Uint64(val)
				ok = 1
			} else {
				ok = -1
			}
		}

		return ok != 0
	})

	return
}

// SignatureVerification describes failure status related to signature verification.
// Instances provide [StatusV2] and error interfaces.
type SignatureVerification struct {
	v2 status.Status
}

const defaultSignatureVerificationMsg = "signature verification failed"

func (x SignatureVerification) Error() string {
	msg := x.v2.Message()
	if msg == "" {
		msg = defaultSignatureVerificationMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(status.SignatureVerificationFail, status.GlobalizeCommonFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x SignatureVerification) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case SignatureVerification, *SignatureVerification:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *SignatureVerification) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: SIGNATURE_VERIFICATION_FAIL;
//   - string message: written message via SetMessage or
//     "signature verification failed" as a default message;
//   - details: empty.
func (x SignatureVerification) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(status.SignatureVerificationFail, status.GlobalizeCommonFail))

	if x.v2.Message() == "" {
		x.v2.SetMessage(defaultSignatureVerificationMsg)
	}

	return &x.v2
}

// SetMessage writes signature verification failure message.
// Message should be used for debug purposes only.
//
// See also Message.
func (x *SignatureVerification) SetMessage(v string) {
	x.v2.SetMessage(v)
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also SetMessage.
func (x SignatureVerification) Message() string {
	return x.v2.Message()
}

// NodeUnderMaintenance describes failure status for nodes being under maintenance.
// Instances provide [StatusV2] and error interfaces.
type NodeUnderMaintenance struct {
	v2 status.Status
}

const defaultNodeUnderMaintenanceMsg = "node is under maintenance"

// Error implements the error interface.
func (x NodeUnderMaintenance) Error() string {
	msg := x.Message()
	if msg == "" {
		msg = defaultNodeUnderMaintenanceMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(status.NodeUnderMaintenance, status.GlobalizeCommonFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x NodeUnderMaintenance) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case NodeUnderMaintenance, *NodeUnderMaintenance:
		return true
	}
}

func (x *NodeUnderMaintenance) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: NODE_UNDER_MAINTENANCE;
//   - string message: written message via SetMessage or
//     "node is under maintenance" as a default message;
//   - details: empty.
func (x NodeUnderMaintenance) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(status.NodeUnderMaintenance, status.GlobalizeCommonFail))
	if x.v2.Message() == "" {
		x.v2.SetMessage(defaultNodeUnderMaintenanceMsg)
	}

	return &x.v2
}

// SetMessage writes signature verification failure message.
// Message should be used for debug purposes only.
//
// See also Message.
func (x *NodeUnderMaintenance) SetMessage(v string) {
	x.v2.SetMessage(v)
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also SetMessage.
func (x NodeUnderMaintenance) Message() string {
	return x.v2.Message()
}
