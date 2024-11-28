package client

import (
	"context"
	"time"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmSessionCreate groups parameters of SessionCreate operation.
type PrmSessionCreate struct {
	prmCommonMeta

	exp uint64
}

// SetExp sets number of the last NepFS epoch in the lifetime of the session after which it will be expired.
func (x *PrmSessionCreate) SetExp(exp uint64) {
	x.exp = exp
}

// ResSessionCreate groups resulting values of SessionCreate operation.
type ResSessionCreate struct {
	id []byte

	sessionKey []byte
}

// NewResSessionCreate is a constructor for NewResSessionCreate.
func NewResSessionCreate(id []byte, sessionKey []byte) ResSessionCreate {
	return ResSessionCreate{
		id:         id,
		sessionKey: sessionKey,
	}
}

func (x *ResSessionCreate) setID(id []byte) {
	x.id = id
}

// ID returns identifier of the opened session in a binary NeoFS API protocol format.
//
// Client doesn't retain value so modification is safe.
func (x ResSessionCreate) ID() []byte {
	return x.id
}

func (x *ResSessionCreate) setSessionKey(key []byte) {
	x.sessionKey = key
}

// PublicKey returns public key of the opened session in a binary NeoFS API protocol format.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x ResSessionCreate) PublicKey() []byte {
	return x.sessionKey
}

// SessionCreate opens a session with the node server on the remote endpoint.
// The session lifetime coincides with the server lifetime. Results can be written
// to session token which can be later attached to the requests.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The account will be used as owner of new session.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) SessionCreate(ctx context.Context, signer user.Signer, prm PrmSessionCreate) (*ResSessionCreate, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodSessionCreate, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	ownerID := signer.UserID()

	var ownerIDV2 refs.OwnerID
	ownerID.WriteToV2(&ownerIDV2)

	// form request body
	reqBody := new(v2session.CreateRequestBody)
	reqBody.SetOwnerID(&ownerIDV2)
	reqBody.SetExpiration(prm.exp)

	// for request
	var req v2session.CreateRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ResSessionCreate
	)

	c.initCallContext(&cc)
	cc.signer = signer
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		resp, err := c.session.Create(ctx, req.ToGRPCMessage().(*protosession.CreateRequest))
		if err != nil {
			return nil, rpcErr(err)
		}
		var respV2 v2session.CreateResponse
		if err = respV2.FromGRPCMessage(resp); err != nil {
			return nil, err
		}
		return &respV2, nil
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2session.CreateResponse)

		body := resp.GetBody()

		if len(body.GetID()) == 0 {
			cc.err = newErrMissingResponseField("session id")
			return
		}

		if len(body.GetSessionKey()) == 0 {
			cc.err = newErrMissingResponseField("session key")
			return
		}

		res.setID(body.GetID())
		res.setSessionKey(body.GetSessionKey())
	}

	// process call
	if !cc.processCall() {
		err = cc.err
		return nil, cc.err
	}

	return &res, nil
}
