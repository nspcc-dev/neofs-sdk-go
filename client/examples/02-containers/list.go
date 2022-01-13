package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

func List(ctx context.Context, cli *client.Client, key *ecdsa.PrivateKey) ([]*cid.ID, error) {
	// ListContainers method requires Owner ID.
	// OwnerID is a binary representation of wallet address.

	response, err := cli.ListContainers(ctx, ownerIDFromPrivateKey(key))
	if err != nil {
		return nil, fmt.Errorf("can't list containers: %w", err)
	}

	return response.IDList(), nil
}

func ownerIDFromPrivateKey(key *ecdsa.PrivateKey) *owner.ID {
	w, err := owner.NEO3WalletFromPublicKey(&key.PublicKey)
	if err != nil {
		panic(fmt.Errorf("invalid private key"))
	}

	return owner.NewIDFromNeo3Wallet(w)
}
