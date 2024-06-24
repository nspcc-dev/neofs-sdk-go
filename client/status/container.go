package apistatus

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

// Container error instances which may be used to check API errors against using
// [errors.Is]. All of them MUST NOT be changed.
var (
	ErrEACLNotFound      EACLNotFound
	ErrContainerNotFound ContainerNotFound
)

// ContainerNotFound describes status of the failure because of the missing
// container.
type ContainerNotFound string

// NewContainerNotFoundError constructs missing container error with specified
// cause.
func NewContainerNotFoundError(cause error) ContainerNotFound {
	return ContainerNotFound(cause.Error())
}

// Error implements built-in error interface.
func (x ContainerNotFound) Error() string {
	const desc = "container not found"
	if x != "" {
		return fmt.Sprintf(errFmt, status.ContainerNotFound, desc, string(x))
	}
	return fmt.Sprintf(errFmtNoMessage, status.ContainerNotFound, desc)
}

// Is checks whether target is of type ContainerNotFound, *ContainerNotFound or
// [Error]. Is implements interface consumed by [errors.Is].
func (x ContainerNotFound) Is(target error) bool { return errorIs(x, target) }

func (x *ContainerNotFound) readFromV2(m *status.Status) error {
	if m.Code != status.ContainerNotFound {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.ContainerNotFound))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	*x = ContainerNotFound(m.Message)
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x ContainerNotFound) ErrorToV2() *status.Status {
	return &status.Status{Code: status.ContainerNotFound, Message: string(x)}
}

// EACLNotFound describes status of the failure because of the missing eACL.
type EACLNotFound string

// NewEACLNotFoundError constructs missing eACL error with specified cause.
func NewEACLNotFoundError(cause error) EACLNotFound {
	return EACLNotFound(cause.Error())
}

// Error implements built-in error interface.
func (x EACLNotFound) Error() string {
	const desc = "eACL not found"
	if x != "" {
		return fmt.Sprintf(errFmt, status.EACLNotFound, desc, string(x))
	}
	return fmt.Sprintf(errFmtNoMessage, status.EACLNotFound, desc)
}

// Is checks whether target is of type EACLNotFound, *EACLNotFound or [Error].
// Is implements interface consumed by [errors.Is].
func (x EACLNotFound) Is(target error) bool { return errorIs(x, target) }

func (x *EACLNotFound) readFromV2(m *status.Status) error {
	if m.Code != status.EACLNotFound {
		panic(fmt.Sprintf("unexpected code %d instead of %d", m.Code, status.EACLNotFound))
	}
	if len(m.Details) > 0 {
		return errors.New("details attached but not supported")
	}
	*x = EACLNotFound(m.Message)
	return nil
}

// ErrorToV2 implements [StatusV2] interface method.
func (x EACLNotFound) ErrorToV2() *status.Status {
	return &status.Status{Code: status.EACLNotFound, Message: string(x)}
}
