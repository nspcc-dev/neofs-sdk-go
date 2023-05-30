package poolv2

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client/clientutil"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

type IDHandler func(i oid.ID)

func (p *Pool) CreateObject(ctx context.Context, builder *CreateObjectBuilder) (oid.ID, error) {
	var params clientutil.CreateObjectPrm
	params.SetSigner(p.signer)
	params.SetPayload(builder.payload)
	for _, a := range builder.attr {
		params.AddAttribute(a.Name, a.Value)
	}

	var id oid.ID
	params.SetIDHandler(builder.idHandler)
	params.Container = builder.containerID

	if err := clientutil.CreateObjectWithClient(ctx, p.client(), params); err != nil {
		return id, fmt.Errorf("create object: %w", err)
	}

	return id, nil
}

func (p *Pool) ReadObject(ctx context.Context, builder *ReadObjectBuilder) error {
	var params clientutil.ReadObjectPrm
	params.Object = builder.objectID
	params.Container = builder.containerID
	params.WritePayloadTo(builder.writer)

	if err := clientutil.ReadObjectWithClient(ctx, p.client(), params); err != nil {
		return fmt.Errorf("read object: %w", err)
	}

	return nil
}
