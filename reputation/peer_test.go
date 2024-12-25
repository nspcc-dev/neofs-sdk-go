package reputation_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/stretchr/testify/require"
)

func TestPeerID_PublicKey(t *testing.T) {
	var val reputation.PeerID

	require.Zero(t, val.PublicKey())

	key := []byte{3, 2, 1}

	val.SetPublicKey(key)

	m := val.ProtoMessage()

	require.Equal(t, key, m.GetPublicKey())

	var val2 reputation.PeerID
	require.NoError(t, val2.FromProtoMessage(m))

	require.Equal(t, key, val.PublicKey())

	require.True(t, reputation.ComparePeerKey(val, key))
}

func TestPeerID_EncodeToString(t *testing.T) {
	val := reputationtest.PeerID()
	var val2 reputation.PeerID

	require.NoError(t, val2.DecodeString(val.EncodeToString()))
	require.Equal(t, val, val2)
}
