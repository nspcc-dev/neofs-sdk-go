package neofs

import (
	"context"
	"crypto/ecdsa"

	"github.com/nspcc-dev/neofs-api-go/pkg/client"
	"github.com/nspcc-dev/neofs-api-go/pkg/owner"
	"github.com/nspcc-dev/neofs-api-go/pkg/token"
	"github.com/nspcc-dev/neofs-sdk-go/pkg/pool"
)

// ClientPlant provides connections to NeoFS nodes from pool and allows to
// get local owner ID.
type ClientPlant interface {
	ConnectionArtifacts() (client.Client, *token.SessionToken, error)
	OwnerID() *owner.ID
}

type neofsClientPlant struct {
	key     *ecdsa.PrivateKey
	ownerID *owner.ID
	pool    pool.Pool
}

// ConnectionArtifacts returns connection from pool.
func (cp *neofsClientPlant) ConnectionArtifacts() (client.Client, *token.SessionToken, error) {
	return cp.pool.ConnectionArtifacts()
}

// OwnerID returns plant's owner ID.
func (cp *neofsClientPlant) OwnerID() *owner.ID {
	return cp.ownerID
}

// NewClientPlant creates new ClientPlant from given context, pool and credentials.
func NewClientPlant(ctx context.Context, pool pool.Pool, creds Credentials) (ClientPlant, error) {
	return &neofsClientPlant{key: creds.PrivateKey(), ownerID: creds.Owner(), pool: pool}, nil
}
