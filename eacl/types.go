package eacl

import (
	"errors"

	bearer "github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// Source is the interface that wraps
// basic methods of extended ACL table source.
type Source interface {
	// GetEACL reads the table from the source by identifier.
	// It returns any error encountered.
	//
	// GetEACL must return exactly one non-nil value.
	//
	// Must return pkg/core/container.ErrEACLNotFound if requested
	// eACL table is not in source.
	GetEACL(*cid.ID) (*Table, error)
}

// Header is an interface of string key-value header.
type Header interface {
	Key() string
	Value() string
}

// TypedHeaderSource is the interface that wraps
// method for selecting typed headers by type.
type TypedHeaderSource interface {
	// HeadersOfType returns the list of key-value headers
	// of particular type.
	//
	// It returns any problem encountered through the boolean
	// false value.
	HeadersOfType(FilterHeaderType) ([]Header, bool)
}

// ValidationUnit represents unit of check for Validator.
type ValidationUnit struct {
	cid *cid.ID

	role Role

	op Operation

	hdrSrc TypedHeaderSource

	key []byte

	bearer *bearer.BearerToken
}

// ErrEACLNotFound is returned by eACL storage implementations when
// requested eACL table is not in storage.
var ErrEACLNotFound = errors.New("extended ACL table is not set for this container")

// WithContainerID configures ValidationUnit to use v as request's container ID.
func (u *ValidationUnit) WithContainerID(v *cid.ID) *ValidationUnit {
	if u != nil {
		u.cid = v
	}

	return u
}

// WithRole configures ValidationUnit to use v as request's role.
func (u *ValidationUnit) WithRole(v Role) *ValidationUnit {
	if u != nil {
		u.role = v
	}

	return u
}

// WithOperation configures ValidationUnit to use v as request's operation.
func (u *ValidationUnit) WithOperation(v Operation) *ValidationUnit {
	if u != nil {
		u.op = v
	}

	return u
}

// WithHeaderSource configures ValidationUnit to use v as a source of headers.
func (u *ValidationUnit) WithHeaderSource(v TypedHeaderSource) *ValidationUnit {
	if u != nil {
		u.hdrSrc = v
	}

	return u
}

// WithSenderKey configures ValidationUnit to use as sender's public key.
func (u *ValidationUnit) WithSenderKey(v []byte) *ValidationUnit {
	if u != nil {
		u.key = v
	}

	return u
}

// WithBearerToken configures ValidationUnit to use v as request's bearer token.
func (u *ValidationUnit) WithBearerToken(bearer *bearer.BearerToken) *ValidationUnit {
	if u != nil {
		u.bearer = bearer
	}

	return u
}
