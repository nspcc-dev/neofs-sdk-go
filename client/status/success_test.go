package apistatus_test

import (
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestIncomplete_Message(t *testing.T) {
	const msg = "some message"

	var st apistatus.Incomplete

	require.Empty(t, st.Message())
	require.Empty(t, apistatus.FromError(st).Message)

	st.SetMessage(msg)

	require.Equal(t, msg, st.Message())
	require.Equal(t, msg, apistatus.FromError(st).Message)
}
