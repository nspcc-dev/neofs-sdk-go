package client

import (
	"crypto/ecdsa"
	"crypto/tls"
	"time"

	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"google.golang.org/grpc"
)

type (
	Option func(*clientOptions)

	clientOptions struct {
		key *ecdsa.PrivateKey

		rawOpts []client.Option

		cbRespInfo func(ResponseMetaInfo) error

		// defines if client parses erroneous NeoFS
		// statuses and returns them as `error`
		//
		// default is false
		parseNeoFSErrors bool

		netMagic uint64
	}
)

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		rawOpts: make([]client.Option, 0, 4),
	}
}

// WithAddress returns option to specify
// network address of the remote server.
//
// Ignored if WithGRPCConnection is provided.
func WithAddress(addr string) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithNetworkAddress(addr))
	}
}

// WithDialTimeout returns option to set connection timeout to the remote node.
//
// Ignored if WithGRPCConn is provided.
func WithDialTimeout(dur time.Duration) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithDialTimeout(dur))
	}
}

// WithRWTimeout returns option to set timeout for single read and write
// operation on protobuf message.
func WithRWTimeout(dur time.Duration) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithRWTimeout(dur))
	}
}

// WithTLSConfig returns option to set connection's TLS config to the remote node.
//
// Ignored if WithGRPCConnection is provided.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithTLSCfg(cfg))
	}
}

// WithDefaultPrivateKey returns option to set default private key
// used for the work.
func WithDefaultPrivateKey(key *ecdsa.PrivateKey) Option {
	return func(opts *clientOptions) {
		opts.key = key
	}
}

// WithURIAddress returns option to specify
// network address of a remote server and connection
// scheme for it.
//
// Format of the URI:
//
//		[scheme://]host:port
//
// Supported schemes:
//  - grpc;
//  - grpcs.
//
// tls.Cfg second argument is optional and is taken into
// account only in case of `grpcs` scheme.
//
// Falls back to WithNetworkAddress if address is not a valid URI.
//
// Do not use along with WithAddress and WithTLSConfig.
//
// Ignored if WithGRPCConnection is provided.
func WithURIAddress(addr string, tlsCfg *tls.Config) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithNetworkURIAddress(addr, tlsCfg)...)
	}
}

// WithGRPCConnection returns option to set GRPC connection to
// the remote node.
func WithGRPCConnection(grpcConn *grpc.ClientConn) Option {
	return func(opts *clientOptions) {
		opts.rawOpts = append(opts.rawOpts, client.WithGRPCConn(grpcConn))
	}
}

// WithNeoFSErrorParsing returns option that makes client parse
// erroneous NeoFS statuses and return them as `error` of the method
// call.
func WithNeoFSErrorParsing() Option {
	return func(opts *clientOptions) {
		opts.parseNeoFSErrors = true
	}
}

// WithNetworkMagic returns option to specify NeoFS network magic.
func WithNetworkMagic(magic uint64) Option {
	return func(opts *clientOptions) {
		opts.netMagic = magic
	}
}
