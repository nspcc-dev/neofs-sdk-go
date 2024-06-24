package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// StartSessionOptions groups optional parameters of [Client.StartSession].
type StartSessionOptions struct{}

// SessionData is a result of [Client.StartSession].
type SessionData struct {
	// Unique identifier of the session.
	ID uuid.UUID
	// Public session key authenticating the subject.
	PublicKey neofscrypto.PublicKey
}

// StartSession opens a session between given user and the node server on the
// remote endpoint expiring after the specified epoch. The session lifetime
// coincides with the server lifetime. Resulting SessionData is used to complete
// a session token issued by provided signer. Once session is started, remote
// server becomes an issuer's trusted party and session token represents a power
// of attorney. Complete session token can be used in some operations performed
// on behalf of the issuer but performed by the server. Now it's just simplified
// creation and deletion of objects ([Client.PutObject] and
// [Client.DeleteObject] respectively).
func (c *Client) StartSession(ctx context.Context, issuer user.Signer, exp uint64, _ StartSessionOptions) (SessionData, error) {
	var res SessionData
	if issuer == nil {
		return res, errMissingSigner
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodSessionCreate, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apisession.CreateRequest{
		Body: &apisession.CreateRequest_Body{
			OwnerId:    new(refs.OwnerID),
			Expiration: exp,
		},
	}
	issuer.UserID().WriteToV2(req.Body.OwnerId)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(issuer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.session.Create(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldID = "ID"
	if len(resp.Body.Id) == 0 {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldID) // for closure above
		return res, err
	} else if err = res.ID.UnmarshalBinary(resp.Body.Id); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldID, err) // for closure above
		return res, err
	} else if v := res.ID.Version(); v != 4 {
		err = fmt.Errorf("%s (%s): wrong UUID version %d", errInvalidResponseBodyField, fieldID, v) // for closure above
		return res, err
	}
	const fieldPubKey = "public session key"
	if len(resp.Body.SessionKey) == 0 {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldPubKey) // for closure above
		return res, err
	}
	res.PublicKey = new(neofsecdsa.PublicKey)
	if err = res.PublicKey.Decode(resp.Body.SessionKey); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldPubKey, err) // for closure above
		return res, err
	}
	return res, nil
}
