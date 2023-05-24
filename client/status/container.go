package apistatus

import (
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

var (
	// ErrEACLNotFound is an instance of EACLNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrEACLNotFound EACLNotFound
	// ErrContainerNotFound is an instance of ContainerNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrContainerNotFound ContainerNotFound
)

// ContainerNotFound describes status of the failure because of the missing container.
// Instances provide [StatusV2] and error interfaces.
type ContainerNotFound struct {
	v2 status.Status
}

const defaultContainerNotFoundMsg = "container not found"

func (x ContainerNotFound) Error() string {
	msg := x.v2.Message()
	if msg == "" {
		msg = defaultContainerNotFoundMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(container.StatusNotFound, container.GlobalizeFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ContainerNotFound) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ContainerNotFound, *ContainerNotFound:
		return true
	}
}

// implements local interface defined in [ErrorFromV2] func.
func (x *ContainerNotFound) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ErrorToV2 implements [StatusV2] interface method.
// If the value was returned by [ErrorFromV2], returns the source message.
// Otherwise, returns message with
//   - code: CONTAINER_NOT_FOUND;
//   - string message: "container not found";
//   - details: empty.
func (x ContainerNotFound) ErrorToV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(container.StatusNotFound, container.GlobalizeFail))
	x.v2.SetMessage(defaultContainerNotFoundMsg)
	return &x.v2
}

// EACLNotFound describes status of the failure because of the missing eACL
// table.
// Instances provide [StatusV2] and error interfaces.
type EACLNotFound struct {
	v2 status.Status
}

const defaultEACLNotFoundMsg = "eACL not found"

func (x EACLNotFound) Error() string {
	msg := x.v2.Message()
	if msg == "" {
		msg = defaultEACLNotFoundMsg
	}

	return errMessageStatusV2(
		globalizeCodeV2(container.StatusEACLNotFound, container.GlobalizeFail),
		msg,
	)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x EACLNotFound) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case EACLNotFound, *EACLNotFound:
		return true
	}
}

// implements local interface defined in [ErrorFromV2] func.
func (x *EACLNotFound) fromStatusV2(st *status.Status) {
	x.v2 = *st
}

// ErrorToV2 implements [StatusV2] interface method.
// If the value was returned by [ErrorFromV2], returns the source message.
// Otherwise, returns message with
//   - code: EACL_NOT_FOUND;
//   - string message: "eACL not found";
//   - details: empty.
func (x EACLNotFound) ErrorToV2() *status.Status {
	x.v2.SetCode(globalizeCodeV2(container.StatusEACLNotFound, container.GlobalizeFail))
	x.v2.SetMessage(defaultEACLNotFoundMsg)
	return &x.v2
}
