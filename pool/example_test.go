package pool_test

import (
	"context"
	"errors"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
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
