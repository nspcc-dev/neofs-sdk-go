package client

import (
	"context"
	"errors"
	"fmt"

	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

var errMalformedResponseBody = errors.New("malformed response body")

type CreateSessionRes struct {
	statusRes

	id []byte

	sessionKey []byte
}

func (x *CreateSessionRes) setID(id []byte) {
	x.id = id
}

func (x CreateSessionRes) ID() []byte {
	return x.id
}

func (x *CreateSessionRes) setSessionKey(key []byte) {
	x.sessionKey = key
}

func (x CreateSessionRes) SessionKey() []byte {
	return x.sessionKey
}

// CreateSession creates session through NeoFS API call.
//
// Any client's internal or transport errors are returned as error,
// NeoFS status codes are included in the returned results.
func (c *Client) CreateSession(ctx context.Context, expiration uint64, opts ...CallOption) (*CreateSessionRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	ownerID := owner.NewIDFromPublicKey(&callOptions.key.PublicKey)

	reqBody := new(v2session.CreateRequestBody)
	reqBody.SetOwnerID(ownerID.ToV2())
	reqBody.SetExpiration(expiration)

	req := new(v2session.CreateRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.CreateSession(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(CreateSessionRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	body := resp.GetBody()

	res.setID(body.GetID())
	res.setSessionKey(body.GetSessionKey())

	return res, nil
}
