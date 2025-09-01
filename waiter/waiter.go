package waiter

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	// DefaultPollInterval is a duration between operation checks. Waiters accept any [time.Duration] for the interval,
	// but if omitted (set to 0) this one will be used.
	DefaultPollInterval = 1 * time.Second
)

var (
	// ErrConfirmationTimeout describes situation, when we sent request, but got timeout during waiting confirmation.
	// For instance: we send containerPut, after that we wait and do containerGet, to be sure it was created successfully.
	// If we get context timeout at this step, this error will be returner.
	// Notice that in many cases this doesn't mean the operation has completely failed (because the request for it was
	// sent without any errors).
	ErrConfirmationTimeout = errors.New("confirmation timeout")

	// errRetry is a special error for using with pollLogic. It tells to some waiter to wait one more tick.
	errRetry = errors.New("retry")
)

// Executor describes requirements for async.
type Executor interface {
	ContainerDeleteExecutor
	ContainerSetEACLExecutor
	ContainerPutExecutor
}

// Waiter combines async [client.Client]/[pool.Pool] methods and gives sync alternative of them with the same func signatures.
type Waiter struct {
	ContainerPutWaiter
	ContainerSetEACLWaiter
	ContainerDeleteWaiter
}

// The function implements poll logic for each waiter.
//
// The return value means:
//   - nil is a total success.
//   - errRetry means waiter should wait one more tick.
//   - another error means a fatal problem.
type pollLogic func() error

// NewWaiter is a constructor for [waiter.Waiter].
//
// Each pollInterval waiter make the request to confirm the operation it waits, was successfully executed.
//   - For instance: ContainerPutWaiter waits until container will be created.
func NewWaiter(executor Executor, pollInterval time.Duration) *Waiter {
	w := &Waiter{
		ContainerPutWaiter:     NewContainerPutWaiter(executor, pollInterval),
		ContainerSetEACLWaiter: NewContainerSetEACLWaiter(executor, pollInterval),
		ContainerDeleteWaiter:  NewContainerDeleteWaiter(executor, pollInterval),
	}

	return w
}

func poll(ctx context.Context, pollInterval time.Duration, callBack pollLogic) error {
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}

	t := time.NewTicker(pollInterval)
	defer func() {
		t.Stop()
	}()

	for {
		select {
		case <-t.C:
			if err := callBack(); err != nil {
				if errors.Is(err, errRetry) {
					// wait one more tick
					continue
				}

				return fmt.Errorf("poller: %w", err)
			}

			return nil
		case <-ctx.Done():
			return ErrConfirmationTimeout
		}
	}
}
