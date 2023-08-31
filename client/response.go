package client

import "github.com/nspcc-dev/neofs-api-go/v2/session"

// ResponseMetaInfo groups meta information about any NeoFS API response.
type ResponseMetaInfo struct {
	key []byte

	epoch uint64
}

type responseV2 interface {
	GetMetaHeader() *session.ResponseMetaHeader
	GetVerificationHeader() *session.ResponseVerificationHeader
}

// ResponderKey returns responder's public key in a binary format.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x ResponseMetaInfo) ResponderKey() []byte {
	return x.key
}

// Epoch returns local NeoFS epoch of the server.
func (x ResponseMetaInfo) Epoch() uint64 {
	return x.epoch
}
