package apistatus_test

import (
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestObjectAccessDenied_WriteReason(t *testing.T) {
	const reason = "any reason"

	var st apistatus.ObjectAccessDenied

	res := st.Reason()
	require.Empty(t, res)
	require.Empty(t, apistatus.FromError(st).Details)

	st.WriteReason(reason)

	res = st.Reason()
	require.Equal(t, reason, res)
	require.Len(t, apistatus.FromError(st).Details, 1)
}
