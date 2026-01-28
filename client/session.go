package client

import (
	"context"
	"fmt"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
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

// ID returns identifier of the opened session in a binary NeoFS API protocol format.
//
// Client doesn't retain value so modification is safe.
func (x ResSessionCreate) ID() []byte {
	return x.id
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

	req := &protosession.CreateRequest{
		Body: &protosession.CreateRequest_Body{
			OwnerId:    signer.UserID().ProtoMessage(),
			Expiration: prm.exp,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protosession.CreateRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	resp, err := c.session.Create(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   c.nodeKey,
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return nil, err
		}
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, err
	}

	body := resp.GetBody()
	var res ResSessionCreate
	if res.id = body.GetId(); len(res.id) == 0 {
		err = newErrMissingResponseField("session id")
		return nil, err
	}

	if res.sessionKey = body.GetSessionKey(); len(res.sessionKey) == 0 {
		err = newErrMissingResponseField("session key")
		return nil, err
	}

	return &res, nil
}
