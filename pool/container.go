package pool

import (
	"context"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ContainerPut sends request to save container in NeoFS.
//
// See details in [client.Client.ContainerPut].
func (p *Pool) ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm client.PrmContainerPut) (cid.ID, error) {
	c, err := p.sdkClient()
	if err != nil {
		return cid.ID{}, err
	}

	return c.ContainerPut(ctx, cont, signer, prm)
}

// ContainerGet reads NeoFS container by ID.
//
// See details in [client.Client.ContainerGet].
func (p *Pool) ContainerGet(ctx context.Context, id cid.ID, prm client.PrmContainerGet) (container.Container, error) {
	c, err := p.sdkClient()
	if err != nil {
		return container.Container{}, err
	}

	return c.ContainerGet(ctx, id, prm)
}

// ContainerList requests identifiers of the account-owned containers.
//
// See details in [client.Client.ContainerList].
func (p *Pool) ContainerList(ctx context.Context, ownerID user.ID, prm client.PrmContainerList) ([]cid.ID, error) {
	c, err := p.sdkClient()
	if err != nil {
		return []cid.ID{}, err
	}

	return c.ContainerList(ctx, ownerID, prm)
}

// ContainerDelete sends request to remove the NeoFS container.
//
// See details in [client.Client.ContainerDelete].
func (p *Pool) ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm client.PrmContainerDelete) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.ContainerDelete(ctx, id, signer, prm)
}

// ContainerEACL reads eACL table of the NeoFS container.
//
// See details in [client.Client.ContainerEACL].
func (p *Pool) ContainerEACL(ctx context.Context, id cid.ID, prm client.PrmContainerEACL) (eacl.Table, error) {
	c, err := p.sdkClient()
	if err != nil {
		return eacl.Table{}, err
	}

	return c.ContainerEACL(ctx, id, prm)
}

// ContainerSetEACL sends request to update eACL table of the NeoFS container.
//
// See details in [client.Client.ContainerSetEACL].
func (p *Pool) ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm client.PrmContainerSetEACL) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.ContainerSetEACL(ctx, table, signer, prm)
}

// SetContainerAttribute selects a suitable connection from the pool and calls
// [client.Client.SetContainerAttribute] on it.
func (p *Pool) SetContainerAttribute(ctx context.Context, prm client.SetContainerAttributeParameters, prmSig neofscrypto.Signature, opts client.SetContainerAttributeOptions) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.SetContainerAttribute(ctx, prm, prmSig, opts)
}

// RemoveContainerAttribute selects a suitable connection from the pool and calls
// [client.Client.RemoveContainerAttribute] on it.
func (p *Pool) RemoveContainerAttribute(ctx context.Context, prm client.RemoveContainerAttributeParameters, prmSig neofscrypto.Signature, opts client.RemoveContainerAttributeOptions) error {
	c, err := p.sdkClient()
	if err != nil {
		return err
	}

	return c.RemoveContainerAttribute(ctx, prm, prmSig, opts)
}
