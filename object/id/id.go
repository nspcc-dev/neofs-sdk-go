package oid

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	signatureV2 "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

// ID represents NeoFS object identifier.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.ObjectID
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = ObjectID(refs.ObjectID{}) // not recommended
type ID refs.ObjectID

var errInvalidIDString = errors.New("incorrect format of the string object ID")

// ReadFromV2 reads ID from the refs.ObjectID message.
//
// See also WriteToV2.
func (id *ID) ReadFromV2(m refs.ObjectID) {
	*id = ID(m)
}

// WriteToV2 writes ID to the refs.ObjectID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (id ID) WriteToV2(m *refs.ObjectID) {
	*m = (refs.ObjectID)(id)
}

// SetSHA256 sets object identifier value to SHA256 checksum.
func (id *ID) SetSHA256(v [sha256.Size]byte) {
	(*refs.ObjectID)(id).SetValue(v[:])
}

// Equals returns true if identifiers are identical.
func (id ID) Equals(id2 *ID) bool {
	v2 := (refs.ObjectID)(id)
	return bytes.Equal(v2.GetValue(), (*refs.ObjectID)(id2).GetValue())
}

// Parse is a reverse action to String().
func (id *ID) Parse(s string) error {
	data, err := base58.Decode(s)
	if err != nil {
		return fmt.Errorf("could not parse object.ID from string: %w", err)
	} else if len(data) != sha256.Size {
		return errInvalidIDString
	}

	(*refs.ObjectID)(id).SetValue(data)

	return nil
}

// String implements fmt.Stringer interface method.
func (id ID) String() string {
	v2 := (refs.ObjectID)(id)
	return base58.Encode(v2.GetValue())
}

// CalculateIDSignature signs object id with provided key.
func (id ID) CalculateIDSignature(key *ecdsa.PrivateKey) (*signature.Signature, error) {
	var idV2 refs.ObjectID
	id.WriteToV2(&idV2)

	sign, err := sigutil.SignData(key,
		signatureV2.StableMarshalerWrapper{
			SM: &idV2,
		},
	)

	return sign, err
}

// Marshal marshals ID into a protobuf binary form.
func (id ID) Marshal() ([]byte, error) {
	v2 := (refs.ObjectID)(id)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of ID.
func (id *ID) Unmarshal(data []byte) error {
	return (*refs.ObjectID)(id).Unmarshal(data)
}

// MarshalJSON encodes ID to protobuf JSON format.
func (id ID) MarshalJSON() ([]byte, error) {
	v2 := (refs.ObjectID)(id)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes ID from protobuf JSON format.
func (id *ID) UnmarshalJSON(data []byte) error {
	return (*refs.ObjectID)(id).UnmarshalJSON(data)
}
