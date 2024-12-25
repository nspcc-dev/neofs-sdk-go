package apistatus

import (
	"errors"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

var (
	// ErrObjectLocked is an instance of ObjectLocked error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrObjectLocked ObjectLocked
	// ErrObjectAlreadyRemoved is an instance of ObjectAlreadyRemoved error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrObjectAlreadyRemoved ObjectAlreadyRemoved
	// ErrLockNonRegularObject is an instance of LockNonRegularObject error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrLockNonRegularObject LockNonRegularObject
	// ErrObjectAccessDenied is an instance of ObjectAccessDenied error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrObjectAccessDenied ObjectAccessDenied
	// ErrObjectNotFound is an instance of ObjectNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrObjectNotFound ObjectNotFound
	// ErrObjectOutOfRange is an instance of ObjectOutOfRange error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrObjectOutOfRange ObjectOutOfRange
)

// ObjectLocked describes status of the failure because of the locked object.
type ObjectLocked struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultObjectLockedMsg = "object is locked"

func (x ObjectLocked) Error() string {
	if x.msg == "" {
		x.msg = defaultObjectLockedMsg
	}

	return errMessageStatus(protostatus.ObjectLocked, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ObjectLocked) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ObjectLocked, *ObjectLocked:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ObjectLocked) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ObjectLocked) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultObjectLockedMsg
	}
	return &protostatus.Status{Code: protostatus.ObjectLocked, Message: x.msg, Details: x.dts}
}

// LockNonRegularObject describes status returned on locking the non-regular object.
type LockNonRegularObject struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultLockNonRegularObjectMsg = "locking non-regular object is forbidden"

func (x LockNonRegularObject) Error() string {
	if x.msg == "" {
		x.msg = defaultLockNonRegularObjectMsg
	}

	return errMessageStatus(protostatus.LockIrregularObject, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x LockNonRegularObject) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case LockNonRegularObject, *LockNonRegularObject:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *LockNonRegularObject) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x LockNonRegularObject) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultLockNonRegularObjectMsg
	}
	return &protostatus.Status{Code: protostatus.LockIrregularObject, Message: x.msg, Details: x.dts}
}

// ObjectAccessDenied describes status of the failure because of the access control violation.
type ObjectAccessDenied struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultObjectAccessDeniedMsg = "access to object operation denied"

func (x ObjectAccessDenied) Error() string {
	if x.msg == "" {
		x.msg = defaultObjectAccessDeniedMsg
	}

	return errMessageStatus(protostatus.ObjectAccessDenied, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ObjectAccessDenied) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ObjectAccessDenied, *ObjectAccessDenied:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ObjectAccessDenied) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ObjectAccessDenied) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultObjectAccessDeniedMsg
	}
	return &protostatus.Status{Code: protostatus.ObjectAccessDenied, Message: x.msg, Details: x.dts}
}

// WriteReason writes human-readable access rejection reason.
func (x *ObjectAccessDenied) WriteReason(reason string) {
	val := []byte(reason)
	for i := range x.dts {
		if x.dts[i].Id == protostatus.DetailObjectAccessDenialReason {
			x.dts[i].Value = val
			return
		}
	}
	x.dts = append(x.dts, &protostatus.Status_Detail{
		Id:    protostatus.DetailObjectAccessDenialReason,
		Value: val,
	})
}

// Reason returns human-readable access rejection reason returned by the server.
// Returns empty value is reason is not presented.
func (x ObjectAccessDenied) Reason() string {
	for i := range x.dts {
		if x.dts[i].Id == protostatus.DetailObjectAccessDenialReason {
			return string(x.dts[i].Value)
		}
	}
	return ""
}

// ObjectNotFound describes status of the failure because of the missing object.
type ObjectNotFound struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultObjectNotFoundMsg = "object not found"

func (x ObjectNotFound) Error() string {
	if x.msg == "" {
		x.msg = defaultObjectNotFoundMsg
	}

	return errMessageStatus(protostatus.ObjectNotFound, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ObjectNotFound) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ObjectNotFound, *ObjectNotFound:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ObjectNotFound) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ObjectNotFound) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultObjectNotFoundMsg
	}
	return &protostatus.Status{Code: protostatus.ObjectNotFound, Message: x.msg, Details: x.dts}
}

// ObjectAlreadyRemoved describes status of the failure because object has been
// already removed.
type ObjectAlreadyRemoved struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultObjectAlreadyRemovedMsg = "object already removed"

func (x ObjectAlreadyRemoved) Error() string {
	if x.msg == "" {
		x.msg = defaultObjectAlreadyRemovedMsg
	}

	return errMessageStatus(protostatus.ObjectAlreadyRemoved, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ObjectAlreadyRemoved) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ObjectAlreadyRemoved, *ObjectAlreadyRemoved:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ObjectAlreadyRemoved) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ObjectAlreadyRemoved) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultObjectAlreadyRemovedMsg
	}
	return &protostatus.Status{Code: protostatus.ObjectAlreadyRemoved, Message: x.msg, Details: x.dts}
}

// ObjectOutOfRange describes status of the failure because of the incorrect
// provided object ranges.
type ObjectOutOfRange struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultObjectOutOfRangeMsg = "out of range"

func (x ObjectOutOfRange) Error() string {
	if x.msg == "" {
		x.msg = defaultObjectOutOfRangeMsg
	}

	return errMessageStatus(protostatus.OutOfRange, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ObjectOutOfRange) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ObjectOutOfRange, *ObjectOutOfRange:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ObjectOutOfRange) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ObjectOutOfRange) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultObjectOutOfRangeMsg
	}
	return &protostatus.Status{Code: protostatus.OutOfRange, Message: x.msg, Details: x.dts}
}
