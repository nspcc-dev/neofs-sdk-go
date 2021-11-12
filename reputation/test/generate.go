package reputationtest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/util/signature"
	"github.com/stretchr/testify/require"
)

func PeerID() *reputation.PeerID {
	v := reputation.NewPeerID()

	p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	key := [signature.PublicKeyCompressedSize]byte{}
	copy(key[:], elliptic.MarshalCompressed(p.Curve, p.X, p.Y))
	v.SetPublicKey(key)

	return v
}

func Trust() *reputation.Trust {
	v := reputation.NewTrust()
	v.SetPeer(PeerID())
	v.SetValue(1.5)

	return v
}

func PeerToPeerTrust() *reputation.PeerToPeerTrust {
	v := reputation.NewPeerToPeerTrust()
	v.SetTrustingPeer(PeerID())
	v.SetTrust(Trust())

	return v
}

func GlobalTrust() *reputation.GlobalTrust {
	v := reputation.NewGlobalTrust()
	v.SetManager(PeerID())
	v.SetTrust(Trust())

	return v
}

func SignedGlobalTrust(t testing.TB) *reputation.GlobalTrust {
	gt := GlobalTrust()

	p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	require.NoError(t, gt.Sign(p))

	return gt
}
