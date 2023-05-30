package poolv2

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func (p *Pool) CreateContainer(ctx context.Context, builder *CreateContainerBuilder) (cid.ID, error) {
	cont := container.NewContainer(builder.basicACL, builder.owner)
	for _, a := range builder.attr {
		cont.SetAttribute(a.Name, a.Value)
	}

	var id cid.ID

	pp, err := builder.placementPolicy()
	if err != nil {
		return id, fmt.Errorf("placement policy: %w", err)
	}

	cont.SetPlacementPolicy(pp)

	var params client.PrmContainerPut
	params.SetSigner(p.signer)

	c := p.client()

	id, err = c.ContainerPut(ctx, cont, params)
	if err != nil {
		return id, fmt.Errorf("create container: %w", err)
	}

	return id, nil
}
