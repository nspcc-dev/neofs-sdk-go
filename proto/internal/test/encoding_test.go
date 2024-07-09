package prototest

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

func TestEqualProtoMessages(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.True(t, equalProtoMessages(nil, nil))
		var x, y *object.GetRequest
		require.True(t, equalProtoMessages(x, y))
		x = new(object.GetRequest)
		require.True(t, equalProtoMessages(x, y))
		x.Body = new(object.GetRequest_Body)
		require.True(t, equalProtoMessages(x, y))
		x.Body.Address = new(refs.Address)
		require.True(t, equalProtoMessages(x, y))
		x.Body.Address.ContainerId = new(refs.ContainerID)
		require.True(t, equalProtoMessages(x, y))
		x.Body.Address.ContainerId.Value = []byte{}
		require.True(t, equalProtoMessages(x, y))
		x.Body.Address.ContainerId.Value = []byte{1}
		require.False(t, equalProtoMessages(x, y))
	})
}
