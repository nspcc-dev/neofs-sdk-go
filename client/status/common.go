package apistatus

import (
	"encoding/binary"
	"errors"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
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
	// ErrBadRequest is an instance of BadRequest error status. It's
	// expected to be used for [errors.Is] and MUST NOT be changed.
	ErrBadRequest BadRequest
	// ErrBusy is an instance of Busy error status. It's expected
	// to be used for [errors.Is] and MUST NOT be changed.
	ErrBusy Busy
)

// ServerInternal describes failure statuses related to internal server errors.
//
// The status is purely informative, the client should not go into details of the error except for debugging needs.
type ServerInternal struct {
	msg string
	dts []*protostatus.Status_Detail
}

func (x ServerInternal) Error() string {
	return errMessageStatus(protostatus.InternalServerError, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *ServerInternal) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ServerInternal) protoMessage() *protostatus.Status {
	return &protostatus.Status{Code: protostatus.InternalServerError, Message: x.msg, Details: x.dts}
}

// SetMessage sets message describing internal error.
//
// Message should be used for debug purposes only.
func (x *ServerInternal) SetMessage(msg string) {
	x.msg = msg
}

// Message returns message describing internal server error.
//
// Message should be used for debug purposes only. By default, it is empty.
func (x ServerInternal) Message() string {
	return x.msg
}

// WriteInternalServerErr writes err message to ServerInternal instance.
func WriteInternalServerErr(x *ServerInternal, err error) {
	x.SetMessage(err.Error())
}

// WrongMagicNumber describes failure status related to incorrect network magic.
// Instances provide [StatusV2] and error interfaces.
type WrongMagicNumber struct {
	msg string
	dts []*protostatus.Status_Detail
}

func (x WrongMagicNumber) Error() string {
	return errMessageStatus(protostatus.WrongNetMagic, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *WrongMagicNumber) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x WrongMagicNumber) protoMessage() *protostatus.Status {
	return &protostatus.Status{Code: protostatus.WrongNetMagic, Message: x.msg, Details: x.dts}
}

// WriteCorrectMagic writes correct network magic.
func (x *WrongMagicNumber) WriteCorrectMagic(magic uint64) {
	// serialize the number
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, magic)

	for i := range x.dts {
		if x.dts[i].Id == protostatus.DetailCorrectNetMagic {
			x.dts[i].Value = buf
			return
		}
	}
	x.dts = append(x.dts, &protostatus.Status_Detail{
		Id:    protostatus.DetailCorrectNetMagic,
		Value: buf,
	})
}

// CorrectMagic returns network magic returned by the server.
// Second value indicates presence status:
//   - -1 if number is presented in incorrect format
//   - 0 if number is not presented
//   - +1 otherwise
func (x WrongMagicNumber) CorrectMagic() (uint64, int8) {
	for i := range x.dts {
		if x.dts[i].Id == protostatus.DetailCorrectNetMagic {
			if len(x.dts[i].Value) == 8 {
				return binary.BigEndian.Uint64(x.dts[i].Value), 1
			}
			return 0, -1
		}
	}
	return 0, 0
}

// SignatureVerification describes failure status related to signature verification.
type SignatureVerification struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultSignatureVerificationMsg = "signature verification failed"

func (x SignatureVerification) Error() string {
	if x.msg == "" {
		x.msg = defaultSignatureVerificationMsg
	}

	return errMessageStatus(protostatus.SignatureVerificationFail, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *SignatureVerification) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x SignatureVerification) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultSignatureVerificationMsg
	}
	return &protostatus.Status{Code: protostatus.SignatureVerificationFail, Message: x.msg, Details: x.dts}
}

// SetMessage writes signature verification failure message.
// Message should be used for debug purposes only.
//
// See also Message.
func (x *SignatureVerification) SetMessage(v string) {
	x.msg = v
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also SetMessage.
func (x SignatureVerification) Message() string {
	return x.msg
}

// NodeUnderMaintenance describes failure status for nodes being under maintenance.
type NodeUnderMaintenance struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultNodeUnderMaintenanceMsg = "node is under maintenance"

// Error implements the error interface.
func (x NodeUnderMaintenance) Error() string {
	if x.msg == "" {
		x.msg = defaultNodeUnderMaintenanceMsg
	}

	return errMessageStatus(protostatus.NodeUnderMaintenance, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *NodeUnderMaintenance) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x NodeUnderMaintenance) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultNodeUnderMaintenanceMsg
	}
	return &protostatus.Status{Code: protostatus.NodeUnderMaintenance, Message: x.msg, Details: x.dts}
}

// SetMessage writes signature verification failure message.
// Message should be used for debug purposes only.
//
// See also Message.
func (x *NodeUnderMaintenance) SetMessage(v string) {
	x.msg = v
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also SetMessage.
func (x NodeUnderMaintenance) Message() string {
	return x.msg
}

// BadRequest describes failure status for requests that don't follow
// the protocol.
type BadRequest struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultBadRequestMsg = "bad request"

// Error implements the error interface.
func (x BadRequest) Error() string {
	if x.msg == "" {
		x.msg = defaultBadRequestMsg
	}

	return errMessageStatus(protostatus.BadRequest, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x BadRequest) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case BadRequest, *BadRequest:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *BadRequest) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x BadRequest) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultBadRequestMsg
	}
	return &protostatus.Status{Code: protostatus.BadRequest, Message: x.msg, Details: x.dts}
}

// SetMessage sets details for why the request is bad.
// Message should be used for debug purposes only.
//
// See also [Message].
func (x *BadRequest) SetMessage(v string) {
	x.msg = v
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also [SetMessage].
func (x BadRequest) Message() string {
	return x.msg
}

// Busy describes failure status for servers that currently incapable of
// performing the requested call, but are likely to be able to do that
// in future.
type Busy struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultBusyMsg = "busy, retry later"

// Error implements the error interface.
func (x Busy) Error() string {
	if x.msg == "" {
		x.msg = defaultBusyMsg
	}

	return errMessageStatus(protostatus.Busy, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x Busy) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case Busy, *Busy:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *Busy) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x Busy) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultBusyMsg
	}
	return &protostatus.Status{Code: protostatus.Busy, Message: x.msg, Details: x.dts}
}

// SetMessage sets details of server state.
// Message should be used for debug purposes only.
//
// See also [Message].
func (x *Busy) SetMessage(v string) {
	x.msg = v
}

// Message returns status message. Zero status returns empty message.
// Message should be used for debug purposes only.
//
// See also [SetMessage].
func (x Busy) Message() string {
	return x.msg
}
