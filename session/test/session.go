package sessiontest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	mrand "math/rand"

	"github.com/google/uuid"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

func container(sign bool) session.Container {
	var tok session.Container

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	tok.ForVerb(session.ContainerVerb(mrand.Uint32() % 20))
	tok.ApplyOnlyTo(cidtest.ID())
	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(&priv.PublicKey))
	tok.SetExp(mrand.Uint64())
	tok.SetNbf(mrand.Uint64())
	tok.SetIat(mrand.Uint64())
	tok.SetIssuer(usertest.ID())

	if sign {
		if err = tok.Sign(user.NewAutoIDSigner(*priv)); err != nil {
			panic(fmt.Errorf("unexpected sign error: %w", err))
		}
	}

	if err = tok.Unmarshal(tok.Marshal()); err != nil { // to fill utility fields
		panic(fmt.Errorf("unexpected container session encode-decode failure: %w", err))
	}

	return tok
}

// Container returns random session.Container.
func Container() session.Container {
	return container(true)
}

// ContainerUnsigned returns random unsigned session.Container.
func ContainerUnsigned() session.Container {
	return container(false)
}

func object(sign bool) session.Object {
	var tok session.Object

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	tok.ForVerb(session.ObjectVerb(mrand.Uint32() % 20))
	tok.BindContainer(cidtest.ID())
	tok.LimitByObjects(oidtest.NIDs(1 + mrand.Int()%3))
	tok.SetID(uuid.New())
	tok.SetAuthKey((*neofsecdsa.PublicKey)(&priv.PublicKey))
	tok.SetExp(mrand.Uint64())
	tok.SetNbf(mrand.Uint64())
	tok.SetIat(mrand.Uint64())
	tok.SetIssuer(usertest.ID())

	if sign {
		if err = tok.Sign(user.NewAutoIDSigner(*priv)); err != nil {
			panic(fmt.Errorf("unexpected sign error: %w", err))
		}
	}

	if err = tok.Unmarshal(tok.Marshal()); err != nil { // to fill utility fields
		panic(fmt.Errorf("unexpected object session encode-decode failure: %w", err))
	}

	return tok
}

// Object returns random session.Object.
func Object() session.Object { return object(true) }

// ObjectUnsigned returns random unsigned session.Object.
func ObjectUnsigned() session.Object { return object(false) }
