package client_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func ExampleClient_createInstance() {
	// Create client instance
	var prm client.PrmInit
	c, err := client.New(prm)
	_ = err

	// Connect to the NeoFS server
	var prmDial client.PrmDial
	prmDial.SetServerURI("grpc://localhost:8080") // endpoint address
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	_ = c.Dial(prmDial)
}

// Session created for the one node, and it will work only for this node. Other nodes don't have info about this session.
// That is why session can't be created with Pool API.
func ExampleClient_SessionCreate() {
	// import "github.com/google/uuid"

	var prmInit client.PrmInit
	// ...
	c, _ := client.New(prmInit)

	// Epoch when session will expire.
	// Note that expiration starts since exp+1 epoch.
	// For instance, now you have 8 epoch. You set exp=10. The session will be still valid during 10th epoch.
	// Expiration starts since 11 epoch.
	var exp uint64
	var prm client.PrmSessionCreate
	prm.SetExp(exp)

	// The key is generated to simplify the example, in reality it's likely to come from configuration/wallet.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	res, _ := c.SessionCreate(context.Background(), signer, prm)

	var id uuid.UUID
	_ = id.UnmarshalBinary(res.ID())

	// Public key for separate private key, which was created inside node for this session.
	var key neofsecdsa.PublicKey
	_ = key.Decode(res.PublicKey())

	// Fill session parameters
	var sessionObject session.Object
	sessionObject.SetID(id)
	sessionObject.SetAuthKey(&key)
	sessionObject.SetExp(exp)

	// Attach verb and container. Session allows to do just one action by time. In this example it is a VerbObjectPut.
	// If you need Get, Delete, etc you should create another session.
	sessionObject.ForVerb(session.VerbObjectPut)
	// Session works only with one container.
	sessionObject.BindContainer(cid.ID{})

	// Finally, token must be signed by container owner or someone who allowed to do the Verb action. In our example
	// it is VerbObjectPut.
	_ = sessionObject.Sign(signer)

	// ...

	// This token will be used in object put operation
	var prmPut client.PrmObjectPutInit
	prmPut.WithinSession(sessionObject)
	// ...
}
