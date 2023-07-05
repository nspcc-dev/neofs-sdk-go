package waiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// ContainerDeleteExecutor represents requirements to async container delete operation.
// See documentation for functions in [client.Client]. The same semantics is expected.
type ContainerDeleteExecutor interface {
	ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm client.PrmContainerDelete) error
	ContainerGet(ctx context.Context, id cid.ID, prm client.PrmContainerGet) (container.Container, error)
}

// ContainerDeleteWaiter implements sync logic to container delete operation.
type ContainerDeleteWaiter struct {
	executor     ContainerDeleteExecutor
	pollInterval time.Duration
}

// NewContainerDeleteWaiter is a constructor for ContainerDeleteWaiter.
func NewContainerDeleteWaiter(executor ContainerDeleteExecutor, pollInterval time.Duration) ContainerDeleteWaiter {
	return ContainerDeleteWaiter{executor: executor, pollInterval: pollInterval}
}

// SetPollInterval allows rewrite default poll interval.
func (w *ContainerDeleteWaiter) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// ContainerDelete sends request to remove the NeoFS container.
//
// ContainerDelete uses ContainerDeleteExecutor to delete and check the container is deleted.
func (w ContainerDeleteWaiter) ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm client.PrmContainerDelete) error {
	if err := w.executor.ContainerDelete(ctx, id, signer, prm); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	var prmGet client.PrmContainerGet

	logic := func() error {
		_, err := w.executor.ContainerGet(ctx, id, prmGet)
		if err != nil {
			if errors.Is(err, apistatus.ErrContainerNotFound) {
				return nil
			}

			return fmt.Errorf("ContainerGet: %w", err)
		}

		return errRetry
	}

	return poll(ctx, w.pollInterval, logic)
}
