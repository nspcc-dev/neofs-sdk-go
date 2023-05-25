package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmSessionCreate groups parameters of SessionCreate operation.
type PrmSessionCreate struct {
	prmCommonMeta

	exp uint64

	signer neofscrypto.Signer
}

// SetExp sets number of the last NepFS epoch in the lifetime of the session after which it will be expired.
func (x *PrmSessionCreate) SetExp(exp uint64) {
	x.exp = exp
}

// UseSigner specifies private signer to sign the requests and compute token owner.
// If signer is not provided, then Client default signer is used.
func (x *PrmSessionCreate) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// ResSessionCreate groups resulting values of SessionCreate operation.
type ResSessionCreate struct {
	id []byte

	sessionKey []byte
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
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) SessionCreate(ctx context.Context, prm PrmSessionCreate) (*ResSessionCreate, error) {
	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
	}

	var ownerID user.ID
	if err = user.IDFromSigner(&ownerID, signer); err != nil {
		return nil, fmt.Errorf("IDFromSigner: %w", err)
	}

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
		return c.server.createSession(&c.c, &req, client.WithContext(ctx))
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
		return nil, cc.err
	}

	return &res, nil
}
