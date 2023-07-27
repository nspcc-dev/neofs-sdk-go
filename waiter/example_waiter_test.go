package waiter

import (
	"context"
	"strconv"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func ExampleNewWaiter() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// The account was taken from https://github.com/nspcc-dev/neofs-aio.
	key, err := keys.NEP2Decrypt("6PYM8VdX2BSm7BSXKzV4Fz6S3R9cDLLWNrD9nMjxW352jEv3fsC8N3wNLY", "one", keys.NEP2ScryptParams())
	if err != nil {
		panic(err)
	}

	signer := user.NewSignerRFC6979(key.PrivateKey)

	account := signer.UserID()

	var cont container.Container
	var pp netmap.PlacementPolicy
	var rd netmap.ReplicaDescriptor

	// prepare container.
	cont.Init()
	cont.SetBasicACL(acl.PublicRW)
	cont.SetOwner(account)
	cont.SetName(strconv.FormatInt(time.Now().UnixNano(), 16))
	cont.SetCreationTime(time.Now())

	// prepare placement policy.
	pp.SetContainerBackupFactor(1)
	rd.SetNumberOfObjects(1)
	pp.AddReplicas(rd)
	cont.SetPlacementPolicy(pp)

	// prepare pool.
	opts := pool.InitParameters{}
	opts.SetSigner(signer)
	opts.SetClientRebalanceInterval(30 * time.Second)
	opts.AddNode(pool.NewNodeParam(1, "grpc://localhost:8080", 1))

	// init pool.
	p, err := pool.NewPool(opts)
	if err != nil {
		panic(err)
	}

	// create waiter.
	wait := NewWaiter(p, DefaultPollInterval)

	var prmPut client.PrmContainerPut

	// Waiter creates container and wait until container created or context timeout.
	containerID, err := wait.ContainerPut(ctx, cont, signer, prmPut)
	if err != nil {
		panic(err)
	}

	var prmDelete client.PrmContainerDelete

	// Waiter deletes container and wait until container deleted or context timeout.
	err = wait.ContainerDelete(ctx, containerID, signer, prmDelete)
	if err != nil {
		panic(err)
	}
}
