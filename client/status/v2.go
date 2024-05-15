package apistatus

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

const (
	errFmtNoMessage = "status: code = %v (%s)"
	errFmt          = errFmtNoMessage + " message = %s"
)

// errorIs checks whether target is of type T, *T or [Error].
func errorIs[T error, PTR *T](_ T, target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case T, PTR:
		return true
	}
}

// StatusV2 defines a variety of status instances compatible with NeoFS API V2
// protocol.
//
// Note: it is not recommended to use this type directly, it is intended for documentation of the library functionality.
type StatusV2 interface {
	error
	// ErrorToV2 returns the status as [status.Status] message structure.
	ErrorToV2() *status.Status
}

// ErrorFromV2 converts [status.Status] message structure to error. Inverse to [ErrorToV2] operation.
//
// If result is not nil, it implements [StatusV2]. This fact should be taken into account only when passing
// the result to the inverse function [ErrorToV2], casts are not compatibility-safe.
//
// Below is the mapping of return codes to status instance types (with a description of parsing details).
// Note: notice if the return type is a pointer.
//
// Successes:
//   - [status.OK]: nil (this also includes nil argument).
//
// Common failures:
//   - [status.InternalServerError]: [InternalServerError];
//   - [status.SignatureVerificationFail]: [SignatureVerification].
//   - [status.WrongMagicNumber]: [WrongMagicNumber].
//   - [status.NodeUnderMaintenance]: [NodeUnderMaintenance].
//
// Object failures:
//   - [status.ObjectLocked]: [ObjectLocked];
//   - [status.LockIrregularObject]: [LockNonRegularObject].
//   - [status.ObjectAccessDenied]: [ObjectAccessDenied].
//   - [status.ObjectNotFound]: [ObjectNotFound].
//   - [status.ObjectAlreadyRemoved]: [ObjectAlreadyRemoved].
//   - [status.OutOfRange]: [ObjectOutOfRange].
//
// Container failures:
//   - [status.ContainerNotFound]: [ContainerNotFound];
//   - [status.EACLNotFound]: [EACLNotFound];
//
// Session failures:
//   - [status.SessionTokenNotFound]: [SessionTokenNotFound];
//   - [status.SessionTokenExpired]: [SessionTokenExpired];
func ErrorFromV2(st *status.Status) (StatusV2, error) {
	switch st.GetCode() {
	default:
		return unrecognizedStatusV2{st.Code, st.Message, st.Details}, nil
	case status.OK:
		return nil, nil
	case status.InternalServerError:
		var e InternalServerError
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid internal server error status: %w", err)
		}
		return e, nil
	case status.WrongNetMagic:
		var e WrongNetMagic
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid wrong network magic status: %w", err)
		}
		return e, nil
	case status.SignatureVerificationFail:
		var e SignatureVerificationFailure
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid signature verification failure status: %w", err)
		}
		return e, nil
	case status.NodeUnderMaintenance:
		var e NodeUnderMaintenance
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid node maintenance status: %w", err)
		}
		return e, nil
	case status.ObjectAccessDenied:
		var e ObjectAccessDenied
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid object access denial status: %w", err)
		}
		return e, nil
	case status.ObjectNotFound:
		var e ObjectNotFound
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid missing object status: %w", err)
		}
		return e, nil
	case status.ObjectLocked:
		var e ObjectLocked
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid locked object status: %w", err)
		}
		return e, nil
	case status.LockIrregularObject:
		var e LockIrregularObject
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid locking irregular object status: %w", err)
		}
		return e, nil
	case status.ObjectAlreadyRemoved:
		var e ObjectAlreadyRemoved
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid already removed object status: %w", err)
		}
		return e, nil
	case status.OutOfRange:
		var e ObjectOutOfRange
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid out-of-range status: %w", err)
		}
		return e, nil
	case status.ContainerNotFound:
		var e ContainerNotFound
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid missing container status: %w", err)
		}
		return e, nil
	case status.EACLNotFound:
		var e EACLNotFound
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid missing eACL status: %w", err)
		}
		return e, nil
	case status.SessionTokenNotFound:
		var e SessionTokenNotFound
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid missing session token status: %w", err)
		}
		return e, nil
	case status.SessionTokenExpired:
		var e SessionTokenExpired
		if err := e.readFromV2(st); err != nil {
			return nil, fmt.Errorf("invalid expired session token status: %w", err)
		}
		return e, nil
	}
}

// ErrorToV2 converts error to status.Status message structure. Inverse to [ErrorFromV2] operation.
//
// If argument is the [StatusV2] instance, it is converted directly. Otherwise,
// successes are returned as nil, failures - with [status.Internal] and error
// text message w/o details.
func ErrorToV2(err error) *status.Status {
	if err == nil {
		return nil
	}

	var instance StatusV2
	if errors.As(err, &instance) {
		return instance.ErrorToV2()
	}

	return &status.Status{Code: status.InternalServerError, Message: err.Error()}
}
