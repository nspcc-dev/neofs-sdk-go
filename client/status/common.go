package apistatus

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

// Error describes common error which is a grouping type for any [apistatus] errors. Any [apistatus] error may be checked
// explicitly via it's type of just check the group via errors.Is(err, [apistatus.Error]).
var Error = errors.New("api error")

// Common error instances which may be used to check API errors against using
// [errors.Is]. All of them MUST NOT be changed.
var (
	ErrServerInternal        InternalServerError
	ErrWrongNetMagic         WrongNetMagic
	ErrSignatureVerification SignatureVerificationFailure
	ErrNodeUnderMaintenance  NodeUnderMaintenance
)

// InternalServerError describes failure statuses related to internal server
// errors.
//
// The status is purely informative, the client should not go into details of the error except for debugging needs.
type InternalServerError string

// NewInternalServerError constructs internal server error with specified cause.
func NewInternalServerError(cause error) InternalServerError {
	return InternalServerError(cause.Error())
}

// Error implements built-in error interface.
func (x InternalServerError) Error() string {
	const desc = "internal server error"
	if x != "" {
		return fmt.Sprintf(errFmt, status.InternalServerError, desc, string(x))
	}
	return fmt.Sprintf(errFmtNoMessage, status.InternalServerError, desc)
}

// Is checks whether target is of type InternalServerError, *InternalServerError
// or [Error]. Is implements interface consumed by [errors.Is].
func (x InternalServerError) Is(target error) bool { return errorIs(x, target) }

func (x *InternalServerError) readFromV2(m *status.Status) error {
	if m.Code != status.InternalServerError {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.InternalServerError))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	*x = InternalServerError(m.Message)
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x InternalServerError) ErrorToV2() *status.Status {
	return &status.Status{Code: status.InternalServerError, Message: string(x)}
}

// WrongNetMagic describes failure status related to incorrect network magic.
type WrongNetMagic struct {
	correctSet bool
	correct    uint64
	msg        string
}

// NewWrongNetMagicError constructs wrong network magic error indicating the
// correct value.
func NewWrongNetMagicError(correct uint64) WrongNetMagic {
	return WrongNetMagic{
		correctSet: true,
		correct:    correct,
	}
}

// Error implements built-in error interface.
func (x WrongNetMagic) Error() string {
	const desc = "wrong network magic"
	if x.msg != "" {
		if x.correctSet {
			return fmt.Sprintf(errFmt, status.WrongNetMagic, fmt.Sprintf("%s, expected %d", desc, x.correct), x.msg)
		}
		return fmt.Sprintf(errFmt, status.WrongNetMagic, desc, x.msg)
	}
	if x.correctSet {
		return fmt.Sprintf(errFmtNoMessage, status.WrongNetMagic, fmt.Sprintf("%s, expected %d", desc, x.correct))
	}
	return fmt.Sprintf(errFmtNoMessage, status.WrongNetMagic, desc)
}

// Is checks whether target is of type WrongNetMagic, *WrongNetMagic or [Error].
// Is implements interface consumed by [errors.Is].
func (x WrongNetMagic) Is(target error) bool { return errorIs(x, target) }

func (x *WrongNetMagic) readFromV2(m *status.Status) error {
	if m.Code != status.WrongNetMagic {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.WrongNetMagic))
	}
	if len(m.Details) > 0 {
		if len(m.Details) > 1 {
			return fmt.Errorf("too many details (%d)", len(m.Details))
		}
		if m.Details[0].Id != status.DetailCorrectNetMagic {
			return fmt.Errorf("unsupported detail ID=%d", m.Details[0].Id)
		}
		if len(m.Details[0].Value) != 8 {
			return fmt.Errorf("invalid correct value detail: invalid length %d", len(m.Details[0].Value))
		}
		x.correct = binary.BigEndian.Uint64(m.Details[0].Value)
		x.correctSet = true
	} else {
		x.correctSet = false
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x WrongNetMagic) ErrorToV2() *status.Status {
	st := status.Status{Code: status.WrongNetMagic, Message: x.msg}
	if x.correctSet {
		st.Details = []*status.Status_Detail{{
			Id:    status.DetailCorrectNetMagic,
			Value: make([]byte, 8),
		}}
		binary.BigEndian.PutUint64(st.Details[0].Value, x.correct)
	}
	return &st
}

// CorrectMagic returns network magic returned by the server. Returns zero if
// the value was not attached.
func (x WrongNetMagic) CorrectMagic() uint64 {
	if x.correctSet {
		return x.correct
	}
	return 0
}

// SignatureVerificationFailure describes failure status related to signature
// verification failures.
type SignatureVerificationFailure string

// NewSignatureVerificationFailure constructs signature verification error with
// specified cause.
func NewSignatureVerificationFailure(cause error) SignatureVerificationFailure {
	return SignatureVerificationFailure(cause.Error())
}

// Error implements built-in error interface.
func (x SignatureVerificationFailure) Error() string {
	const desc = "signature verification failed"
	if x != "" {
		return fmt.Sprintf(errFmt, status.SignatureVerificationFail, desc, string(x))
	}
	return fmt.Sprintf(errFmtNoMessage, status.SignatureVerificationFail, desc)
}

// Is checks whether target is of type SignatureVerificationFailure,
// *SignatureVerificationFailure or [Error]. Is implements interface consumed by
// [errors.Is].
func (x SignatureVerificationFailure) Is(target error) bool { return errorIs(x, target) }

func (x *SignatureVerificationFailure) readFromV2(m *status.Status) error {
	if m.Code != status.SignatureVerificationFail {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.SignatureVerificationFail))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	*x = SignatureVerificationFailure(m.Message)
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x SignatureVerificationFailure) ErrorToV2() *status.Status {
	return &status.Status{Code: status.SignatureVerificationFail, Message: string(x)}
}

// NodeUnderMaintenance describes failure status for nodes being under
// maintenance.
type NodeUnderMaintenance struct{ msg string }

// Error implements built-in error interface.
func (x NodeUnderMaintenance) Error() string {
	const desc = "node is under maintenance"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.NodeUnderMaintenance, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.NodeUnderMaintenance, desc)
}

// Is checks whether target is of type NodeUnderMaintenance,
// *NodeUnderMaintenance or [Error]. Is implements interface consumed by
// [errors.Is].
func (x NodeUnderMaintenance) Is(target error) bool { return errorIs(x, target) }

func (x *NodeUnderMaintenance) readFromV2(m *status.Status) error {
	if m.Code != status.NodeUnderMaintenance {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.NodeUnderMaintenance))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x NodeUnderMaintenance) ErrorToV2() *status.Status {
	return &status.Status{Code: status.NodeUnderMaintenance, Message: x.msg}
}
