package apistatus

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// StatusV2 defines a variety of Status instances compatible with NeoFS API V2 protocol.
//
// Note: it is not recommended to use this type directly, it is intended for documentation of the library functionality.
type StatusV2 interface {
	Status

	// ToStatusV2 returns the status as github.com/nspcc-dev/neofs-api-go/v2/status.Status message structure.
	ToStatusV2() *status.Status
}

// FromStatusV2 converts status.Status message structure to Status instance. Inverse to ToStatusV2 operation.
//
// If result is not nil, it implements StatusV2. This fact should be taken into account only when passing
// the result to the inverse function ToStatusV2, casts are not compatibility-safe.
//
// Below is the mapping of return codes to Status instance types (with a description of parsing details).
// Note: notice if the return type is a pointer.
//
// Successes:
//   * status.OK: *SuccessDefaultV2 (this also includes nil argument).
//
// Common failures:
//   * status.Internal: *ServerInternal.
//
// Object failures:
//   * object.StatusLocked: *ObjectLocked;
//   * object.StatusLockNonRegularObject: *LockNonRegularObject.
func FromStatusV2(st *status.Status) Status {
	var decoder interface {
		fromStatusV2(*status.Status)
	}

	switch code := st.Code(); {
	case status.IsSuccess(code):
		//nolint:exhaustive
		switch status.LocalizeSuccess(&code); code {
		case status.OK:
			decoder = new(SuccessDefaultV2)
		}
	case status.IsCommonFail(code):
		switch status.LocalizeCommonFail(&code); code {
		case status.Internal:
			decoder = new(ServerInternal)
		case status.WrongMagicNumber:
			decoder = new(WrongMagicNumber)
		}
	case object.LocalizeFailStatus(&code):
		switch code {
		case object.StatusLocked:
			decoder = new(ObjectLocked)
		case object.StatusLockNonRegularObject:
			decoder = new(LockNonRegularObject)
		}
	}

	if decoder == nil {
		decoder = new(unrecognizedStatusV2)
	}

	decoder.fromStatusV2(st)

	return decoder
}

// ToStatusV2 converts Status instance to status.Status message structure. Inverse to FromStatusV2 operation.
//
// If argument is the StatusV2 instance, it is converted directly.
// Otherwise, successes are converted with status.OK code w/o details and message, failures - with status.Internal.
func ToStatusV2(st Status) *status.Status {
	if v, ok := st.(StatusV2); ok {
		return v.ToStatusV2()
	}

	if IsSuccessful(st) {
		return newStatusV2WithLocalCode(status.OK, status.GlobalizeSuccess)
	}

	return newStatusV2WithLocalCode(status.Internal, status.GlobalizeCommonFail)
}

func errMessageStatusV2(code interface{}, msg string) string {
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
