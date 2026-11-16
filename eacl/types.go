package eacl

import (
	"errors"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Header is an interface of string key-value header.
type Header interface {
	Key() string
	Value() string
}

// ErrHeaderUnavailable indicates that headers of the requested type cannot be
// obtained. The validator reports a non-final action in this case.
var ErrHeaderUnavailable = errors.New("header unavailable")

// TypedHeaderSource provides a single value for each header type and key.
type TypedHeaderSource interface {
	// HeaderValue returns nil if the key is absent. A non-nil empty slice is a
	// present empty value. If headers of the requested type cannot be obtained,
	// it returns [ErrHeaderUnavailable] instead of reporting an absent key.
	// Returned bytes must remain valid until the next call to HeaderValue.
	HeaderValue(FilterHeaderType, string) ([]byte, error)
}

// ValidationUnit represents unit of check for Validator.
type ValidationUnit struct {
	cid *cid.ID

	role Role

	op Operation

	hdrSrc TypedHeaderSource

	key []byte

	account []byte

	table *Table
}

// WithContainerID configures ValidationUnit to use v as request's container ID.
// ID value must not be zero.
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
//
// Parameter v is a serialized compressed public key. See [elliptic.MarshalCompressed].
func (u *ValidationUnit) WithSenderKey(v []byte) *ValidationUnit {
	if u != nil {
		u.key = v
	}

	return u
}

// WithEACLTable configures ValidationUnit to use v as request's bearer token.
func (u *ValidationUnit) WithEACLTable(table *Table) *ValidationUnit {
	if u != nil {
		u.table = table
	}

	return u
}

// WithAccount configures ValidationUnit to use as sender's [user.ID].
func (u *ValidationUnit) WithAccount(v user.ID) *ValidationUnit {
	if u != nil {
		u.account = v[:]
	}

	return u
}
