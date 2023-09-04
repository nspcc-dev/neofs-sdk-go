package reputation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
)

// PeerID represents unique identifier of the peer participating in the NeoFS
// reputation system.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/reputation.PeerID
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type PeerID struct {
	m reputation.PeerID
}

// ReadFromV2 reads PeerID from the reputation.PeerID message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *PeerID) ReadFromV2(m reputation.PeerID) error {
	val := m.GetPublicKey()
	if len(val) == 0 {
		return errors.New("missing ID bytes")
	}

	x.m = m

	return nil
}

// WriteToV2 writes PeerID to the reputation.PeerID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x PeerID) WriteToV2(m *reputation.PeerID) {
	*m = x.m
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
	x.m.SetPublicKey(key)
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
	return x.m.GetPublicKey()
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
	return base58.Encode(x.m.GetPublicKey())
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

	x.m.SetPublicKey(data)

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
