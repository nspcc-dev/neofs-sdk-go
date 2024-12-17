package reputation_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
	"github.com/stretchr/testify/require"
)

// returns random reputation.PeerID with all non-zero fields.
func randPeerID() *reputation.PeerID {
	return &reputation.PeerID{PublicKey: prototest.RandBytes()}
}

// returns random reputation.Trust with all non-zero fields.
func randTrust() *reputation.Trust {
	return &reputation.Trust{
		Peer:  randPeerID(),
		Value: prototest.RandFloat64(),
	}
}

// returns non-empty list of reputation.Trust up to 10 elements. Each element
// may be nil and pointer to zero.
func randTrusts() []*reputation.Trust { return prototest.RandRepeated(randTrust) }

// returns random reputation.PeerToPeerTrust with all non-zero fields.
func randPeerToPeerTrust() *reputation.PeerToPeerTrust {
	return &reputation.PeerToPeerTrust{
		TrustingPeer: randPeerID(),
		Trust:        randTrust(),
	}
}

func TestPeerID_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.PeerID{
		randPeerID(),
	})
}

func TestTrust_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.Trust{
		randTrust(),
	})
}

func TestPeerToPeerTrust_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.PeerToPeerTrust{
		randPeerToPeerTrust(),
	})
}

func TestGlobalTrust_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.GlobalTrust_Body{
		{
			Manager: randPeerID(),
			Trust:   randTrust(),
		},
	})
}

func TestAnnounceLocalTrustRequest_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &reputation.AnnounceLocalTrustRequest_Body{
			Trusts: []*reputation.Trust{nil, {}},
		}

		var dst reputation.AnnounceLocalTrustRequest_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ts := dst.GetTrusts()
		require.Len(t, ts, 2)
		require.Equal(t, ts[0], new(reputation.Trust))
		require.Equal(t, ts[1], new(reputation.Trust))
	})

	prototest.TestMarshalStable(t, []*reputation.AnnounceLocalTrustRequest_Body{
		{
			Epoch:  prototest.RandUint64(),
			Trusts: randTrusts(),
		},
	})
}

func TestAnnounceLocalTrustResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.AnnounceLocalTrustResponse_Body{})
}

func TestAnnounceIntermediateResultRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.AnnounceIntermediateResultRequest_Body{
		{
			Epoch:     prototest.RandUint64(),
			Iteration: prototest.RandUint32(),
			Trust:     randPeerToPeerTrust(),
		},
	})
}

func TestAnnounceIntermediateResultResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*reputation.AnnounceIntermediateResultResponse_Body{})
}
