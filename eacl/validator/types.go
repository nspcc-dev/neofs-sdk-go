package validator

import (
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
)

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
	HeadersOfType(eacl.FilterHeaderType) ([]Header, bool)
}

// ValidationUnit represents unit of check for Validator.
type ValidationUnit struct {
	cid *cid.ID

	role eacl.Role

	op eacl.Operation

	hdrSrc TypedHeaderSource

	key []byte

	table *eacl.Table
}

// WithContainerID configures ValidationUnit to use v as request's container ID.
func (u *ValidationUnit) WithContainerID(v *cid.ID) *ValidationUnit {
	if u != nil {
		u.cid = v
	}

	return u
}

// WithRole configures ValidationUnit to use v as request's role.
func (u *ValidationUnit) WithRole(v eacl.Role) *ValidationUnit {
	if u != nil {
		u.role = v
	}

	return u
}

// WithOperation configures ValidationUnit to use v as request's operation.
func (u *ValidationUnit) WithOperation(v eacl.Operation) *ValidationUnit {
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

// WithEACLTable configures ValidationUnit to use v as request's eACL table.
func (u *ValidationUnit) WithEACLTable(v *eacl.Table) *ValidationUnit {
	if u != nil {
		u.table = v
	}

	return u
}
