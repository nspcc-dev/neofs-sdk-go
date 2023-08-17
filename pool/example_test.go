package pool_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/waiter"
)

// Create pool instance with 3 nodes connection.
// This InitParameters will make pool use 192.168.130.71 node while it is healthy.
// Otherwise, it will make the pool use 192.168.130.72 for 90% of requests and 192.168.130.73 for remaining 10%.
func ExampleNewPool() {
	// import "github.com/nspcc-dev/neofs-sdk-go/user"

	var signer user.Signer
	var prm pool.InitParameters
	prm.SetSigner(signer)
	prm.AddNode(pool.NewNodeParam(1, "192.168.130.71", 1))
	prm.AddNode(pool.NewNodeParam(2, "192.168.130.72", 9))
	prm.AddNode(pool.NewNodeParam(2, "192.168.130.73", 1))
	// ...

	p, err := pool.NewPool(prm)

	_ = p
	_ = err
	// ...
}

func ExampleNew_easiestWay() {
	// Signer generation, like example.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := user.NewAutoIDSignerRFC6979(*pk)

	p, _ := pool.New(
		pool.NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}),
		signer,
		pool.DefaultOptions(),
	)
	_ = p

	// ...
}

func ExampleNew_adjustingParameters() {
	// Signer generation, like example.
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

	// ...
}

func ExamplePool_ObjectPutInit_explicitAutoSessionDisabling() {
	var prm client.PrmObjectPutInit

	// If you don't provide the session manually with prm.WithinSession, the request will be executed without session.
	prm.IgnoreSession()
	// ...
}

func ExamplePool_ObjectPutInit_autoSessionDisabling() {
	var sess session.Object
	var prm client.PrmObjectPutInit

	// Auto-session disabled, because you provided session already.
	prm.WithinSession(sess)

	// ...
}

// Connect to the NeoFS server.
func ExamplePool_Dial() {
	var p pool.Pool

	// Connect to the NeoFS server
	_ = p.Dial(context.Background())
}

func ExamplePool_ContainerPut() {
	// import "github.com/nspcc-dev/neofs-sdk-go/waiter"
	// import "github.com/nspcc-dev/neofs-sdk-go/container"
	// import neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"

	var p pool.Pool
	// ... init pool

	// Connect to the NeoFS server
	_ = p.Dial(context.Background())

	var cont container.Container
	// ... fill container

	var signer neofscrypto.Signer
	// ... create signer

	var prmPut client.PrmContainerPut
	// ... fill params, if required

	// waits until container created or context canceled.
	w := waiter.NewContainerPutWaiter(&p, waiter.DefaultPollInterval)

	containerID, err := w.ContainerPut(context.Background(), cont, signer, prmPut)

	_ = containerID
	_ = err
}

func ExamplePool_ObjectHead() {
	// import "github.com/nspcc-dev/neofs-sdk-go/waiter"
	// import "github.com/nspcc-dev/neofs-sdk-go/container"
	// import "github.com/nspcc-dev/neofs-sdk-go/user"

	var p pool.Pool
	// ... init pool

	// Connect to the NeoFS server
	_ = p.Dial(context.Background())

	var signer user.Signer
	// ... create signer

	var prmHead client.PrmObjectHead
	// ... fill params, if required

	hdr, err := p.ObjectHead(context.Background(), cid.ID{}, oid.ID{}, signer, prmHead)
	if err != nil {
		if errors.Is(err, apistatus.ErrObjectNotFound) {
			return
		}

		// ...
	}

	_ = hdr

	p.Close()
}
