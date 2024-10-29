package pool_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func ExampleNew_easiestWay() {
	// The key is generated to simplify the example, in reality it's likely to come from configuration/wallet.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	p, _ := pool.New(
		pool.NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}),
		signer,
		pool.DefaultOptions(),
	)
	_ = p
}

func ExampleNew_adjustingParameters() {
	// The key is generated to simplify the example, in reality it's likely to come from configuration/wallet.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	opts := pool.DefaultOptions()
	opts.SetErrorThreshold(10)

	p, _ := pool.New(
		pool.NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}),
		signer,
		opts,
	)
	_ = p
}

func ExamplePool_ObjectGetInit_explicitAutoSessionDisabling() {
	// The key is generated to simplify the example, in reality it's likely to come from configuration/wallet.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	p, _ := pool.New(
		pool.NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}),
		signer,
		pool.DefaultOptions(),
	)

	_ = p.Dial(context.Background())

	var prm client.PrmObjectGet
	// If you don't provide the session manually with prm.WithinSession, the request will be executed without session.
	prm.IgnoreSession()

	var ownerID user.ID
	var hdr object.Object
	hdr.SetContainerID(cid.ID{})
	hdr.SetOwner(ownerID)

	var containerID cid.ID
	// fill containerID
	var objetID oid.ID
	// fill objectID

	// In case of a session wasn't provided with prm.WithinSession, the signer must be for account which is a container
	// owner, otherwise there will be an error.

	// In case of a session was provided with prm.WithinSession, the signer can be ether container owner account or
	// third party, who can use a session token signed by container owner.
	_, _, _ = p.ObjectGetInit(context.Background(), containerID, objetID, signer, prm)

	// ...
}

func ExamplePool_ObjectPutInit_autoSessionDisabling() {
	// The key is generated to simplify the example, in reality it's likely to come from configuration/wallet.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	p, _ := pool.New(
		pool.NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}),
		signer,
		pool.DefaultOptions(),
	)

	_ = p.Dial(context.Background())

	// Session should be initialized with Client.SessionCreate function.
	var sess session.Object

	var prm client.PrmObjectPutInit
	// Auto-session disabled, because you provided session already.
	prm.WithinSession(sess)

	// For ObjectPutInit operation prm.IgnoreSession shouldn't be called ever, because putObject without session is not
	// acceptable, and it will be an error.
	// prm.IgnoreSession()

	var ownerID user.ID
	var hdr object.Object
	hdr.SetContainerID(cid.ID{})
	hdr.SetOwner(ownerID)

	// In case of a session wasn't provided with prm.WithinSession, the signer must be for account which is a container
	// owner, otherwise there will be an error.

	// In case of a session was provided with prm.WithinSession, the signer can be ether container owner account or
	// third party, who can use a session token signed by container owner.
	_, _ = p.ObjectPutInit(context.Background(), hdr, signer, prm)

	// ...
}
