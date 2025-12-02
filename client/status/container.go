package apistatus

import (
	"errors"

	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

var (
	// ErrEACLNotFound is an instance of EACLNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrEACLNotFound EACLNotFound
	// ErrContainerNotFound is an instance of ContainerNotFound error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrContainerNotFound ContainerNotFound
	// ErrContainerLocked is an instance of ContainerLocked error status. It's expected to be used for [errors.Is]
	// and MUST NOT be changed.
	ErrContainerLocked ContainerLocked
)

// ContainerNotFound describes status of the failure because of the missing container.
type ContainerNotFound struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultContainerNotFoundMsg = "container not found"

func (x ContainerNotFound) Error() string {
	if x.msg == "" {
		x.msg = defaultContainerNotFoundMsg
	}

	return errMessageStatus(protostatus.ContainerNotFound, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *ContainerNotFound) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ContainerNotFound) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultContainerNotFoundMsg
	}
	return &protostatus.Status{Code: protostatus.ContainerNotFound, Message: x.msg, Details: x.dts}
}

// EACLNotFound describes status of the failure because of the missing eACL
// table.
type EACLNotFound struct {
	msg string
	dts []*protostatus.Status_Detail
}

const defaultEACLNotFoundMsg = "eACL not found"

func (x EACLNotFound) Error() string {
	if x.msg == "" {
		x.msg = defaultEACLNotFoundMsg
	}

	return errMessageStatus(protostatus.EACLNotFound, x.msg)
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

// implements local interface defined in [ToError] func.
func (x *EACLNotFound) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x EACLNotFound) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultEACLNotFoundMsg
	}
	return &protostatus.Status{Code: protostatus.EACLNotFound, Message: x.msg, Details: x.dts}
}

// ContainerLocked describes status of the failure because of the locked container.
type ContainerLocked struct {
	msg string
	dts []*protostatus.Status_Detail
}

// NewContainerLocked constructs ContainerLocked with given message.
func NewContainerLocked(msg string) ContainerLocked {
	return ContainerLocked{msg: msg}
}

const defaultContainerLockedMsg = "container is locked"

// Error implements built-in [error] interface.
func (x ContainerLocked) Error() string {
	if x.msg == "" {
		x.msg = defaultContainerLockedMsg
	}

	return errMessageStatus(protostatus.ContainerLocked, x.msg)
}

// Is implements interface for correct checking current error type with [errors.Is].
func (x ContainerLocked) Is(target error) bool {
	switch target.(type) {
	default:
		return errors.Is(Error, target)
	case ContainerLocked, *ContainerLocked:
		return true
	}
}

// implements local interface defined in [ToError] func.
func (x *ContainerLocked) fromProtoMessage(st *protostatus.Status) {
	x.msg = st.Message
	x.dts = st.Details
}

// implements local interface defined in [FromError] func.
func (x ContainerLocked) protoMessage() *protostatus.Status {
	if x.msg == "" {
		x.msg = defaultContainerLockedMsg
	}
	return &protostatus.Status{Code: protostatus.ContainerLocked, Message: x.msg, Details: x.dts}
}
