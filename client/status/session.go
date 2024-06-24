package apistatus

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

// Session error instances which may be used to check API errors against using
// [errors.Is]. All of them MUST NOT be changed.
var (
	ErrSessionTokenNotFound SessionTokenNotFound
	ErrSessionTokenExpired  SessionTokenExpired
)

// SessionTokenNotFound describes status of the failure because of the missing session token.
type SessionTokenNotFound struct{ msg string }

// Error implements built-in error interface.
func (x SessionTokenNotFound) Error() string {
	const desc = "session token not found"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.SessionTokenNotFound, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.SessionTokenNotFound, desc)
}

// Is checks whether target is of type SessionTokenNotFound,
// *SessionTokenNotFound or [Error]. Is implements interface consumed by
// [errors.Is].
func (x SessionTokenNotFound) Is(target error) bool { return errorIs(x, target) }

func (x *SessionTokenNotFound) readFromV2(m *status.Status) error {
	if m.Code != status.SessionTokenNotFound {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.SessionTokenNotFound))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x SessionTokenNotFound) ErrorToV2() *status.Status {
	return &status.Status{Code: status.SessionTokenNotFound, Message: x.msg}
}

// SessionTokenExpired describes status of the failure because of the expired
// session token.
type SessionTokenExpired struct{ msg string }

// Error implements built-in error interface.
func (x SessionTokenExpired) Error() string {
	const desc = "session token has expired"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, status.SessionTokenExpired, desc, x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, status.SessionTokenExpired, desc)
}

// Is checks whether target is of type SessionTokenExpired, *SessionTokenExpired
// or [Error]. Is implements interface consumed by [errors.Is].
func (x SessionTokenExpired) Is(target error) bool { return errorIs(x, target) }

func (x *SessionTokenExpired) readFromV2(m *status.Status) error {
	if m.Code != status.SessionTokenExpired {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.SessionTokenExpired))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	x.msg = m.Message
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x SessionTokenExpired) ErrorToV2() *status.Status {
	return &status.Status{Code: status.SessionTokenExpired, Message: x.msg}
}
