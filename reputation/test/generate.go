package reputationtest

import (
	"fmt"
	"math/rand"

	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// PeerID returns random reputation.PeerID.
func PeerID() reputation.PeerID {
	var res reputation.PeerID
	rand.Read(res[:])
	return res
}

// ChangePeerID returns reputation.PeerID other than the given one.
func ChangePeerID(id reputation.PeerID) reputation.PeerID {
	id[0]++
	return id
}

// Trust returns random reputation.Trust.
func Trust() reputation.Trust {
	var res reputation.Trust
	res.SetPeer(PeerID())
	res.SetValue(rand.Float64())
	return res
}

// NTrusts returns n random reputation.Trust instances.
func NTrusts(n int) []reputation.Trust {
	res := make([]reputation.Trust, n)
	for i := range res {
		res[i] = Trust()
	}
	return res
}

// PeerToPeerTrust returns random reputation.PeerToPeerTrust.
func PeerToPeerTrust() reputation.PeerToPeerTrust {
	var v reputation.PeerToPeerTrust
	v.SetTrustingPeer(PeerID())
	v.SetTrust(Trust())
	return v
}

func globalTrustUnsigned() reputation.GlobalTrust {
	return reputation.NewGlobalTrust(PeerID(), Trust())
}

// GlobalTrust returns random reputation.GlobalTrust.
func GlobalTrust() reputation.GlobalTrust {
	tr := globalTrustUnsigned()
	if err := tr.Sign(neofscryptotest.RandomSigner()); err != nil {
		panic(fmt.Errorf("unexpected sign error: %w", err))
	}
	return tr
}

// GlobalTrustUnsigned returns random unsigned reputation.GlobalTrust.
func GlobalTrustUnsigned() reputation.GlobalTrust {
	return globalTrustUnsigned()
}
