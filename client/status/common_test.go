package apistatus_test

import (
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestServerInternal_Message(t *testing.T) {
	const msg = "some message"

	var st apistatus.ServerInternal

	res := st.Message()
	resv2 := apistatus.ToStatusV2(st).Message()
	require.Empty(t, res)
	require.Empty(t, resv2)

	st.SetMessage(msg)

	res = st.Message()
	resv2 = apistatus.ToStatusV2(st).Message()
	require.Equal(t, msg, res)
	require.Equal(t, msg, resv2)
}
