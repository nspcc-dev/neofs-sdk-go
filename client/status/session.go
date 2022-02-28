package apistatus

import (
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

// SessionTokenNotFound describes status of the failure because of the missing session token.
// Instances provide Status and StatusV2 interfaces.
type SessionTokenNotFound struct {
	v2 status.Status
}

func (x SessionTokenNotFound) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(session.StatusTokenNotFound, session.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *SessionTokenNotFound) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: TOKEN_NOT_FOUND;
//  * string message: "session token not found";
//  * details: empty.
func (x SessionTokenNotFound) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(session.StatusTokenNotFound, session.GlobalizeFail))
	x.v2.SetMessage("session token not found")
	return &x.v2
}

// SessionTokenExpired describes status of the failure because of the expired session token.
// Instances provide Status and StatusV2 interfaces.
type SessionTokenExpired struct {
	v2 status.Status
}

func (x SessionTokenExpired) Error() string {
	return errMessageStatusV2(
		globalizeCodeV2(session.StatusTokenExpired, session.GlobalizeFail),
		x.v2.Message(),
	)
}

// implements local interface defined in FromStatusV2 func.
func (x *SessionTokenExpired) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//  * code: TOKEN_EXPIRED;
//  * string message: "expired session token";
//  * details: empty.
func (x SessionTokenExpired) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(session.StatusTokenExpired, session.GlobalizeFail))
	x.v2.SetMessage("expired session token")
	return &x.v2
}
