package pool

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type containerSessionParams interface {
	GetSession() (*session.Object, error)
	WithinSession(session.Object)
}

func (p *Pool) actualSigner(signer neofscrypto.Signer) neofscrypto.Signer {
	if signer != nil {
		return signer
	}

	return p.signer
}

// ObjectPutInit initiates writing an object through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectPutInit].
func (p *Pool) ObjectPutInit(ctx context.Context, hdr object.Object, signer neofscrypto.Signer, prm client.PrmObjectPutInit) (*client.ObjectWriter, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	cnr, isSet := hdr.ContainerID()
	if !isSet {
		return nil, errContainerRequired
	}

	if err = p.withinContainerSession(
		ctx,
		c,
		cnr,
		p.actualSigner(signer),
		session.VerbObjectPut,
		&prm,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectPutInit(ctx, hdr, signer, prm)
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectGetInit].
func (p *Pool) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm client.PrmObjectGet) (object.Object, *client.ObjectReader, error) {
	var hdr object.Object
	c, err := p.sdkClient()
	if err != nil {
		return hdr, nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectGet,
		&prm,
	); err != nil {
		return hdr, nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectGetInit(ctx, containerID, objectID, signer, prm)
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectHead].
func (p *Pool) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm client.PrmObjectHead) (*client.ResObjectHead, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectHead,
		&prm,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectHead(ctx, containerID, objectID, signer, prm)
}

// ObjectRangeInit initiates reading an object's payload range through a remote
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectRangeInit].
func (p *Pool) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer neofscrypto.Signer, prm client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectRange,
		&prm,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectRangeInit(ctx, containerID, objectID, offset, length, signer, prm)
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectDelete].
func (p *Pool) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm client.PrmObjectDelete) (oid.ID, error) {
	c, err := p.sdkClient()
	if err != nil {
		return oid.ID{}, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectDelete,
		&prm,
	); err != nil {
		return oid.ID{}, fmt.Errorf("session: %w", err)
	}

	return c.ObjectDelete(ctx, containerID, objectID, signer, prm)
}

// ObjectHash requests checksum of the range list of the object payload using
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectHash].
func (p *Pool) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm client.PrmObjectHash) ([][]byte, error) {
	c, err := p.sdkClient()
	if err != nil {
		return [][]byte{}, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectRangeHash,
		&prm,
	); err != nil {
		return [][]byte{}, fmt.Errorf("session: %w", err)
	}

	return c.ObjectHash(ctx, containerID, objectID, signer, prm)
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) ObjectSearchInit(ctx context.Context, containerID cid.ID, signer neofscrypto.Signer, prm client.PrmObjectSearch) (*client.ObjectListReader, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(signer),
		session.VerbObjectSearch,
		&prm,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectSearchInit(ctx, containerID, signer, prm)
}
