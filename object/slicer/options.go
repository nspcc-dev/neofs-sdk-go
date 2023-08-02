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

	sessionToken *session.Object
	bearerToken  *bearer.Token
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
func (x *Options) SetSession(sess *session.Object) {
	x.sessionToken = sess
}

// SetBearerToken allows to attach signed Extended ACL rules to the request.
func (x *Options) SetBearerToken(bearerToken *bearer.Token) {
	x.bearerToken = bearerToken
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
