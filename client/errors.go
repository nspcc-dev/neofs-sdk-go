package client

import (
	"errors"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
)

// unwraps err using errors.Unwrap and returns the result.
func unwrapErr(err error) error {
	for e := errors.Unwrap(err); e != nil; e = errors.Unwrap(err) {
		err = e
	}

	return err
}

// IsErrContainerNotFound checks if err corresponds to NeoFS status
// return corresponding to missing container. Supports wrapped errors.
func IsErrContainerNotFound(err error) bool {
	switch unwrapErr(err).(type) {
	default:
		return false
	case
		apistatus.ContainerNotFound,
		*apistatus.ContainerNotFound:
		return true
	}
}

// IsErrObjectNotFound checks if err corresponds to NeoFS status
// return corresponding to missing object. Supports wrapped errors.
func IsErrObjectNotFound(err error) bool {
	switch unwrapErr(err).(type) {
	default:
		return false
	case
		apistatus.ObjectNotFound,
		*apistatus.ObjectNotFound:
		return true
	}
}

// IsErrObjectAlreadyRemoved checks if err corresponds to NeoFS status
// return corresponding to already removed object. Supports wrapped errors.
func IsErrObjectAlreadyRemoved(err error) bool {
	switch unwrapErr(err).(type) {
	default:
		return false
	case
		apistatus.ObjectAlreadyRemoved,
		*apistatus.ObjectAlreadyRemoved:
		return true
	}
}

// IsErrSessionExpired checks if err corresponds to NeoFS status return
// corresponding to expired session. Supports wrapped errors.
func IsErrSessionExpired(err error) bool {
	switch unwrapErr(err).(type) {
	default:
		return false
	case
		apistatus.SessionTokenExpired,
		*apistatus.SessionTokenExpired:
		return true
	}
}

// IsErrSessionNotFound checks if err corresponds to NeoFS status return
// corresponding to missing session. Supports wrapped errors.
func IsErrSessionNotFound(err error) bool {
	switch unwrapErr(err).(type) {
	default:
		return false
	case
		apistatus.SessionTokenNotFound,
		*apistatus.SessionTokenNotFound:
		return true
	}
}
