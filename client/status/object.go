package apistatus

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// ObjectLocked describes status of the failure because of the locked object.
// Instances provide Status and StatusV2 interfaces.
type ObjectLocked struct {
	v2 status.Status
}

func (x ObjectLocked) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(object.StatusLocked, object.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *ObjectLocked) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: LOCKED;
//  * string message: "object is locked";
//  * details: empty.
func (x ObjectLocked) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(object.StatusLocked, object.GlobalizeFail))
	x.v2.SetMessage("object is locked")
	return &x.v2
}

// LockNonRegularObject describes status returned on locking the non-regular object.
// Instances provide Status and StatusV2 interfaces.
type LockNonRegularObject struct {
	v2 status.Status
}

func (x LockNonRegularObject) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(object.StatusLockNonRegularObject, object.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *LockNonRegularObject) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: LOCK_NON_REGULAR_OBJECT;
//  * string message: "locking non-regular object is forbidden";
//  * details: empty.
func (x LockNonRegularObject) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(object.StatusLockNonRegularObject, object.GlobalizeFail))
	x.v2.SetMessage("locking non-regular object is forbidden")
	return &x.v2
}

// ObjectAccessDenied describes status of the failure because of the access control violation.
// Instances provide Status and StatusV2 interfaces.
type ObjectAccessDenied struct {
	v2 status.Status
}

func (x ObjectAccessDenied) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(object.StatusAccessDenied, object.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *ObjectAccessDenied) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: ACCESS_DENIED;
//  * string message: "access to object operation denied";
//  * details: empty.
func (x ObjectAccessDenied) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(object.StatusAccessDenied, object.GlobalizeFail))
	x.v2.SetMessage("access to object operation denied")
	return &x.v2
}

// WriteReason writes human-readable access rejection reason.
func (x *ObjectAccessDenied) WriteReason(reason string) {
	object.WriteAccessDeniedDesc(&x.v2, reason)
}

// Reason returns human-readable access rejection reason returned by the server.
// Returns empty value is reason is not presented.
func (x ObjectAccessDenied) Reason() string {
	return object.ReadAccessDeniedDesc(x.v2)
}

// ObjectNotFound describes status of the failure because of the missing object.
// Instances provide Status and StatusV2 interfaces.
type ObjectNotFound struct {
	v2 status.Status
}

func (x ObjectNotFound) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(object.StatusNotFound, object.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *ObjectNotFound) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: OBJECT_NOT_FOUND;
//  * string message: "object not found";
//  * details: empty.
func (x ObjectNotFound) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(object.StatusNotFound, object.GlobalizeFail))
	x.v2.SetMessage("object not found")
	return &x.v2
}
