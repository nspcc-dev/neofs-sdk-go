package sessiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

var p ecdsa.PrivateKey

func init() {
	k, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	p = k.PrivateKey
}

// Container returns random session.Container.
//
// Resulting token is unsigned.
func Container() *session.Container {
	var tok session.Container

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	tok.ForVerb(session.VerbContainerPut)
	tok.ApplyOnlyTo(cidtest.ID())
	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(&priv.PublicKey))
	tok.SetExp(11)
	tok.SetNbf(22)
	tok.SetIat(33)

	return &tok
}

// ContainerSigned returns signed random session.Container.
//
// Panics if token could not be signed (actually unexpected).
func ContainerSigned() *session.Container {
	tok := Container()

	err := tok.Sign(p)
	if err != nil {
		panic(err)
	}

	return tok
}

// Object returns random session.Object.
//
// Resulting token is unsigned.
func Object() *session.Object {
	var tok session.Object

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	tok.ForVerb(session.VerbObjectPut)
	tok.BindContainer(cidtest.ID())
	tok.LimitByObjects(oidtest.ID(), oidtest.ID())
	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(&priv.PublicKey))
	tok.SetExp(11)
	tok.SetNbf(22)
	tok.SetIat(33)

	return &tok
}

// ObjectSigned returns signed random session.Object.
//
// Panics if token could not be signed (actually unexpected).
func ObjectSigned() *session.Object {
	tok := Object()

	err := tok.Sign(p)
	if err != nil {
		panic(err)
	}

	return tok
}
