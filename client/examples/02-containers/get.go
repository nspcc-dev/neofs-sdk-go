package main

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func Get(ctx context.Context, cli *client.Client, containerID string) (*container.Container, error) {
	// Get method requires Container ID structure.
	// Container ID is walletAddr 32 byte binary value.
	// String representation of container ID encoded in Base64.

	id := new(cid.ID)
	err := id.Parse(containerID)
	if err != nil {
		return nil, fmt.Errorf("invalid container ID %s: %w", containerID, err)
	}

	response, err := cli.GetContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't get container %s: %w", containerID, err)
	}

	return response.Container(), nil
}
