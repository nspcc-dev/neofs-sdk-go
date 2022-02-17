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
// Result must not be mutated.
func (x ResponseMetaInfo) ResponderKey() []byte {
	return x.key
}

// Epoch return current epoch.
func (x ResponseMetaInfo) Epoch() uint64 {
	return x.epoch
}

// WithResponseInfoHandler allows specifying handler of response meta information for the all Client operations.
// The handler is called right after the response is received. Client returns handler's error immediately.
func WithResponseInfoHandler(f func(ResponseMetaInfo) error) Option {
	return func(opts *clientOptions) {
		opts.cbRespInfo = f
	}
}
