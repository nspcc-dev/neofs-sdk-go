package sessiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

var p *keys.PrivateKey

func init() {
	var err error

	p, err = keys.NewPrivateKey()
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

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	ownerID := owner.NewID()
	ownerID.SetPublicKey(&priv.PublicKey)

	keyBin := p.PublicKey().Bytes()

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

	err := tok.Sign(&p.PrivateKey)
	if err != nil {
		panic(err)
	}

	return tok
}
