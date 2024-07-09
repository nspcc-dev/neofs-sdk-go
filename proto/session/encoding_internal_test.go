package session

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type unsupportedSessionContext struct{}

func (unsupportedSessionContext) isSessionToken_Body_Context() {} //nolint:revive

func TestSessionToken_Body(t *testing.T) {
	var v SessionToken_Body
	v.Context = unsupportedSessionContext{}
	require.PanicsWithValue(t, "unexpected context session.unsupportedSessionContext", func() {
		v.MarshaledSize()
	})
	require.PanicsWithValue(t, "unexpected context session.unsupportedSessionContext", func() {
		v.MarshalStable(make([]byte, 100))
	})
}
