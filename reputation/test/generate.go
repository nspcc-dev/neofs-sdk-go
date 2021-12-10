package reputationtest

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/util/signature"
	"github.com/stretchr/testify/require"
)

func PeerID() *reputation.PeerID {
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

	p, err := keys.NewPrivateKey()
	require.NoError(t, err)
	require.NoError(t, gt.Sign(&p.PrivateKey))

	return gt
}
