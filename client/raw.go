package client

import (
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
)

// Raw returns underlying raw protobuf client.
func (c *Client) Raw() *client.Client {
	return c.raw
}

// implements Client.Conn method.
func (c *Client) Conn() io.Closer {
	return c.raw.Conn()
}
