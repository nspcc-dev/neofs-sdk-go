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
//   - [protostatus.RequestExpired]: *[RequestExpired].
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
//   - [protostatus.ContainerLocked]: *[ContainerLocked];
//   - [protostatus.ContainerAwaitTimeout]: *[ContainerAwaitTimeout];
//   - [protostatus.ContainerRevisionMismatch]: *[ContainerRevisionMismatch];
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

	var (
		code    = st.GetCode()
		offset  = code % 1024
		section = code / 1024
	)

	switch protostatus.Section(section) {
	case protostatus.Section_SECTION_SUCCESS:
		switch protostatus.Success(offset) {
		case protostatus.Success_OK:
			return nil
		case protostatus.Success_INCOMPLETE:
			decoder = new(Incomplete)
		}
	case protostatus.Section_SECTION_FAILURE_COMMON:
		switch protostatus.CommonFail(offset) {
		case protostatus.CommonFail_INTERNAL:
			decoder = new(ServerInternal)
		case protostatus.CommonFail_WRONG_MAGIC_NUMBER:
			decoder = new(WrongMagicNumber)
		case protostatus.CommonFail_SIGNATURE_VERIFICATION_FAIL:
			decoder = new(SignatureVerification)
		case protostatus.CommonFail_NODE_UNDER_MAINTENANCE:
			decoder = new(NodeUnderMaintenance)
		case protostatus.CommonFail_BAD_REQUEST:
			decoder = new(BadRequest)
		case protostatus.CommonFail_BUSY:
			decoder = new(Busy)
		case protostatus.CommonFail_REQUEST_EXPIRED:
			decoder = new(RequestExpired)
		}
	case protostatus.Section_SECTION_OBJECT:
		switch protostatus.Object(offset) {
		case protostatus.Object_ACCESS_DENIED:
			decoder = new(ObjectAccessDenied)
		case protostatus.Object_OBJECT_NOT_FOUND:
			decoder = new(ObjectNotFound)
		case protostatus.Object_LOCKED:
			decoder = new(ObjectLocked)
		case protostatus.Object_LOCK_NON_REGULAR_OBJECT:
			decoder = new(LockNonRegularObject)
		case protostatus.Object_OBJECT_ALREADY_REMOVED:
			decoder = new(ObjectAlreadyRemoved)
		case protostatus.Object_OUT_OF_RANGE:
			decoder = new(ObjectOutOfRange)
		case protostatus.Object_QUOTA_EXCEEDED:
			decoder = new(QuotaExceeded)
		}
	case protostatus.Section_SECTION_CONTAINER:
		switch protostatus.Container(offset) {
		case protostatus.Container_CONTAINER_NOT_FOUND:
			decoder = new(ContainerNotFound)
		case protostatus.Container_EACL_NOT_FOUND:
			decoder = new(EACLNotFound)
		case protostatus.Container_CONTAINER_LOCKED:
			decoder = new(ContainerLocked)
		case protostatus.Container_CONTAINER_AWAIT_TIMEOUT:
			decoder = new(ContainerAwaitTimeout)
		case protostatus.Container_CONTAINER_VERSION_MISMATCH:
			decoder = new(ContainerRevisionMismatch)
		}
	case protostatus.Section_SECTION_SESSION:
		switch protostatus.Session(offset) {
		case protostatus.Session_TOKEN_NOT_FOUND:
			decoder = new(SessionTokenNotFound)
		case protostatus.Session_TOKEN_EXPIRED:
			decoder = new(SessionTokenExpired)
		}
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

	if m, ok := errors.AsType[interface {
		protoMessage() *protostatus.Status
		Error() string
	}](err); ok {
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
