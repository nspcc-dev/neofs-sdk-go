package poolv2

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client/clientutil"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

type IDHandler func(id oid.ID)

func (p *Pool) CreateObject(ctx context.Context, builder *CreateObjectBuilder) error {
	params := clientutil.NewCreateObjectPrm(builder.containerID, builder.payload).
		SetSigner(p.signer).
		SetIDHandler(builder.idHandler)

	for _, a := range builder.attr {
		params.AddAttribute(a.Name, a.Value)
	}

	if err := clientutil.CreateObjectWithClient(ctx, p.client(), *params); err != nil {
		return fmt.Errorf("create object: %w", err)
	}

	return nil
}

// CreateObject2 . We may even pass not CreateObjectBuilder, but clientutil.CreateObjectPrm builder.
func (p *Pool) CreateObject2(ctx context.Context, builder *clientutil.CreateObjectPrm) error {
	// in this case pool only manage clients, but we use the same interface.
	// it gives opportunity to use single client, if someone wants.
	if err := clientutil.CreateObjectWithClient(ctx, p.client(), *builder); err != nil {
		return fmt.Errorf("create object: %w", err)
	}

	return nil
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
