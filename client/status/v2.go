package apistatus

import (
	"errors"
	"fmt"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

// ToError converts [status.Status] message structure to error. Inverse to [FromError] operation.
//
// Below is the mapping of return codes to status instance types (with a description of parsing details).
// Note: notice if the return type is a pointer.
//
// Successes:
//   - [status.OK]: nil (this also includes nil argument).
//
// Common failures:
//   - [protostatus.InternalServerError]: *[ServerInternal];
//   - [protostatus.SignatureVerificationFail]: *[SignatureVerification].
//   - [protostatus.WrongNetMagic]: *[WrongMagicNumber].
//   - [protostatus.NodeUnderMaintenance]: *[NodeUnderMaintenance].
//
// Object failures:
//   - [protostatus.ObjectLocked]: *[ObjectLocked];
//   - [protostatus.LockIrregularObject]: *[LockNonRegularObject].
//   - [protostatus.ObjectAccessDenied]: *[ObjectAccessDenied].
//   - [protostatus.ObjectNotFound]: *[ObjectNotFound].
//   - [protostatus.ObjectAlreadyRemoved]: *[ObjectAlreadyRemoved].
//   - [protostatus.OutOfRange]: *[ObjectOutOfRange].
//   - [protostatus.QuotaExceeded]: *[QuotaExceeded].
//
// Container failures:
//   - [protostatus.ContainerNotFound]: *[ContainerNotFound];
//   - [protostatus.EACLNotFound]: *[EACLNotFound];
//
// Session failures:
//   - [protostatus.SessionTokenNotFound]: *[SessionTokenNotFound];
//   - [protostatus.SessionTokenExpired]: *[SessionTokenExpired];
func ToError(st *protostatus.Status) error {
	for i, d := range st.GetDetails() {
		if d == nil {
			return fmt.Errorf("nil detail #%d", i)
		}
	}

	var decoder interface {
		fromProtoMessage(*protostatus.Status)
		Error() string
	}

	switch code := st.GetCode(); code {
	case protostatus.OK:
		return nil
	case protostatus.InternalServerError:
		decoder = new(ServerInternal)
	case protostatus.WrongNetMagic:
		decoder = new(WrongMagicNumber)
	case protostatus.SignatureVerificationFail:
		decoder = new(SignatureVerification)
	case protostatus.NodeUnderMaintenance:
		decoder = new(NodeUnderMaintenance)
	case protostatus.ObjectLocked:
		decoder = new(ObjectLocked)
	case protostatus.LockIrregularObject:
		decoder = new(LockNonRegularObject)
	case protostatus.ObjectAccessDenied:
		decoder = new(ObjectAccessDenied)
	case protostatus.ObjectNotFound:
		decoder = new(ObjectNotFound)
	case protostatus.ObjectAlreadyRemoved:
		decoder = new(ObjectAlreadyRemoved)
	case protostatus.OutOfRange:
		decoder = new(ObjectOutOfRange)
	case protostatus.QuotaExceeded:
		decoder = new(QuotaExceeded)
	case protostatus.ContainerNotFound:
		decoder = new(ContainerNotFound)
	case protostatus.EACLNotFound:
		decoder = new(EACLNotFound)
	case protostatus.SessionTokenNotFound:
		decoder = new(SessionTokenNotFound)
	case protostatus.SessionTokenExpired:
		decoder = new(SessionTokenExpired)
	}

	if decoder == nil {
		decoder = new(UnrecognizedStatus)
	}

	decoder.fromProtoMessage(st)

	return decoder
}

// FromError converts error to status.Status message structure. Inverse to [ToError] operation.
//
// Nil corresponds to [protostatus.OK] code, any unknown error to
// [protostatus.InternalServerError].
func FromError(err error) *protostatus.Status {
	if err == nil {
		return nil
	}

	var m interface{ protoMessage() *protostatus.Status }
	if errors.As(err, &m) {
		return m.protoMessage()
	}

	return &protostatus.Status{
		Code:    protostatus.InternalServerError,
		Message: err.Error(),
	}
}

func errMessageStatus(code any, msg string) string {
	const (
		noMsgFmt = "status: code = %v"
		msgFmt   = noMsgFmt + " message = %s"
	)

	if msg != "" {
		return fmt.Sprintf(msgFmt, code, msg)
	}

	return fmt.Sprintf(noMsgFmt, code)
}
