package reputation_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGlobalTrust_Body(t *testing.T) {
	v := &reputation.GlobalTrust_Body{
		Manager: &reputation.PeerID{PublicKey: []byte("any_manager_key")},
		Trust: &reputation.Trust{
			Peer:  &reputation.PeerID{PublicKey: []byte("any_peer_key")},
			Value: 0.5,
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res reputation.GlobalTrust_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Manager, res.Manager)
	require.Equal(t, v.Trust, res.Trust)
}

func TestAnnounceLocalTrustRequest_Body(t *testing.T) {
	v := &reputation.AnnounceLocalTrustRequest_Body{
		Epoch: 1,
		Trusts: []*reputation.Trust{
			{Peer: &reputation.PeerID{PublicKey: []byte("any_public_key1")}, Value: 2.3},
			{Peer: &reputation.PeerID{PublicKey: []byte("any_public_key2")}, Value: 3.4},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res reputation.AnnounceLocalTrustRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Epoch, res.Epoch)
	require.Equal(t, v.Trusts, res.Trusts)
}

func TestAnnounceLocalTrustResponse_Body(t *testing.T) {
	var v reputation.AnnounceLocalTrustResponse_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestAnnounceIntermediateResultRequest_Body(t *testing.T) {
	v := &reputation.AnnounceIntermediateResultRequest_Body{
		Epoch:     1,
		Iteration: 2,
		Trust: &reputation.PeerToPeerTrust{
			TrustingPeer: &reputation.PeerID{PublicKey: []byte("any_public_key1")},
			Trust: &reputation.Trust{
				Peer:  &reputation.PeerID{PublicKey: []byte("any_public_key2")},
				Value: 3.4,
			},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res reputation.AnnounceIntermediateResultRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Epoch, res.Epoch)
	require.Equal(t, v.Iteration, res.Iteration)
	require.Equal(t, v.Trust, res.Trust)
}

func TestAnnounceIntermediateResultResponse_Body(t *testing.T) {
	var v reputation.AnnounceIntermediateResultResponse_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}
