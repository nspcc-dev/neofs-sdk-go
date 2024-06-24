package reputationtest_test

import (
	"testing"

	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/stretchr/testify/require"
)

func TestPeerID(t *testing.T) {
	id := reputationtest.PeerID()
	require.NotEqual(t, id, reputationtest.PeerID())

	var m apireputation.PeerID
	id.WriteToV2(&m)
	var id2 reputation.PeerID
	require.NoError(t, id2.ReadFromV2(&m))
	require.Equal(t, id, id2)
}

func TestTrust(t *testing.T) {
	tr := reputationtest.Trust()
	require.NotEqual(t, tr, reputationtest.Trust())

	var m apireputation.Trust
	tr.WriteToV2(&m)
	var tr2 reputation.Trust
	require.NoError(t, tr2.ReadFromV2(&m))
	require.Equal(t, tr, tr2)
}

func TestPeerToPeerTrust(t *testing.T) {
	tr := reputationtest.PeerToPeerTrust()
	require.NotEqual(t, tr, reputationtest.PeerToPeerTrust())

	var m apireputation.PeerToPeerTrust
	tr.WriteToV2(&m)
	var tr2 reputation.PeerToPeerTrust
	require.NoError(t, tr2.ReadFromV2(&m))
	require.Equal(t, tr, tr2)
}

func TestGlobalTrust(t *testing.T) {
	v := reputationtest.GlobalTrust()
	require.NotEqual(t, v, reputationtest.GlobalTrust())
	require.True(t, v.VerifySignature())

	var v2 reputation.GlobalTrust
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)
}

func TestGlobalTrustUnsigned(t *testing.T) {
	v := reputationtest.GlobalTrustUnsigned()
	require.NotEqual(t, v, reputationtest.GlobalTrustUnsigned())
	require.False(t, v.VerifySignature())

	var v2 reputation.GlobalTrust
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)
}
