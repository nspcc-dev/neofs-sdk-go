package pool

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
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
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	cnr, isSet := hdr.ContainerID()
	if !isSet {
		return nil, errContainerRequired
	}

	if err = p.withinContainerSession(ctx, c, cnr, p.actualSigner(prm.Signer()), session.VerbObjectPut, &prm); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectPutInit(ctx, hdr, prm)
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectGetInit].
func (p *Pool) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectGet) (object.Object, *client.ObjectReader, error) {
	var hdr object.Object
	c, err := p.sdkClient()
	if err != nil {
		return hdr, nil, err
	}

	if err = p.withinContainerSession(ctx, c, containerID, p.actualSigner(prm.Signer()), session.VerbObjectGet, &prm); err != nil {
		return hdr, nil, fmt.Errorf("session: %w", err)
	}

	return c.ObjectGetInit(ctx, containerID, objectID, prm)
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectHead].
func (p *Pool) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectHead) (*client.ResObjectHead, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	return c.ObjectHead(ctx, containerID, objectID, prm)
}

// ObjectRangeInit initiates reading an object's payload range through a remote
//
// See details in [client.Client.ObjectRangeInit].
func (p *Pool) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, prm client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	return c.ObjectRangeInit(ctx, containerID, objectID, offset, length, prm)
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
//
// See details in [client.Client.ObjectDelete].
func (p *Pool) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectDelete) (oid.ID, error) {
	c, err := p.sdkClient()
	if err != nil {
		return oid.ID{}, err
	}

	if err = p.withinContainerSession(ctx, c, containerID, p.actualSigner(prm.Signer()), session.VerbObjectDelete, &prm); err != nil {
		return oid.ID{}, fmt.Errorf("session: %w", err)
	}

	return c.ObjectDelete(ctx, containerID, objectID, prm)
}

// ObjectHash requests checksum of the range list of the object payload using
//
// See details in [client.Client.ObjectHash].
func (p *Pool) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, prm client.PrmObjectHash) ([][]byte, error) {
	c, err := p.sdkClient()
	if err != nil {
		return [][]byte{}, err
	}

	return c.ObjectHash(ctx, containerID, objectID, prm)
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) ObjectSearchInit(ctx context.Context, containerID cid.ID, prm client.PrmObjectSearch) (*client.ObjectListReader, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	return c.ObjectSearchInit(ctx, containerID, prm)
}

// AnnounceLocalTrust sends client's trust values to the NeoFS network participants.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) AnnounceLocalTrust(ctx context.Context, epoch uint64, trusts []reputation.Trust, prm client.PrmAnnounceLocalTrust) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.AnnounceLocalTrust(ctx, epoch, trusts, prm)
}

// AnnounceIntermediateTrust sends global trust values calculated for the specified NeoFS network participants
// at some stage of client's calculation algorithm.
//
// See details in [client.Client.ObjectSearchInit].
func (p *Pool) AnnounceIntermediateTrust(ctx context.Context, epoch uint64, trust reputation.PeerToPeerTrust, prm client.PrmAnnounceIntermediateTrust) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.AnnounceIntermediateTrust(ctx, epoch, trust, prm)
}
