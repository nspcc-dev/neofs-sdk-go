//go:build aiotest

package pool

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func getSigner() neofscrypto.Signer {
	key, err := keys.NEP2Decrypt("6PYM8VdX2BSm7BSXKzV4Fz6S3R9cDLLWNrD9nMjxW352jEv3fsC8N3wNLY", "one", keys.NEP2ScryptParams())
	if err != nil {
		panic(err)
	}

	return neofsecdsa.SignerRFC6979(key.PrivateKey)
}

func TestPoolInterfaceWithAIO(t *testing.T) {
	ctx := context.Background()

	signer := getSigner()
	containerName := "test-1"
	creationTime := time.Now()

	opts := InitParameters{
		signer: signer,
		nodeParams: []NodeParam{
			{1, "grpc://localhost:8080", 1},
		},
		clientRebalanceInterval: 30 * time.Second,
	}

	pool, err := NewPool(opts)
	require.NoError(t, err)
	require.NoError(t, pool.Dial(ctx))

	var account user.ID
	require.NoError(t, user.IDFromSigner(&account, signer))

	var cont container.Container
	cont.Init()
	cont.SetBasicACL(acl.PublicRW)
	cont.SetOwner(account)
	cont.SetName(containerName)
	cont.SetCreationTime(creationTime)

	var containerID cid.ID
	var objectID oid.ID

	payload := make([]byte, 8)
	_, err = rand.Read(payload)
	require.NoError(t, err)

	t.Run("create container", func(t *testing.T) {
		var pp netmap.PlacementPolicy
		pp.SetContainerBackupFactor(1)

		var rd netmap.ReplicaDescriptor
		rd.SetNumberOfObjects(1)
		pp.AddReplicas(rd)

		cont.SetPlacementPolicy(pp)

		var cmd client.PrmContainerPut

		containerID, err = pool.ContainerPut(ctx, cont, cmd)
		require.NoError(t, err)

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = isBucketCreated(ctxTimeout, pool, containerID)
		require.NoError(t, err)
	})

	t.Run("upload object", func(t *testing.T) {
		rf := object.RequiredFields{
			Container: containerID,
			Owner:     account,
		}

		var hdr object.Object
		object.InitCreation(&hdr, rf)

		var prm client.PrmObjectPutInit
		prm.UseSigner(signer)
		prm.SetCopiesNumber(1)

		w, err := pool.ObjectPutInit(context.Background(), hdr, prm)
		require.NoError(t, err)

		require.True(t, w.WritePayloadChunk(payload))

		resp, err := w.Close()
		require.NoError(t, err)

		objectID = resp.StoredObjectID()
	})

	t.Run("download object", func(t *testing.T) {
		var cmd client.PrmObjectGet
		cmd.UseSigner(signer)

		hdr, read, err := pool.ObjectGetInit(ctx, containerID, objectID, cmd)
		defer func() {
			_ = read.Close()
		}()

		require.NoError(t, err)
		require.NotNil(t, hdr.OwnerID())
		require.True(t, hdr.OwnerID().Equals(account))

		downloadedPayload := make([]byte, len(payload))

		l, ok := read.ReadChunk(downloadedPayload)
		require.True(t, ok)
		require.Equal(t, l, len(payload))

		require.True(t, bytes.Equal(payload, downloadedPayload))
	})

	t.Run("delete object", func(t *testing.T) {
		var cmd client.PrmObjectDelete
		cmd.UseSigner(signer)

		_, err = pool.ObjectDelete(ctx, containerID, objectID, cmd)
		require.NoError(t, err)

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = isObjectDeleted(ctxTimeout, pool, containerID, objectID)
		require.NoError(t, err)
	})

	t.Run("container delete", func(t *testing.T) {
		var cmd client.PrmContainerDelete
		cmd.SetSigner(signer)

		require.NoError(t, pool.ContainerDelete(ctx, containerID, cmd))

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = isBucketDeleted(ctxTimeout, pool, containerID)
		require.NoError(t, err)
	})
}

func isBucketCreated(ctx context.Context, c *Pool, id cid.ID) error {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	var cmdGet client.PrmContainerGet

	for {
		select {
		case <-t.C:
			_, err := c.ContainerGet(ctx, id, cmdGet)
			if err != nil {
				if errors.Is(err, apistatus.ErrContainerNotFound) {
					continue
				}

				return fmt.Errorf("ContainerGet %w", err)
			}
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isObjectDeleted(ctx context.Context, c *Pool, id cid.ID, oid oid.ID) error {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	var prmHead client.PrmObjectHead

	for {
		select {
		case <-t.C:
			_, err := c.ObjectHead(ctx, id, oid, prmHead)
			if err != nil {
				if errors.Is(err, apistatus.ErrObjectNotFound) ||
					errors.Is(err, apistatus.ErrObjectAlreadyRemoved) {
					return nil
				}

				return fmt.Errorf("ObjectGetInit %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isBucketDeleted(ctx context.Context, c *Pool, id cid.ID) error {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	var cmdGet client.PrmContainerGet

	for {
		select {
		case <-t.C:
			_, err := c.ContainerGet(ctx, id, cmdGet)
			if err != nil {
				if errors.Is(err, apistatus.ErrContainerNotFound) {
					return nil
				}

				return fmt.Errorf("ContainerGet %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
