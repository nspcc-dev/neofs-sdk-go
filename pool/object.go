package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type containerSessionParams interface {
	IsSessionSet() bool
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
// See details in [client.Client.ObjectPutInit].
func (p *Pool) ObjectPutInit(ctx context.Context, hdr object.Object, prm client.PrmObjectPutInit) (*client.ObjectWriter, error) {
	c, statUpdater, err := p.sdkClient()
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
		p.actualSigner(prm.Signer()),
		session.VerbObjectPut,
		&prm,
		statUpdater,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	w, err := c.ObjectPutInit(ctx, hdr, prm)
	statUpdater.incRequests(time.Since(start), methodObjectPut)
	statUpdater.updateErrorRate(err)

	return w, err
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectGetInit].
func (p *Pool) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectGet) (object.Object, *client.ObjectReader, error) {
	var hdr object.Object
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return hdr, nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectGet,
		&prm,
		statUpdater,
	); err != nil {
		return hdr, nil, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	obj, reader, err := c.ObjectGetInit(ctx, containerID, objectID, prm)
	statUpdater.incRequests(time.Since(start), methodObjectGet)
	statUpdater.updateErrorRate(err)

	return obj, reader, err
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectHead].
func (p *Pool) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectHead) (*client.ResObjectHead, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectHead,
		&prm,
		statUpdater,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	head, err := c.ObjectHead(ctx, containerID, objectID, prm)
	statUpdater.incRequests(time.Since(start), methodObjectHead)
	statUpdater.updateErrorRate(err)

	return head, err
}

// ObjectRangeInit initiates reading an object's payload range through a remote
//
// See details in [client.Client.ObjectRangeInit].
func (p *Pool) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, prm client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectRange,
		&prm,
		statUpdater,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	reader, err := c.ObjectRangeInit(ctx, containerID, objectID, offset, length, prm)
	statUpdater.incRequests(time.Since(start), methodObjectRange)
	statUpdater.updateErrorRate(err)

	return reader, err
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
//
// See details in [client.Client.ObjectDelete].
func (p *Pool) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectDelete) (oid.ID, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return oid.ID{}, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectDelete,
		&prm,
		statUpdater,
	); err != nil {
		return oid.ID{}, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	id, err := c.ObjectDelete(ctx, containerID, objectID, prm)
	statUpdater.incRequests(time.Since(start), methodObjectDelete)
	statUpdater.updateErrorRate(err)

	return id, err
}

// ObjectHash requests checksum of the range list of the object payload using
//
// See details in [client.Client.ObjectHash].
func (p *Pool) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectHash) ([][]byte, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return [][]byte{}, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectRangeHash,
		&prm,
		statUpdater,
	); err != nil {
		return [][]byte{}, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	hashes, err := c.ObjectHash(ctx, containerID, objectID, prm)
	statUpdater.incRequests(time.Since(start), methodObjectHash)
	statUpdater.updateErrorRate(err)

	return hashes, err
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) ObjectSearchInit(ctx context.Context, containerID cid.ID, prm client.PrmObjectSearch) (*client.ObjectListReader, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return nil, err
	}
	if err = p.withinContainerSession(
		ctx,
		c,
		containerID,
		p.actualSigner(prm.Signer()),
		session.VerbObjectSearch,
		&prm,
		statUpdater,
	); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	start := time.Now()
	reader, err := c.ObjectSearchInit(ctx, containerID, prm)
	statUpdater.incRequests(time.Since(start), methodObjectSearch)
	statUpdater.updateErrorRate(err)

	return reader, err
}
