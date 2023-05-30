package poolv2

import (
	"bytes"
	"context"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

func initSigner() neofscrypto.Signer {
	pk, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	return neofsecdsa.Signer(pk.PrivateKey)
}

func initPool(signer neofscrypto.Signer) *Pool {
	var nodes []NodeParam

	nodes = append(nodes, NewNodeParam("grpc://localhost:8080", 1, 1.0))

	p, err := NewPool(nodes, signer)
	if err != nil {
		panic(err)
	}

	return p
}

func ExampleNewPool() {
	ctx := context.Background()

	signer := initSigner()
	p := initPool(signer)

	var owner user.ID
	cont := NewCreateContainerBuilder(owner).
		WithAcl(acl.PublicRO).
		// hide complicated internal structure
		WithPPFilter("name", "value", "key", netmap.EQ, nil).
		WithPPBackupFactor(2)

	containerID, err := p.CreateContainer(ctx, cont)
	if err != nil {
		panic(err)
	}

	var objectID oid.ID

	payload := bytes.NewReader([]byte{1, 2, 3, 4, 5})
	obj := NewCreateObjectBuilder(containerID, payload).
		WithAttribute(Attribute{Name: "is_it_ok", Value: "it_is_ok"}).WithIDHandler(func(id oid.ID) {
		objectID = id
	})

	if err = p.CreateObject(ctx, obj); err != nil {
		panic(err)
	}

	writer := bytes.NewBuffer(nil)

	readBuilder := NewReadObjectBuilder(containerID, objectID, writer)
	if err = p.ReadObject(ctx, readBuilder); err != nil {
		panic(err)
	}
}
