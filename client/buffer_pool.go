package client

import (
	"github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
)

const defaultRequestBufferLength = 32 << 10

var defaultRequestBufferPool = protobuf.NewBufferPool(defaultRequestBufferLength)
