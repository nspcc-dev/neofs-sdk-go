package status_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStatus(t *testing.T) {
	v := &status.Status{
		Code:    1,
		Message: "any_message",
		Details: []*status.Status_Detail{
			{Id: 2, Value: []byte("any_detail1")},
			{Id: 3, Value: []byte("any_detail2")},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res status.Status
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Code, res.Code)
	require.Equal(t, v.Message, res.Message)
	require.Equal(t, v.Details, res.Details)
}
