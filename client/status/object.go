package apistatus

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

// Object error instances which may be used to check API errors against using
// [errors.Is]. All of them MUST NOT be changed.
var (
	ErrObjectAccessDenied   ObjectAccessDenied
	ErrObjectNotFound       ObjectNotFound
	ErrObjectLocked         ObjectLocked
	ErrLockIrregularObject  LockIrregularObject
	ErrObjectAlreadyRemoved ObjectAlreadyRemoved
	ErrObjectOutOfRange     ObjectOutOfRange
)

// ObjectLocked describes status of the failure because of the locked object.
type ObjectLocked struct{ msg string }

// Error implements built-in error interface.
func (x ObjectLocked) Error() string {
	const desc = "object is locked"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.ObjectLocked, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.ObjectLocked, desc)
}

// Is checks whether target is of type ObjectLocked, *ObjectLocked or [Error].
// Is implements interface consumed by [errors.Is].
func (x ObjectLocked) Is(target error) bool { return errorIs(x, target) }

func (x *ObjectLocked) readFromV2(m *status.Status) error {
	if m.Code != status.ObjectLocked {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.ObjectLocked))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ObjectLocked) ErrorToV2() *status.Status {
	return &status.Status{Code: status.ObjectLocked, Message: x.msg}
}

// LockIrregularObject describes status returned on locking the irregular
// object.
type LockIrregularObject struct{ msg string }

// Error implements built-in error interface.
func (x LockIrregularObject) Error() string {
	const desc = "locking irregular object is forbidden"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.LockIrregularObject, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.LockIrregularObject, desc)
}

// Is checks whether target is of type LockIrregularObject, *LockIrregularObject
// or [Error]. Is implements interface consumed by [errors.Is].
func (x LockIrregularObject) Is(target error) bool { return errorIs(x, target) }

func (x *LockIrregularObject) readFromV2(m *status.Status) error {
	if m.Code != status.LockIrregularObject {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.LockIrregularObject))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x LockIrregularObject) ErrorToV2() *status.Status {
	return &status.Status{Code: status.LockIrregularObject, Message: x.msg}
}

// ObjectAccessDenied describes status of the failure because of the access
// control violation.
type ObjectAccessDenied struct{ reason, msg string }

// NewObjectAccessDeniedError constructs object access denial error indicating
// the reason.
func NewObjectAccessDeniedError(reason string) ObjectAccessDenied {
	return ObjectAccessDenied{reason: reason}
}

// Error implements built-in error interface.
func (x ObjectAccessDenied) Error() string {
	const desc = "object access denied"
	if x.msg != "" {
		if x.reason != "" {
			return fmt.Sprintf(errFmt, status.ObjectAccessDenied, fmt.Sprintf("%s, reason: %s", desc, x.reason), x.msg)
		}
		return fmt.Sprintf(errFmt, status.ObjectAccessDenied, desc, x.msg)
	}
	if x.reason != "" {
		return fmt.Sprintf(errFmtNoMessage, status.ObjectAccessDenied, fmt.Sprintf("%s, reason: %s", desc, x.reason))
	}
	return fmt.Sprintf(errFmtNoMessage, status.ObjectAccessDenied, desc)
}

// Is checks whether target is of type ObjectAccessDenied, *ObjectAccessDenied
// or [Error]. Is implements interface consumed by [errors.Is].
func (x ObjectAccessDenied) Is(target error) bool { return errorIs(x, target) }

func (x *ObjectAccessDenied) readFromV2(m *status.Status) error {
	if m.Code != status.ObjectAccessDenied {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.ObjectAccessDenied))
	}
	if len(m.Details) > 0 {
		if len(m.Details) > 1 {
			return fmt.Errorf("too many details (%d)", len(m.Details))
		}
		if m.Details[0].Id != status.DetailObjectAccessDenialReason {
			return fmt.Errorf("unsupported detail ID=%d", m.Details[0].Id)
		}
		if !utf8.Valid(m.Details[0].Value) {
			return errors.New("invalid reason detail: invalid UTF-8 string")
		}
		x.reason = string(m.Details[0].Value)
	} else {
		x.reason = ""
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ObjectAccessDenied) ErrorToV2() *status.Status {
	st := status.Status{Code: status.ObjectAccessDenied, Message: x.msg}
	if x.reason != "" {
		st.Details = []*status.Status_Detail{{
			Id:    status.DetailObjectAccessDenialReason,
			Value: []byte(x.reason),
		}}
	}
	return &st
}

// Reason returns human-readable access rejection reason returned by the server.
// Returns empty value is reason is not presented.
func (x ObjectAccessDenied) Reason() string {
	return x.reason
}

// ObjectNotFound describes status of the failure because of the missing object.
type ObjectNotFound string

// NewObjectNotFoundError constructs missing object error with specified cause.
func NewObjectNotFoundError(cause error) ObjectNotFound {
	return ObjectNotFound(cause.Error())
}

// Error implements built-in error interface.
func (x ObjectNotFound) Error() string {
	const desc = "object not found"
	if x != "" {
		return fmt.Sprintf(errFmt, status.ObjectNotFound, desc, string(x))
	}
	return fmt.Sprintf(errFmtNoMessage, status.ObjectNotFound, desc)
}

// Is checks whether target is of type ObjectNotFound, *ObjectNotFound or
// [Error]. Is implements interface consumed by [errors.Is].
func (x ObjectNotFound) Is(target error) bool { return errorIs(x, target) }

func (x *ObjectNotFound) readFromV2(m *status.Status) error {
	if m.Code != status.ObjectNotFound {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.ObjectNotFound))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	*x = ObjectNotFound(m.Message)
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ObjectNotFound) ErrorToV2() *status.Status {
	return &status.Status{Code: status.ObjectNotFound, Message: string(x)}
}

// ObjectAlreadyRemoved describes status of the failure because object has been
// already removed.
type ObjectAlreadyRemoved struct{ msg string }

// Error implements built-in error interface.
func (x ObjectAlreadyRemoved) Error() string {
	const desc = "object already removed"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.ObjectAlreadyRemoved, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.ObjectAlreadyRemoved, desc)
}

// Is checks whether target is of type ObjectAlreadyRemoved,
// *ObjectAlreadyRemoved or [Error]. Is implements interface consumed by
// [errors.Is].
func (x ObjectAlreadyRemoved) Is(target error) bool { return errorIs(x, target) }

func (x *ObjectAlreadyRemoved) readFromV2(m *status.Status) error {
	if m.Code != status.ObjectAlreadyRemoved {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.ObjectAlreadyRemoved))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ObjectAlreadyRemoved) ErrorToV2() *status.Status {
	return &status.Status{Code: status.ObjectAlreadyRemoved, Message: x.msg}
}

// ObjectOutOfRange describes status of the failure because of the incorrect
// provided object ranges.
type ObjectOutOfRange struct{ msg string }

// Error implements built-in error interface.
func (x ObjectOutOfRange) Error() string {
	const desc = "out of range"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.OutOfRange, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.OutOfRange, desc)
}

// Is checks whether target is of type ObjectOutOfRange, *ObjectOutOfRange or
// [Error]. Is implements interface consumed by [errors.Is].
func (x ObjectOutOfRange) Is(target error) bool { return errorIs(x, target) }

func (x *ObjectOutOfRange) readFromV2(m *status.Status) error {
	if m.Code != status.OutOfRange {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.OutOfRange))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ObjectOutOfRange) ErrorToV2() *status.Status {
	return &status.Status{Code: status.OutOfRange, Message: x.msg}
}
