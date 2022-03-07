package client

import "github.com/nspcc-dev/neofs-api-go/v2/session"

// ResponseMetaInfo groups meta information about any NeoFS API response.
type ResponseMetaInfo struct {
	key []byte
}

type responseV2 interface {
	GetMetaHeader() *session.ResponseMetaHeader
	GetVerificationHeader() *session.ResponseVerificationHeader
}

// ResponderKey returns responder's public key in a binary format.
//
// Result must not be mutated.
func (x ResponseMetaInfo) ResponderKey() []byte {
	return x.key
}
