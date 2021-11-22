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

func GeneratePeerID() *reputation.PeerID {
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

func GenerateTrust() *reputation.Trust {
	v := reputation.NewTrust()
	v.SetPeer(GeneratePeerID())
	v.SetValue(1.5)

	return v
}

func GeneratePeerToPeerTrust() *reputation.PeerToPeerTrust {
	v := reputation.NewPeerToPeerTrust()
	v.SetTrustingPeer(GeneratePeerID())
	v.SetTrust(GenerateTrust())

	return v
}

func GenerateGlobalTrust() *reputation.GlobalTrust {
	v := reputation.NewGlobalTrust()
	v.SetManager(GeneratePeerID())
	v.SetTrust(GenerateTrust())

	return v
}

func GenerateSignedGlobalTrust(t testing.TB) *reputation.GlobalTrust {
	gt := GenerateGlobalTrust()

	p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	require.NoError(t, gt.Sign(p))

	return gt
}
