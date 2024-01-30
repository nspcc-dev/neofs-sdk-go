package slicer

import (
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// Options groups Slicer options.
type Options struct {
	objectPayloadLimit uint64

	currentNeoFSEpoch uint64

	withHomoChecksum bool

	copiesNumber uint32

	payloadBuffer []byte

	sessionToken *session.Object
	bearerToken  *bearer.Token

	payloadSizeFixed bool
	payloadSize      uint64
}

// SetObjectPayloadLimit specifies data size limit for produced physically
// stored objects.
func (x *Options) SetObjectPayloadLimit(l uint64) {
	x.objectPayloadLimit = l
}

// SetCurrentNeoFSEpoch sets current NeoFS epoch.
func (x *Options) SetCurrentNeoFSEpoch(e uint64) {
	x.currentNeoFSEpoch = e
}

// CalculateHomomorphicChecksum makes Slicer to calculate and set homomorphic
// checksum of the processed objects.
func (x *Options) CalculateHomomorphicChecksum() {
	x.withHomoChecksum = true
}

// SetSession sets session object.
func (x *Options) SetSession(sess session.Object) {
	x.sessionToken = &sess
}

// SetBearerToken allows to attach signed Extended ACL rules to the request.
func (x *Options) SetBearerToken(bearerToken bearer.Token) {
	x.bearerToken = &bearerToken
}

// SetCopiesNumber sets the minimal number of copies (out of the number specified by container placement policy) for
// the object PUT operation to succeed. This means that object operation will return with successful status even before
// container placement policy is completely satisfied.
func (x *Options) SetCopiesNumber(copiesNumber uint32) {
	x.copiesNumber = copiesNumber
}

// SetPayloadBuffer sets pre-allocated payloadBuffer to be used to object uploading.
// For better performance payloadBuffer length should be MaxObjectSize from NeoFS.
func (x *Options) SetPayloadBuffer(payloadBuffer []byte) {
	x.payloadBuffer = payloadBuffer
}

// SetPayloadSize allows to specify object's payload size known in advance. If
// set, reading functions will read at least size bytes while writing functions
// will expect exactly size bytes.
//
// If the size is known, the option is recommended as it improves the
// performance of the application using the [Slicer].
func (x *Options) SetPayloadSize(size uint64) {
	x.payloadSizeFixed = true
	x.payloadSize = size
}

// ObjectPayloadLimit returns required max object size.
func (x *Options) ObjectPayloadLimit() uint64 {
	return x.objectPayloadLimit
}

// CurrentNeoFSEpoch returns epoch.
func (x *Options) CurrentNeoFSEpoch() uint64 {
	return x.currentNeoFSEpoch
}

// IsHomomorphicChecksumEnabled indicates homomorphic checksum calculation status.
func (x *Options) IsHomomorphicChecksumEnabled() bool {
	return x.withHomoChecksum
}

// Session returns session object.
func (x *Options) Session() *session.Object {
	return x.sessionToken
}

// BearerToken returns bearer token.
func (x *Options) BearerToken() *bearer.Token {
	return x.bearerToken
}

// CopiesNumber returns the number of object copies.
func (x *Options) CopiesNumber() uint32 {
	return x.copiesNumber
}

// PayloadBuffer returns chunk which are using to object uploading.
func (x *Options) PayloadBuffer() []byte {
	return x.payloadBuffer
}
