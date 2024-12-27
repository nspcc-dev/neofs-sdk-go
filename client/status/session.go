package apistatus

import (
	"errors"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
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
type SessionTokenNotFound struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultSessionTokenNotFoundMsg = "session token not found"

func (x SessionTokenNotFound) Error() string {
	if x.msg == "" {
		x.msg = defaultSessionTokenNotFoundMsg
	}

	return errMessageStatus(protostatus.SessionTokenNotFound, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *SessionTokenNotFound) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x SessionTokenNotFound) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultSessionTokenNotFoundMsg
	}
	return &protostatus.Status{Code: protostatus.SessionTokenNotFound, Message: x.msg, Details: x.dts}
}

// SessionTokenExpired describes status of the failure because of the expired session token.
type SessionTokenExpired struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultSessionTokenExpiredMsg = "expired session token"

func (x SessionTokenExpired) Error() string {
	if x.msg == "" {
		x.msg = defaultSessionTokenExpiredMsg
	}

	return errMessageStatus(protostatus.SessionTokenExpired, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *SessionTokenExpired) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x SessionTokenExpired) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultSessionTokenExpiredMsg
	}
	return &protostatus.Status{Code: protostatus.SessionTokenExpired, Message: x.msg, Details: x.dts}
}
