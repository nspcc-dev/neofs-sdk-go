package apistatus

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// StatusV2 defines a variety of status instances compatible with NeoFS API V2 protocol.
//
// Note: it is not recommended to use this type directly, it is intended for documentation of the library functionality.
type StatusV2 interface {
	// ToStatusV2 returns the status as github.com/nspcc-dev/neofs-api-go/v2/status.Status message structure.
	ToStatusV2() *status.Status
}

// FromStatusV2 converts [status.Status] message structure to error. Inverse to [ToStatusV2] operation.
//
// If result is not nil, it implements [StatusV2]. This fact should be taken into account only when passing
// the result to the inverse function [ToStatusV2], casts are not compatibility-safe.
//
// Below is the mapping of return codes to status instance types (with a description of parsing details).
// Note: notice if the return type is a pointer.
//
// Successes:
//   - [status.OK]: nil (this also includes nil argument).
//
// Common failures:
//   - [status.Internal]: *[ServerInternal];
//   - [status.SignatureVerificationFail]: *[SignatureVerification].
//   - [status.WrongMagicNumber]: *[WrongMagicNumber].
//   - [status.NodeUnderMaintenance]: *[NodeUnderMaintenance].
//
// Object failures:
//   - [object.StatusLocked]: *[ObjectLocked];
//   - [object.StatusLockNonRegularObject]: *[LockNonRegularObject].
//   - [object.StatusAccessDenied]: *[ObjectAccessDenied].
//   - [object.StatusNotFound]: *[ObjectNotFound].
//   - [object.StatusAlreadyRemoved]: *[ObjectAlreadyRemoved].
//   - [object.StatusOutOfRange]: *[ObjectOutOfRange].
//
// Container failures:
//   - [container.StatusNotFound]: *[ContainerNotFound];
//   - [container.StatusEACLNotFound]: *[EACLNotFound];
//
// Session failures:
//   - [session.StatusTokenNotFound]: *[SessionTokenNotFound];
//   - [session.StatusTokenExpired]: *[SessionTokenExpired];
func FromStatusV2(st *status.Status) error {
	var decoder interface {
		fromStatusV2(*status.Status)
		Error() string
	}

	switch code := st.Code(); {
	case status.IsSuccess(code):
		//nolint:exhaustive
		switch status.LocalizeSuccess(&code); code {
		case status.OK:
			return nil
		}
	case status.IsCommonFail(code):
		switch status.LocalizeCommonFail(&code); code {
		case status.Internal:
			decoder = new(ServerInternal)
		case status.WrongMagicNumber:
			decoder = new(WrongMagicNumber)
		case status.SignatureVerificationFail:
			decoder = new(SignatureVerification)
		case status.NodeUnderMaintenance:
			decoder = new(NodeUnderMaintenance)
		}
	case object.LocalizeFailStatus(&code):
		switch code {
		case object.StatusLocked:
			decoder = new(ObjectLocked)
		case object.StatusLockNonRegularObject:
			decoder = new(LockNonRegularObject)
		case object.StatusAccessDenied:
			decoder = new(ObjectAccessDenied)
		case object.StatusNotFound:
			decoder = new(ObjectNotFound)
		case object.StatusAlreadyRemoved:
			decoder = new(ObjectAlreadyRemoved)
		case object.StatusOutOfRange:
			decoder = new(ObjectOutOfRange)
		}
	case container.LocalizeFailStatus(&code):
		//nolint:exhaustive
		switch code {
		case container.StatusNotFound:
			decoder = new(ContainerNotFound)
		case container.StatusEACLNotFound:
			decoder = new(EACLNotFound)
		}
	case session.LocalizeFailStatus(&code):
		//nolint:exhaustive
		switch code {
		case session.StatusTokenNotFound:
			decoder = new(SessionTokenNotFound)
		case session.StatusTokenExpired:
			decoder = new(SessionTokenExpired)
		}
	}

	if decoder == nil {
		decoder = new(UnrecognizedStatusV2)
	}

	decoder.fromStatusV2(st)

	return decoder
}

// ToStatusV2 converts error to status.Status message structure. Inverse to [FromStatusV2] operation.
//
// If argument is the [StatusV2] instance, it is converted directly.
// Otherwise, successes are converted with [status.OK] code w/o details and message,
// failures - with [status.Internal] and error text message w/o details.
func ToStatusV2(err error) *status.Status {
	if err == nil {
		return newStatusV2WithLocalCode(status.OK, status.GlobalizeSuccess)
	}

	var instance StatusV2
	if errors.As(err, &instance) {
		return instance.ToStatusV2()
	}

	internalErrorStatus := newStatusV2WithLocalCode(status.Internal, status.GlobalizeCommonFail)
	internalErrorStatus.SetMessage(err.Error())

	return internalErrorStatus
}

func errMessageStatusV2(code any, msg string) string {
	const (
		noMsgFmt = "status: code = %v"
		msgFmt   = noMsgFmt + " message = %s"
	)

	if msg != "" {
		return fmt.Sprintf(msgFmt, code, msg)
	}

	return fmt.Sprintf(noMsgFmt, code)
}

func newStatusV2WithLocalCode(code status.Code, globalizer func(*status.Code)) *status.Status {
	var st status.Status

	st.SetCode(globalizeCodeV2(code, globalizer))

	return &st
}

func globalizeCodeV2(code status.Code, globalizer func(*status.Code)) status.Code {
	globalizer(&code)
	return code
}
