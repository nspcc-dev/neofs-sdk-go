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
	detailNum := apistatus.ErrorToV2(st).NumberOfDetails()
	require.Zero(t, detailNum)

	st.WriteReason(reason)

	res = st.Reason()
	require.Equal(t, reason, res)
	detailNum = apistatus.ErrorToV2(st).NumberOfDetails()
	require.EqualValues(t, 1, detailNum)
}
