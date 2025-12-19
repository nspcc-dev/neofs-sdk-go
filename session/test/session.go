package sessiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Container returns random session.Container.
//
// Resulting token is unsigned.
func Container() session.Container {
	var tok session.Container

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	tok.ForVerb(session.VerbContainerPut)
	tok.ApplyOnlyTo(cidtest.ID())
	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(&priv.PublicKey))
	tok.SetIat(11)
	tok.SetNbf(22)
	tok.SetExp(33)

	return tok
}

// ContainerSigned returns signed random session.Container.
//
// Panics if token could not be signed (actually unexpected).
func ContainerSigned(signer user.Signer) session.Container {
	tok := Container()

	err := tok.Sign(signer)
	if err != nil {
		panic(err)
	}

	return tok
}

// Object returns random session.Object.
//
// Resulting token is unsigned.
func Object() session.Object {
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
	tok.SetIat(11)
	tok.SetNbf(22)
	tok.SetExp(33)

	return tok
}

// ObjectSigned returns signed random session.Object.
//
// Panics if token could not be signed (actually unexpected).
func ObjectSigned(signer user.Signer) session.Object {
	tok := Object()

	err := tok.Sign(signer)
	if err != nil {
		panic(err)
	}

	return tok
}

// Token returns random session.Token.
//
// Resulting token is unsigned.
func Token() sessionv2.Token {
	var tok sessionv2.Token

	tok.SetVersion(sessionv2.TokenCurrentVersion)
	tok.SetNonce(sessionv2.RandomNonce())
	err := tok.AddSubject(sessionv2.NewTargetUser(usertest.ID()))
	if err != nil {
		panic(err)
	}
	now := time.Now()
	tok.SetIat(now)
	tok.SetNbf(now)
	tok.SetExp(now.Add(time.Hour))

	ctx, err := sessionv2.NewContext(cidtest.ID(), []sessionv2.Verb{
		sessionv2.VerbObjectPut,
		sessionv2.VerbObjectGet,
	})
	if err != nil {
		panic(err)
	}
	err = tok.AddContext(ctx)
	if err != nil {
		panic(err)
	}

	return tok
}

// TokenSigned returns signed random session.Token.
//
// Panics if token could not be signed (actually unexpected).
func TokenSigned(signer user.Signer) sessionv2.Token {
	tok := Token()

	err := tok.Sign(signer)
	if err != nil {
		panic(err)
	}

	return tok
}
