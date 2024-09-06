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
// Header length is limited to [object.MaxHeaderLen].
//
// Operation is executed within a session automatically created by [Pool] unless parameters explicitly override session settings.
//
// See details in [client.Client.ObjectPutInit].
func (p *Pool) ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm client.PrmObjectPutInit) (client.ObjectWriter, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	cnr := hdr.GetContainerID()
	if cnr.IsZero() {
		return nil, cid.ErrZero
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
	c, err := p.sdkClient()
	if err != nil {
		return object.Object{}, nil, err
	}
	return c.ObjectGetInit(ctx, containerID, objectID, signer, prm)
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
	return c.ObjectHead(ctx, containerID, objectID, signer, prm)
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
	return c.ObjectRangeInit(ctx, containerID, objectID, offset, length, signer, prm)
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
	return c.ObjectHash(ctx, containerID, objectID, signer, prm)
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
	return c.ObjectSearchInit(ctx, containerID, signer, prm)
}
