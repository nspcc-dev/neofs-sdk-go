package sessiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"math/rand"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

var p *ecdsa.PrivateKey

func init() {
	var err error

	p, err = ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		panic(err)
	}
}

// Token returns random session.Token.
//
// Resulting token is unsigned.
func Token() *session.Token {
	tok := session.NewToken()

	uid, err := uuid.New().MarshalBinary()
	if err != nil {
		panic(err)
	}

	w := new(owner.NEO3Wallet)
	rand.Read(w.Bytes())

	ownerID := owner.NewID()
	ownerID.SetNeo3Wallet(w)

	keyBin := elliptic.MarshalCompressed(p.PublicKey.Curve, p.PublicKey.X, p.PublicKey.Y)

	tok.SetID(uid)
	tok.SetOwnerID(ownerID)
	tok.SetSessionKey(keyBin)
	tok.SetExp(11)
	tok.SetNbf(22)
	tok.SetIat(33)

	return tok
}

// SignedToken returns signed random session.Token.
//
// Panics if token could not be signed (actually unexpected).
func SignedToken() *session.Token {
	tok := Token()

	err := tok.Sign(p)
	if err != nil {
		panic(err)
	}

	return tok
}
