package pool

import (
	"context"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ContainerPut sends request to save container in NeoFS.
//
// See details in [client.Client.ContainerPut].
func (p *Pool) ContainerPut(ctx context.Context, cont container.Container, prm client.PrmContainerPut) (cid.ID, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return cid.ID{}, err
	}

	start := time.Now()
	id, err := c.ContainerPut(ctx, cont, prm)
	statUpdater.incRequests(time.Since(start), methodContainerPut)
	statUpdater.updateErrorRate(err)

	return id, err
}

// ContainerGet reads NeoFS container by ID.
//
// See details in [client.Client.ContainerGet].
func (p *Pool) ContainerGet(ctx context.Context, id cid.ID, prm client.PrmContainerGet) (container.Container, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return container.Container{}, err
	}

	start := time.Now()
	cnr, err := c.ContainerGet(ctx, id, prm)
	statUpdater.incRequests(time.Since(start), methodContainerGet)
	statUpdater.updateErrorRate(err)

	return cnr, err
}

// ContainerList requests identifiers of the account-owned containers.
//
// See details in [client.Client.ContainerList].
func (p *Pool) ContainerList(ctx context.Context, ownerID user.ID, prm client.PrmContainerList) ([]cid.ID, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return []cid.ID{}, err
	}

	start := time.Now()
	ids, err := c.ContainerList(ctx, ownerID, prm)
	statUpdater.incRequests(time.Since(start), methodContainerList)
	statUpdater.updateErrorRate(err)

	return ids, err
}

// ContainerDelete sends request to remove the NeoFS container.
//
// See details in [client.Client.ContainerDelete].
func (p *Pool) ContainerDelete(ctx context.Context, id cid.ID, prm client.PrmContainerDelete) error {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return err
	}

	start := time.Now()
	err = c.ContainerDelete(ctx, id, prm)
	statUpdater.incRequests(time.Since(start), methodContainerDelete)
	statUpdater.updateErrorRate(err)

	return err
}

// ContainerEACL reads eACL table of the NeoFS container.
//
// See details in [client.Client.ContainerEACL].
func (p *Pool) ContainerEACL(ctx context.Context, id cid.ID, prm client.PrmContainerEACL) (eacl.Table, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return eacl.Table{}, err
	}

	start := time.Now()
	table, err := c.ContainerEACL(ctx, id, prm)
	statUpdater.incRequests(time.Since(start), methodContainerEACL)
	statUpdater.updateErrorRate(err)

	return table, err
}

// ContainerSetEACL sends request to update eACL table of the NeoFS container.
//
// See details in [client.Client.ContainerSetEACL].
func (p *Pool) ContainerSetEACL(ctx context.Context, table eacl.Table, prm client.PrmContainerSetEACL) error {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return err
	}

	start := time.Now()
	err = c.ContainerSetEACL(ctx, table, prm)
	statUpdater.incRequests(time.Since(start), methodContainerSetEACL)
	statUpdater.updateErrorRate(err)

	return err
}
