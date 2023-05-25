package client

import (
	"context"
	"testing"

	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestClient_Get(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t, nil, nil)
		ctx := context.Background()

		var nonilAddr v2refs.Address
		nonilAddr.SetObjectID(new(v2refs.ObjectID))
		nonilAddr.SetContainerID(new(v2refs.ContainerID))

		tt := []struct {
			name       string
			methodCall func() error
		}{
			{
				"get",
				func() error {
					_, err := c.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, PrmObjectGet{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_range",
				func() error {
					_, err := c.ObjectRangeInit(ctx, cid.ID{}, oid.ID{}, 0, 1, PrmObjectRange{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_head",
				func() error {
					_, err := c.ObjectHead(ctx, cid.ID{}, oid.ID{}, PrmObjectHead{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
		}

		for _, test := range tt {
			t.Run(test.name, func(t *testing.T) {
				require.ErrorIs(t, test.methodCall(), ErrMissingSigner)
			})
		}
	})
}
