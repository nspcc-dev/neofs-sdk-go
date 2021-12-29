package client

import (
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
)

type Client struct {
	raw *client.Client

	opts *clientOptions
}

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
