package client

import (
	"context"

	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// CreateSessionPrm groups parameters of CreateSession operation.
type CreateSessionPrm struct {
	exp uint64
}

// SetExp sets number of the last NepFS epoch in the lifetime of the session after which it will be expired.
func (x *CreateSessionPrm) SetExp(exp uint64) {
	x.exp = exp
}

// CreateSessionRes groups resulting values of CreateSession operation.
type CreateSessionRes struct {
	statusRes

	id []byte

	sessionKey []byte
}

func (x *CreateSessionRes) setID(id []byte) {
	x.id = id
}

// ID returns identifier of the opened session in a binary NeoFS API protocol format.
//
// Client doesn't retain value so modification is safe.
func (x CreateSessionRes) ID() []byte {
	return x.id
}

func (x *CreateSessionRes) setSessionKey(key []byte) {
	x.sessionKey = key
}

// PublicKey returns public key of the opened session in a binary NeoFS API protocol format.
func (x CreateSessionRes) PublicKey() []byte {
	return x.sessionKey
}

// CreateSession opens a session with the node server on the remote endpoint.
// The session lifetime coincides with the server lifetime. Results can be written
// to session token which can be later attached to the requests.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see CreateSessionPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) CreateSession(ctx context.Context, prm CreateSessionPrm) (*CreateSessionRes, error) {
	// check context
	if ctx == nil {
		panic(panicMsgMissingContext)
	}

	ownerID := owner.NewIDFromPublicKey(&c.opts.key.PublicKey)

	// form request body
	reqBody := new(v2session.CreateRequestBody)
	reqBody.SetOwnerID(ownerID.ToV2())
	reqBody.SetExpiration(prm.exp)

	// for request
	var req v2session.CreateRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res CreateSessionRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.CreateSession(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2session.CreateResponse)

		body := resp.GetBody()

		res.setID(body.GetID())
		res.setSessionKey(body.GetSessionKey())
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
