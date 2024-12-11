package reputation_test

import (
	"testing"

	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func TestTrust_Peer(t *testing.T) {
	var trust reputation.Trust

	require.Zero(t, trust.Peer())

	trust = reputationtest.Trust()

	peer := reputationtest.PeerID()

	trust.SetPeer(peer)

	mt := trust.ProtoMessage()

	require.Equal(t, peer.ProtoMessage(), mt.GetPeer())

	var val2 reputation.Trust
	require.NoError(t, val2.FromProtoMessage(mt))

	require.Equal(t, peer, val2.Peer())
}

func TestTrust_Value(t *testing.T) {
	var val reputation.Trust

	require.Zero(t, val.Value())

	val = reputationtest.Trust()

	const value = 0.75

	val.SetValue(value)

	m := val.ProtoMessage()

	require.EqualValues(t, value, m.GetValue())

	var val2 reputation.Trust
	require.NoError(t, val2.FromProtoMessage(m))

	require.EqualValues(t, value, val2.Value())
}

func TestPeerToPeerTrust_TrustingPeer(t *testing.T) {
	var val reputation.PeerToPeerTrust

	require.Zero(t, val.TrustingPeer())

	val = reputationtest.PeerToPeerTrust()

	peer := reputationtest.PeerID()

	val.SetTrustingPeer(peer)

	m := val.ProtoMessage()

	require.Equal(t, peer.ProtoMessage(), m.GetTrustingPeer())

	var val2 reputation.PeerToPeerTrust
	require.NoError(t, val2.FromProtoMessage(m))

	require.Equal(t, peer, val2.TrustingPeer())
}

func TestPeerToPeerTrust_Trust(t *testing.T) {
	var val reputation.PeerToPeerTrust

	require.Zero(t, val.Trust())

	val = reputationtest.PeerToPeerTrust()

	trust := reputationtest.Trust()

	val.SetTrust(trust)

	m := val.ProtoMessage()

	require.Equal(t, trust.ProtoMessage(), m.GetTrust())

	var val2 reputation.PeerToPeerTrust
	require.NoError(t, val2.FromProtoMessage(m))

	require.Equal(t, trust, val2.Trust())
}

func TestGlobalTrust_Init(t *testing.T) {
	var val reputation.GlobalTrust
	val.Init()

	m := val.ProtoMessage()

	require.Equal(t, version.Current().ProtoMessage(), m.GetVersion())
}

func TestGlobalTrust_Manager(t *testing.T) {
	var val reputation.GlobalTrust

	require.Zero(t, val.Manager())

	val = reputationtest.SignedGlobalTrust()

	peer := reputationtest.PeerID()

	val.SetManager(peer)

	m := val.ProtoMessage()

	require.Equal(t, peer.ProtoMessage(), m.GetBody().GetManager())

	var val2 reputation.GlobalTrust
	require.NoError(t, val2.FromProtoMessage(m))

	require.Equal(t, peer, val2.Manager())
}

func TestGlobalTrust_Trust(t *testing.T) {
	var val reputation.GlobalTrust

	require.Zero(t, val.Trust())

	val = reputationtest.SignedGlobalTrust()

	trust := reputationtest.Trust()

	val.SetTrust(trust)

	m := val.ProtoMessage()

	require.Equal(t, trust.ProtoMessage(), m.GetBody().GetTrust())

	var val2 reputation.GlobalTrust
	require.NoError(t, val2.FromProtoMessage(m))

	require.Equal(t, trust, val2.Trust())
}

func TestGlobalTrust_Sign(t *testing.T) {
	val := reputationtest.GlobalTrust()

	require.False(t, val.VerifySignature())

	require.NoError(t, val.Sign(neofscryptotest.Signer()))

	m := val.ProtoMessage()

	require.NotZero(t, m.GetSignature())

	var val2 reputation.GlobalTrust
	require.NoError(t, val2.FromProtoMessage(m))

	require.True(t, val2.VerifySignature())
}

func TestGlobalTrust_SignedData(t *testing.T) {
	val := reputationtest.GlobalTrust()

	require.False(t, val.VerifySignature())

	neofscryptotest.TestSignedData(t, neofscryptotest.Signer(), &val)
}

func TestGlobalTrustEncoding(t *testing.T) {
	val := reputationtest.SignedGlobalTrust()

	t.Run("binary", func(t *testing.T) {
		data := val.Marshal()

		var val2 reputation.GlobalTrust
		require.NoError(t, val2.Unmarshal(data))

		require.Equal(t, val, val2)
	})
}
