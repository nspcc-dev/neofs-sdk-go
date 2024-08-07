package reputationtest

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

func PeerID() (v reputation.PeerID) {
	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	v.SetPublicKey(p.PublicKey().Bytes())

	return
}

func Trust() (v reputation.Trust) {
	v.SetPeer(PeerID())
	v.SetValue(0.5)

	return
}

func PeerToPeerTrust() (v reputation.PeerToPeerTrust) {
	v.SetTrustingPeer(PeerID())
	v.SetTrust(Trust())

	return
}

func GlobalTrust() (v reputation.GlobalTrust) {
	v.Init()
	v.SetManager(PeerID())
	v.SetTrust(Trust())

	return
}

func SignedGlobalTrust() reputation.GlobalTrust {
	gt := GlobalTrust()

	if err := gt.Sign(neofscryptotest.Signer()); err != nil {
		panic(fmt.Sprintf("unexpected error from GlobalTrust.Sign: %v", err))
	}

	return gt
}
