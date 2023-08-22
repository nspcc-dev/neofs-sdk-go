package pool

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type containerSessionParams interface {
	GetSession() (*session.Object, error)
	WithinSession(session.Object)
}

func (p *Pool) actualSigner(signer user.Signer) user.Signer {
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
func (p *Pool) ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm client.PrmObjectPutInit) (client.ObjectWriter, error) {
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

	ow, err := c.ObjectPutInit(ctx, hdr, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return ow, err
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectGetInit].
func (p *Pool) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectGet) (object.Object, *client.PayloadReader, error) {
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

	hdr, payloadReader, err := c.ObjectGetInit(ctx, containerID, objectID, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return hdr, payloadReader, err
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectHead].
func (p *Pool) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectHead) (*object.Object, error) {
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

	hdr, err := c.ObjectHead(ctx, containerID, objectID, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return hdr, err
}

// ObjectRangeInit initiates reading an object's payload range through a remote
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectRangeInit].
func (p *Pool) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer user.Signer, prm client.PrmObjectRange) (*client.ObjectRangeReader, error) {
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

	reader, err := c.ObjectRangeInit(ctx, containerID, objectID, offset, length, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return reader, err
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectDelete].
func (p *Pool) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectDelete) (oid.ID, error) {
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

	id, err := c.ObjectDelete(ctx, containerID, objectID, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return id, err
}

// ObjectHash requests checksum of the range list of the object payload using
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectHash].
func (p *Pool) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm client.PrmObjectHash) ([][]byte, error) {
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

	data, err := c.ObjectHash(ctx, containerID, objectID, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return data, err
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) ObjectSearchInit(ctx context.Context, containerID cid.ID, signer user.Signer, prm client.PrmObjectSearch) (*client.ObjectListReader, error) {
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

	reader, err := c.ObjectSearchInit(ctx, containerID, signer, prm)
	p.checkSessionTokenErr(err, c.addr, c.nodeSession)

	return reader, err
}
