package sessiontest

import (
	"math/rand"

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

// Generate returns random session.Token.
//
// Resulting token is unsigned.
func Generate() *session.Token {
	tok := session.NewToken()

	uid, err := uuid.New().MarshalBinary()
	if err != nil {
		panic(err)
	}

	w := new(owner.NEO3Wallet)
	rand.Read(w.Bytes())

	ownerID := owner.NewID()
	ownerID.SetNeo3Wallet(w)

	keyBin := p.PublicKey().Bytes()

	tok.SetID(uid)
	tok.SetOwnerID(ownerID)
	tok.SetSessionKey(keyBin)
	tok.SetExp(11)
	tok.SetNbf(22)
	tok.SetIat(33)

	return tok
}

// GenerateSigned returns signed random session.Token.
//
// Panics if token could not be signed (actually unexpected).
func GenerateSigned() *session.Token {
	tok := Generate()

	err := tok.Sign(&p.PrivateKey)
	if err != nil {
		panic(err)
	}

	return tok
}
