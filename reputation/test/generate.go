package reputationtest

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/util/signature"
	"github.com/stretchr/testify/require"
)

func GeneratePeerID() *reputation.PeerID {
	v := reputation.NewPeerID()

	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	key := [signature.PublicKeyCompressedSize]byte{}
	copy(key[:], p.Bytes())
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

	priv, err := keys.NewPrivateKey()
	require.NoError(t, err)
	require.NoError(t, gt.Sign(&priv.PrivateKey))

	return gt
}
