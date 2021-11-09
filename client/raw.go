package client

import (
	"io"

	"github.com/nspcc-dev/neofs-api-go/rpc/client"
)

// Raw returns underlying raw protobuf client.
func (c *clientImpl) Raw() *client.Client {
	return c.raw
}

// implements Client.Conn method.
func (c *clientImpl) Conn() io.Closer {
	return c.raw.Conn()
}
