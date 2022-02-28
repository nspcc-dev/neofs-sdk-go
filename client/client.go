package client

import (
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
)

// Client is a wrapper over raw NeoFS API client.
//
// It is not allowed to override client's behaviour:
// the parameters for the all operations are write-only
// and the results of the all operations are read-only.
//
// Working client must be created via constructor New.
// Using the Client that has been created with new(Client)
// expression (or just declaring a Client variable) is unsafe
// and can lead to panic.
//
// Each method which produces a NeoFS API call may return a server response.
// Status responses are returned in the result structure, and can be cast
// to built-in error instance (or in the returned error if the client is
// configured accordingly). Certain statuses can be checked using `apistatus`
// and standard `errors` packages. Note that package provides some helper
// functions to work with status returns (e.g. IsErrContainerNotFound).
// All possible responses are documented in methods, however, some may be
// returned from all of them (pay attention to the presence of the pointer sign):
//  - *apistatus.ServerInternal on internal server error;
//  - *apistatus.SuccessDefaultV2 on default success.
type Client struct {
	raw *client.Client

	opts *clientOptions
}

// New creates, initializes and returns the Client instance.
//
// If multiple options of the same config value are supplied,
// the option with the highest index in the arguments will be used.
func New(opts ...Option) (*Client, error) {
	clientOptions := defaultClientOptions()

	for i := range opts {
		opts[i](clientOptions)
	}

	return &Client{
		opts: clientOptions,
		raw:  client.New(clientOptions.rawOpts...),
	}, nil
}
