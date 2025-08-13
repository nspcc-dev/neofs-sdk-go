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

// ContainerPutExecutor represents requirements to async container put operation.
// See documentation for functions in [client.Client]. The same semantics is expected.
type ContainerPutExecutor interface {
	ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm client.PrmContainerPut) (cid.ID, error)
	ContainerGet(ctx context.Context, id cid.ID, prm client.PrmContainerGet) (container.Container, error)
}

// ContainerPutWaiter implements sync logic to container put operation.
type ContainerPutWaiter struct {
	executor     ContainerPutExecutor
	pollInterval time.Duration
}

// NewContainerPutWaiter is a constructor for ContainerPutWaiter.
func NewContainerPutWaiter(c ContainerPutExecutor, pollInterval time.Duration) ContainerPutWaiter {
	return ContainerPutWaiter{executor: c, pollInterval: pollInterval}
}

// SetPollInterval allows rewrite default poll interval.
func (w *ContainerPutWaiter) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// ContainerPut sends request to save container in NeoFS.
//
// ContainerPut uses ContainerPutExecutor to create container and check the container is created.
func (w ContainerPutWaiter) ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm client.PrmContainerPut) (cid.ID, error) {
	id, err := w.executor.ContainerPut(ctx, cont, signer, prm)
	if err != nil {
		return cid.ID{}, fmt.Errorf("put: %w", err)
	}

	var prmGet client.PrmContainerGet

	logic := func() error {
		contaier, err := w.executor.ContainerGet(ctx, id, prmGet)
		if err != nil {
			if errors.Is(err, apistatus.ErrContainerNotFound) {
				return errRetry
			}

			return fmt.Errorf("ContainerGet: %w", err)
		}

		fmt.Println("ContainerPutWaiter", "id", id.String(), "owner", contaier.Owner().String())

		return nil
	}

	return id, poll(ctx, w.pollInterval, logic)
}
