package reputation_test

import (
	"math/rand"
	"testing"

	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTrustDecoding(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(trust *apireputation.Trust)
		}{
			{name: "peer", err: "missing peer", corrupt: func(tr *apireputation.Trust) {
				tr.Peer = nil
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst reputation.Trust
				var m apireputation.Trust

				// set required data just to not collide with other cases
				src.SetPeer(reputationtest.PeerID())

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)
			})
		}
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(trust *apireputation.Trust)
		}{
			{name: "value/negative", err: "invalid trust value -0.1", corrupt: func(tr *apireputation.Trust) {
				tr.Value = -0.1
			}},
			{name: "value/overflow", err: "invalid trust value 1.1", corrupt: func(tr *apireputation.Trust) {
				tr.Value = 1.1
			}},
			{name: "peer/value/nil", err: "invalid peer: missing value field", corrupt: func(tr *apireputation.Trust) {
				tr.Peer.PublicKey = nil
			}},
			{name: "peer/value/empty", err: "invalid peer: missing value field", corrupt: func(tr *apireputation.Trust) {
				tr.Peer.PublicKey = []byte{}
			}},
			{name: "peer/value/wrong length", err: "invalid peer: invalid value length 32", corrupt: func(tr *apireputation.Trust) {
				tr.Peer.PublicKey = make([]byte, 32)
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst reputation.Trust
				var m apireputation.Trust

				// set required data just to not collide with other cases
				src.SetPeer(reputationtest.PeerID())

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)
			})
		}
	})
}

func TestTrust_SetPeer(t *testing.T) {
	var tr reputation.Trust

	require.Zero(t, tr.Peer())

	peer := reputationtest.PeerID()
	tr.SetPeer(peer)
	require.Equal(t, peer, tr.Peer())

	peerOther := reputationtest.ChangePeerID(peer)
	tr.SetPeer(peerOther)
	require.Equal(t, peerOther, tr.Peer())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst reputation.Trust
			var msg apireputation.Trust

			src.WriteToV2(&msg)
			require.Zero(t, msg.Peer)
			require.ErrorContains(t, dst.ReadFromV2(&msg), "missing peer")

			dst.SetPeer(peerOther)
			src.SetPeer(peer)
			src.WriteToV2(&msg)
			require.Equal(t, &apireputation.PeerID{PublicKey: peer[:]}, msg.Peer)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, peer, dst.Peer())
		})
	})
}

func TestTrust_SetValue(t *testing.T) {
	var tr reputation.Trust

	require.Zero(t, tr.Value())
	require.Panics(t, func() { tr.SetValue(-0.1) })
	require.Panics(t, func() { tr.SetValue(1.1) })

	const val = 0.5
	tr.SetValue(val)
	require.EqualValues(t, val, tr.Value())

	const valOther = val + 0.1
	tr.SetValue(valOther)
	require.EqualValues(t, valOther, tr.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst reputation.Trust
			var msg apireputation.Trust

			// set required data just to satisfy decoder
			src.SetPeer(reputationtest.PeerID())

			dst.SetValue(val)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Value())

			src.SetValue(val)
			src.WriteToV2(&msg)
			require.EqualValues(t, val, msg.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.EqualValues(t, val, dst.Value())
		})
	})
}

func TestPeerToPeerTrustDecoding(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(trust *apireputation.PeerToPeerTrust)
		}{
			{name: "trusting peer", err: "missing trusting peer", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.TrustingPeer = nil
			}},
			{name: "value", err: "missing trust", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust = nil
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst reputation.PeerToPeerTrust
				var m apireputation.PeerToPeerTrust

				// set required data just to not collide with other cases
				src.SetTrustingPeer(reputationtest.PeerID())
				src.SetTrust(reputationtest.Trust())

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)
			})
		}
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(trust *apireputation.PeerToPeerTrust)
		}{
			{name: "trusting peer/value/nil", err: "invalid trusting peer: missing value field", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.TrustingPeer.PublicKey = nil
			}},
			{name: "trusting peer/value/empty", err: "invalid trusting peer: missing value field", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.TrustingPeer.PublicKey = []byte{}
			}},
			{name: "trusting peer/value/wrong length", err: "invalid trusting peer: invalid value length 32", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.TrustingPeer.PublicKey = make([]byte, 32)
			}},
			{name: "trust/value/negative", err: "invalid trust: invalid trust value -0.1", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust.Value = -0.1
			}},
			{name: "trust/value/overflow", err: "invalid trust: invalid trust value -0.1", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust.Value = -0.1
			}},
			{name: "trust/peer/value/nil", err: "invalid trust: invalid peer: missing value field", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust.Peer.PublicKey = nil
			}},
			{name: "trust/peer/value/empty", err: "invalid trust: invalid peer: missing value field", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust.Peer.PublicKey = []byte{}
			}},
			{name: "trust/peer/value/wrong length", err: "invalid trust: invalid peer: invalid value length 32", corrupt: func(tr *apireputation.PeerToPeerTrust) {
				tr.Trust.Peer.PublicKey = make([]byte, 32)
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst reputation.PeerToPeerTrust
				var m apireputation.PeerToPeerTrust

				// set required data just to not collide with other cases
				src.SetTrustingPeer(reputationtest.PeerID())
				src.SetTrust(reputationtest.Trust())

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)
			})
		}
	})
}

func TestPeerToPeerTrust_SetTrustingPeer(t *testing.T) {
	var tr reputation.PeerToPeerTrust

	require.Zero(t, tr.TrustingPeer())

	peer := reputationtest.PeerID()
	tr.SetTrustingPeer(peer)
	require.EqualValues(t, peer, tr.TrustingPeer())

	otherPeer := reputationtest.ChangePeerID(peer)
	tr.SetTrustingPeer(otherPeer)
	require.EqualValues(t, otherPeer, tr.TrustingPeer())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst reputation.PeerToPeerTrust
			var msg apireputation.PeerToPeerTrust

			// set required data just to satisfy decoder
			src.SetTrust(reputationtest.Trust())

			src.WriteToV2(&msg)
			require.Zero(t, msg.TrustingPeer)
			require.ErrorContains(t, dst.ReadFromV2(&msg), "missing trusting peer")

			src.SetTrustingPeer(peer)
			src.WriteToV2(&msg)
			require.Equal(t, &apireputation.PeerID{PublicKey: peer[:]}, msg.TrustingPeer)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, peer, dst.TrustingPeer())
		})
	})
}

func TestPeerToPeerTrust_SetTrust(t *testing.T) {
	var tr reputation.PeerToPeerTrust

	require.Zero(t, tr.Trust())

	peer := reputationtest.PeerID()
	val := rand.Float64()
	var trust reputation.Trust
	trust.SetPeer(peer)
	trust.SetValue(val)

	tr.SetTrust(trust)
	require.EqualValues(t, trust, tr.Trust())

	otherVal := reputationtest.Trust()
	tr.SetTrust(otherVal)
	require.EqualValues(t, otherVal, tr.Trust())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst reputation.PeerToPeerTrust
			var msg apireputation.PeerToPeerTrust

			// set required data just to satisfy decoder
			src.SetTrustingPeer(reputationtest.PeerID())

			src.WriteToV2(&msg)
			require.Zero(t, msg.Trust)
			require.ErrorContains(t, dst.ReadFromV2(&msg), "missing trust")

			src.SetTrust(trust)
			src.WriteToV2(&msg)
			require.Equal(t, &apireputation.Trust{
				Peer:  &apireputation.PeerID{PublicKey: peer[:]},
				Value: val,
			}, msg.Trust)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, trust, dst.Trust())
		})
	})
}

func TestGlobalTrustDecoding(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var tr reputation.GlobalTrust
		msg := []byte("definitely_not_protobuf")
		err := tr.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(trust *apireputation.GlobalTrust)
		}{
			{name: "missing version", err: "missing version", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Version = nil
			}},
			{name: "missing body", err: "missing body", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body = nil
			}},
			{name: "body/manager/missing", err: "missing manager", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Manager = nil
			}},
			{name: "body/manager/value/nil", err: "invalid manager: missing value field", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Manager.PublicKey = nil
			}},
			{name: "body/manager/value/empty", err: "invalid manager: missing value field", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Manager.PublicKey = []byte{}
			}},
			{name: "body/manager/peer/value/wrong length", err: "invalid manager: invalid value length 32", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Manager.PublicKey = make([]byte, 32)
			}},
			{name: "body/trust/peer/value/nil", err: "invalid trust: invalid peer: missing value field", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Trust.Peer.PublicKey = nil
			}},
			{name: "body/trust/peer/value/empty", err: "invalid trust: invalid peer: missing value field", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Trust.Peer.PublicKey = []byte{}
			}},
			{name: "body/trust/peer/value/wrong length", err: "invalid trust: invalid peer: invalid value length 32", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Trust.Peer.PublicKey = make([]byte, 32)
			}},
			{name: "body/trust/value/negative", err: "invalid trust: invalid trust value -0.1", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Trust.Value = -0.1
			}},
			{name: "body/trust/value/overflow", err: "invalid trust: invalid trust value 1.1", corrupt: func(tr *apireputation.GlobalTrust) {
				tr.Body.Trust.Value = 1.1
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				src := reputation.NewGlobalTrust(reputationtest.PeerID(), reputationtest.Trust())
				var dst reputation.GlobalTrust
				var m apireputation.GlobalTrust

				// set required data just to not collide with other cases
				src.SetManager(reputationtest.PeerID())
				src.SetTrust(reputationtest.Trust())

				require.NoError(t, proto.Unmarshal(src.Marshal(), &m))
				testCase.corrupt(&m)
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.Unmarshal(b), testCase.err)
			})
		}
	})
}

func TestNewGlobalTrust(t *testing.T) {
	peer := reputationtest.PeerID()
	tr := reputationtest.Trust()

	gt := reputation.NewGlobalTrust(peer, tr)
	require.False(t, gt.VerifySignature())
	require.Equal(t, peer, gt.Manager())
	require.Equal(t, tr, gt.Trust())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			src := reputation.NewGlobalTrust(peer, tr)
			dst := reputationtest.GlobalTrust()

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, tr, gt.Trust())
			require.Equal(t, peer, gt.Manager())
		})
	})
}

func TestGlobalTrust_Sign(t *testing.T) {
	var tr reputation.GlobalTrust

	require.False(t, tr.VerifySignature())

	usr, otherUsr := usertest.TwoUsers()

	require.Error(t, tr.Sign(usertest.FailSigner(usr)))
	require.False(t, tr.VerifySignature())
	require.Error(t, tr.Sign(usertest.FailSigner(otherUsr)))
	require.False(t, tr.VerifySignature())

	require.NoError(t, tr.Sign(usr))
	require.True(t, tr.VerifySignature())

	require.NoError(t, tr.Sign(otherUsr))
	require.True(t, tr.VerifySignature())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			src := reputationtest.GlobalTrustUnsigned()
			var dst reputation.GlobalTrust

			require.NoError(t, dst.Sign(usr))
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.VerifySignature())

			require.NoError(t, dst.Sign(otherUsr))
			require.NoError(t, src.Sign(usr))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())

			require.NoError(t, src.Sign(otherUsr))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
		})
	})
}

func TestGlobalTrust_SetManager(t *testing.T) {
	var tr reputation.GlobalTrust

	require.Zero(t, tr.Manager())

	peer := reputationtest.PeerID()
	tr.SetManager(peer)
	require.Equal(t, peer, tr.Manager())

	otherPeer := reputationtest.ChangePeerID(peer)
	tr.SetManager(otherPeer)
	require.Equal(t, otherPeer, tr.Manager())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			src := reputation.NewGlobalTrust(peer, reputationtest.Trust())
			var dst reputation.GlobalTrust

			dst.SetManager(otherPeer)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Equal(t, peer, dst.Manager())
		})
	})
}

func TestGlobalTrust_SetTrust(t *testing.T) {
	var tr reputation.GlobalTrust

	require.Zero(t, tr.Trust())

	val := reputationtest.Trust()
	tr.SetTrust(val)
	require.Equal(t, val, tr.Trust())

	otherVal := reputationtest.Trust()
	tr.SetTrust(otherVal)
	require.Equal(t, otherVal, tr.Trust())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			src := reputation.NewGlobalTrust(reputationtest.PeerID(), val)
			var dst reputation.GlobalTrust

			dst.SetTrust(otherVal)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Equal(t, val, dst.Trust())
		})
	})
}
