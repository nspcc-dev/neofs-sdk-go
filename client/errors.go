package client

import apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"

// IsErrContainerNotFound checks if err corresponds to NeoFS status
// return corresponding to missing container.
func IsErrContainerNotFound(err error) bool {
	switch err.(type) {
	default:
		return false
	case
		apistatus.ContainerNotFound,
		*apistatus.ContainerNotFound:
		return true
	}
}

// IsErrObjectNotFound checks if err corresponds to NeoFS status
// return corresponding to missing object.
func IsErrObjectNotFound(err error) bool {
	switch err.(type) {
	default:
		return false
	case
		apistatus.ObjectNotFound,
		*apistatus.ObjectNotFound:
		return true
	}
}

// IsErrObjectAlreadyRemoved checks if err corresponds to NeoFS status
// return corresponding to already removed object.
func IsErrObjectAlreadyRemoved(err error) bool {
	switch err.(type) {
	default:
		return false
	case
		apistatus.ObjectAlreadyRemoved,
		*apistatus.ObjectAlreadyRemoved:
		return true
	}
}
