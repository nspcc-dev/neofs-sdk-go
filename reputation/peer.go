package reputation

import (
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-sdk-go/api/reputation"
)

// PeerIDSize is an ID size of the peer participating in the NeoFS reputation
// system.
const PeerIDSize = 33

// PeerID represents unique identifier of the peer participating in the NeoFS
// reputation system. PeerID corresponds to the binary-encoded public key in
// a format similar to the NeoFS network map.
//
// PeerID implements built-in comparable interface.
//
// ID is mutually compatible with [reputation.PeerID] message. See
// [PeerID.ReadFromV2] / [PeerID.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type PeerID [PeerIDSize]byte

func (x *PeerID) decodeBinary(b []byte) error {
	if len(b) != PeerIDSize {
		return fmt.Errorf("invalid value length %d", len(b))
	}
	copy(x[:], b)
	return nil
}

// ReadFromV2 reads PeerID from the reputation.PeerID message. Returns an error
// if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PeerID.WriteToV2].
func (x *PeerID) ReadFromV2(m *reputation.PeerID) error {
	if len(m.PublicKey) == 0 {
		return errors.New("missing value field")
	}
	return x.decodeBinary(m.PublicKey)
}

// WriteToV2 writes PeerID to the reputation.PeerID message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PeerID.ReadFromV2].
func (x PeerID) WriteToV2(m *reputation.PeerID) {
	m.PublicKey = x[:]
}

// EncodeToString encodes PeerID into NeoFS API V2 protocol string.
//
// Zero ID is base58 encoding of [PeerIDSize] zeros.
//
// See also [PeerID.DecodeString].
func (x PeerID) EncodeToString() string {
	return base58.Encode(x[:])
}

// DecodeString decodes string into PeerID according to NeoFS API protocol.
// Returns an error if s is malformed.
//
// See also [PeerID.EncodeToString].
func (x *PeerID) DecodeString(s string) error {
	var b []byte
	if s != "" {
		var err error
		b, err = base58.Decode(s)
		if err != nil {
			return fmt.Errorf("decode base58: %w", err)
		}
	}
	return x.decodeBinary(b)
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as EncodeToString. String MUST
// NOT be used to encode PeerID into NeoFS protocol string.
func (x PeerID) String() string {
	return x.EncodeToString()
}
