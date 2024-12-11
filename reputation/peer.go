package reputation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	protoreputation "github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
)

// PeerID represents unique identifier of the peer participating in the NeoFS
// reputation system.
//
// ID is mutually compatible with [protoreputation.PeerID] message. See
// [PeerID.FromProtoMessage] / [PeerID.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type PeerID struct {
	key []byte
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [PeerID.ProtoMessage].
func (x *PeerID) FromProtoMessage(m *protoreputation.PeerID) error {
	if x.key = m.PublicKey; len(m.PublicKey) == 0 {
		return errors.New("missing ID bytes")
	}
	return nil
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [PeerID.FromProtoMessage].
func (x PeerID) ProtoMessage() *protoreputation.PeerID {
	return &protoreputation.PeerID{PublicKey: x.key}
}

// SetPublicKey sets [PeerID] as a binary-encoded public key which authenticates
// the participant of the NeoFS reputation system.
//
// Argument MUST NOT be mutated, make a copy first.
//
// Parameter key is a serialized compressed public key. See [elliptic.MarshalCompressed].
//
// See also [ComparePeerKey].
func (x *PeerID) SetPublicKey(key []byte) {
	x.key = key
}

// PublicKey return public key set using [PeerID.SetPublicKey].
//
// Zero [PeerID] has zero key which is incorrect according to NeoFS API
// protocol.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x PeerID) PublicKey() []byte {
	return x.key
}

// ComparePeerKey checks if the given PeerID corresponds to the party
// authenticated by the given binary public key.
//
// The key parameter is a slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
func ComparePeerKey(peer PeerID, key []byte) bool {
	return bytes.Equal(peer.PublicKey(), key)
}

// EncodeToString encodes ID into NeoFS API protocol string.
//
// Zero PeerID is base58 encoding of PeerIDSize zeros.
//
// See also DecodeString.
func (x PeerID) EncodeToString() string {
	return base58.Encode(x.key)
}

// DecodeString decodes string into PeerID according to NeoFS API protocol.
// Returns an error if s is malformed.
//
// See also DecodeString.
func (x *PeerID) DecodeString(s string) error {
	data, err := base58.Decode(s)
	if err != nil {
		return fmt.Errorf("decode base58: %w", err)
	}

	x.key = data

	return nil
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as EncodeToString. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (x PeerID) String() string {
	return x.EncodeToString()
}
