package apistatus

import (
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

var (
	// ErrSessionTokenNotFound is an instance of SessionTokenNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrSessionTokenNotFound SessionTokenNotFound
	// ErrSessionTokenExpired is an instance of SessionTokenExpired error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrSessionTokenExpired SessionTokenExpired
)

// SessionTokenNotFound describes status of the failure because of the missing session token.
// Instances provide [Status], [StatusV2] and error interfaces.
type SessionTokenNotFound struct {
	v2 status.Status
}

const defaultSessionTokenNotFoundMsg = "session token not found"

func (x SessionTokenNotFound) Error() string {
	msg := x.v2.Message()
	if msg == "" {
		msg = defaultSessionTokenNotFoundMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(session.StatusTokenNotFound, session.GlobalizeFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x SessionTokenNotFound) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case SessionTokenNotFound, *SessionTokenNotFound:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *SessionTokenNotFound) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: TOKEN_NOT_FOUND;
//   - string message: "session token not found";
//   - details: empty.
func (x SessionTokenNotFound) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(session.StatusTokenNotFound, session.GlobalizeFail))
	x.v2.SetMessage(defaultSessionTokenNotFoundMsg)
	return &x.v2
}

// SessionTokenExpired describes status of the failure because of the expired session token.
// Instances provide [Status], [StatusV2] and error interfaces.
type SessionTokenExpired struct {
	v2 status.Status
}

const defaultSessionTokenExpiredMsg = "expired session token"

func (x SessionTokenExpired) Error() string {
	msg := x.v2.Message()
	if msg == "" {
		msg = defaultSessionTokenExpiredMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(session.StatusTokenExpired, session.GlobalizeFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x SessionTokenExpired) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case SessionTokenExpired, *SessionTokenExpired:
		return true
	}
}

// implements local interface defined in FromStatusV2 func.
func (x *SessionTokenExpired) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ToStatusV2 implements StatusV2 interface method.
// If the value was returned by FromStatusV2, returns the source message.
// Otherwise, returns message with
//   - code: TOKEN_EXPIRED;
//   - string message: "expired session token";
//   - details: empty.
func (x SessionTokenExpired) ToStatusV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(session.StatusTokenExpired, session.GlobalizeFail))
	x.v2.SetMessage(defaultSessionTokenExpiredMsg)
	return &x.v2
}
