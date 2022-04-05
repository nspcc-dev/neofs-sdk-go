package cid

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// ID represents NeoFS container identifier.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.ContainerID
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = ID(refs.ContainerID) // not recommended
type ID refs.ContainerID

// ReadFromV2 reads ID from the refs.ContainerID message.
//
// See also WriteToV2.
func (id *ID) ReadFromV2(m refs.ContainerID) {
	*id = ID(m)
}

// WriteToV2 writes ID to the refs.ContainerID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (id ID) WriteToV2(m *refs.ContainerID) {
	*m = (refs.ContainerID)(id)
}

// SetSHA256 sets container identifier value to SHA256 checksum of container body.
func (id *ID) SetSHA256(v [sha256.Size]byte) {
	(*refs.ContainerID)(id).SetValue(v[:])
}

// Equals returns true if identifiers are identical.
func (id ID) Equals(id2 ID) bool {
	idv2 := (refs.ContainerID)(id)
	id2v2 := (refs.ContainerID)(id2)
	return bytes.Equal(
		idv2.GetValue(),
		id2v2.GetValue(),
	)
}

// Parse is a reverse action to String().
func (id *ID) Parse(s string) error {
	data, err := base58.Decode(s)
	if err != nil {
		return err
	} else if len(data) != sha256.Size {
		return errors.New("incorrect format of the string container ID")
	}

	(*refs.ContainerID)(id).SetValue(data)

	return nil
}

// String implements fmt.Stringer interface method.
func (id ID) String() string {
	v2 := (refs.ContainerID)(id)
	return base58.Encode(v2.GetValue())
}

// Empty returns true if it is called on
// zero container ID.
func (id ID) Empty() bool {
	v2 := (refs.ContainerID)(id)
	return v2.GetValue() == nil
}
