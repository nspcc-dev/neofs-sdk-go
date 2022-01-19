package headers

import (
	objectSDK "github.com/nspcc-dev/neofs-sdk-go/object"
)

// WithObjectStorage return option that specifies
// storage that provides object by its address.
func WithObjectStorage(v ObjectStorage) Option {
	return func(c *cfg) {
		c.storage = v
	}
}

// WithMessageAndRequest returns option that specifies
// version independent message request/response xHeader
// source.
func WithMessageAndRequest(msg interface{}, request interface{}) Option {
	return func(c *cfg) {
		switch m := msg.(type) {
		default:
			return
		case Request:
			c.msg = &requestXHeaderSource{
				req: m,
			}
		case Response:
			c.msg = &responseXHeaderSource{
				resp: m,
				req:  request.(Request),
			}
		}
	}
}

func WithAddress(v *objectSDK.Address) Option {
	return func(c *cfg) {
		c.addr = v
	}
}
