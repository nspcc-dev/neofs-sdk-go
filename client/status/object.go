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
