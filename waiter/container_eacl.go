package waiter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ContainerSetEACLExecutor represents requirements to async container setEACL operation.
// See documentation for functions in [client.Client]. The same semantics is expected.
type ContainerSetEACLExecutor interface {
	ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm client.PrmContainerSetEACL) error
	ContainerEACL(ctx context.Context, id cid.ID, prm client.PrmContainerEACL) (eacl.Table, error)
}

// ContainerSetEACLWaiter implements sync logic to container setEACL operation.
type ContainerSetEACLWaiter struct {
	executor     ContainerSetEACLExecutor
	pollInterval time.Duration
}

// NewContainerSetEACLWaiter is a constructor for NewContainerSetEACLWaiter.
func NewContainerSetEACLWaiter(c ContainerSetEACLExecutor, pollInterval time.Duration) ContainerSetEACLWaiter {
	return ContainerSetEACLWaiter{executor: c, pollInterval: pollInterval}
}

// SetPollInterval allows rewrite default poll interval.
func (w *ContainerSetEACLWaiter) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// ContainerSetEACL sends request to update eACL table of the NeoFS container.
//
// ContainerSetEACL uses ContainerSetEACLExecutor to setEacl and check the eacl is set.
func (w ContainerSetEACLWaiter) ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm client.PrmContainerSetEACL) error {
	if err := w.executor.ContainerSetEACL(ctx, table, signer, prm); err != nil {
		return fmt.Errorf("container setEacl: %w", err)
	}

	contID, ok := table.CID()
	if !ok {
		return client.ErrMissingEACLContainer
	}

	newBinary, err := table.Marshal()
	if err != nil {
		return fmt.Errorf("newTable.Marshal: %w", err)
	}

	var prmEacl client.PrmContainerEACL

	logic := func() error {
		actualTable, err := w.executor.ContainerEACL(ctx, contID, prmEacl)
		if err != nil {
			if errors.Is(err, apistatus.ErrEACLNotFound) {
				return errRetry
			}

			return fmt.Errorf("ContainerEACL: %w", err)
		}

		actualBinary, err := actualTable.Marshal()
		if err != nil {
			return fmt.Errorf("table.Marshal: %w", err)
		}

		if bytes.Equal(newBinary, actualBinary) {
			return nil
		}

		return errRetry
	}

	return poll(ctx, w.pollInterval, logic)
}
