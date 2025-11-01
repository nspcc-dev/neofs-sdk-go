package client

import (
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// sessionContainer is a special type which unifies session logic management for client parameters.
// All methods make public, because sessionContainer is included in Prm* structs.
type sessionContainer struct {
	isSessionIgnored bool
	session          *session.Object
	sessionV2        *session.TokenV2
}

// GetSession returns session object.
//
// Returns:
//   - [ErrNoSession] err if session wasn't set.
//   - [ErrNoSessionExplicitly] if IgnoreSession was used.
func (x *sessionContainer) GetSession() (*session.Object, error) {
	if x.isSessionIgnored {
		return nil, ErrNoSessionExplicitly
	}

	if x.session == nil {
		return nil, ErrNoSession
	}
	return x.session, nil
}

// GetSessionV2 returns session token V2.
//
// Returns:
//   - [ErrNoSession] err if session wasn't set.
//   - [ErrNoSessionExplicitly] if IgnoreSession was used.
func (x *sessionContainer) GetSessionV2() (*session.TokenV2, error) {
	if x.isSessionIgnored {
		return nil, ErrNoSessionExplicitly
	}

	if x.sessionV2 == nil {
		return nil, ErrNoSession
	}
	return x.sessionV2, nil
}

// WithinSession specifies session within which the query must be executed.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// See also IgnoreSession.
//
// Must be signed.
func (x *sessionContainer) WithinSession(t session.Object) {
	x.session = &t
	x.sessionV2 = nil
	x.isSessionIgnored = false
}

// WithinSessionV2 specifies session token V2 within which the query must be executed.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// V2 tokens support multiple subjects, delegation chains, and unified contexts.
// When both V1 and V2 tokens are present, V2 takes precedence.
//
// See also WithinSession, IgnoreSession.
//
// Must be signed.
func (x *sessionContainer) WithinSessionV2(t session.TokenV2) {
	x.sessionV2 = &t
	x.session = nil
	x.isSessionIgnored = false
}

// IgnoreSession disables auto-session creation.
//
// See also WithinSession.
func (x *sessionContainer) IgnoreSession() {
	x.isSessionIgnored = true
	x.session = nil
	x.sessionV2 = nil
}
