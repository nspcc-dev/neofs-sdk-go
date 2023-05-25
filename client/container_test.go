package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestClient_Container(t *testing.T) {
	c := newClient(t, nil, nil)
	ctx := context.Background()

	t.Run("missing signer", func(t *testing.T) {
		tt := []struct {
			name       string
			methodCall func() error
		}{
			{
				"put",
				func() error {
					_, err := c.ContainerPut(ctx, container.Container{}, PrmContainerPut{})
					return err
				},
			},
			{
				"get",
				func() error {
					_, err := c.ContainerGet(ctx, cid.ID{}, PrmContainerGet{})
					return err
				},
			},
			{
				"list",
				func() error {
					_, err := c.ContainerList(ctx, user.ID{}, PrmContainerList{})
					return err
				},
			},
			{
				"delete",
				func() error {
					return c.ContainerDelete(ctx, cid.ID{}, PrmContainerDelete{})
				},
			},
			{
				"eacl",
				func() error {
					_, err := c.ContainerEACL(ctx, PrmContainerEACL{idSet: true})
					return err
				},
			},
			{
				"set_eacl",
				func() error {
					return c.ContainerSetEACL(ctx, PrmContainerSetEACL{tableSet: true})
				},
			},
			{
				"announce_space",
				func() error {
					return c.ContainerAnnounceUsedSpace(ctx, PrmAnnounceSpace{announcements: make([]container.SizeEstimation, 1)})
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
