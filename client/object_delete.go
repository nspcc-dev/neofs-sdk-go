package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

var (
	// special variable for test purposes only, to overwrite real RPC calls.
	rpcAPIDeleteObject = rpcapi.DeleteObject

	// ErrNoSession indicates that session wasn't set in some Prm* structure.
	ErrNoSession = errors.New("session is not set")
)

// PrmObjectDelete groups optional parameters of ObjectDelete operation.
type PrmObjectDelete struct {
	sessionContainer
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectDelete) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *PrmObjectDelete) WithXHeaders(hs ...string) {
	writeXHeadersToMeta(hs, &x.meta)
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
// As a marker, a special unit called a tombstone is placed in the container.
// It confirms the user's intent to delete the object, and is itself a container object.
// Explicit deletion is done asynchronously, and is generally not guaranteed.
//
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of
// the account corresponding to the specified Signer, which is taken into account, in particular, for access control.
//
// Return errors:
//   - global (see Client docs)
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectLocked]
//   - [apistatus.ErrSessionTokenExpired]
func (c *Client) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm PrmObjectDelete) (oid.ID, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.DeleteRequestBody
		err   error
	)

	defer func() {
		c.sendStatistic(stat.MethodObjectDelete, err)()
	}()

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	if signer == nil {
		return oid.ID{}, ErrMissingSigner
	}

	// form request body
	body.SetAddress(&addr)

	// form request
	var req v2object.DeleteRequest
	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)

	err = signServiceMessage(signer, &req)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return oid.ID{}, err
	}

	resp, err := rpcAPIDeleteObject(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		return oid.ID{}, err
	}

	var res oid.ID
	if err = c.processResponse(resp); err != nil {
		return oid.ID{}, err
	}

	const fieldTombstone = "tombstone"

	idTombV2 := resp.GetBody().GetTombstone().GetObjectID()
	if idTombV2 == nil {
		err = newErrMissingResponseField(fieldTombstone)
		return oid.ID{}, err
	}

	err = res.ReadFromV2(*idTombV2)
	if err != nil {
		err = newErrInvalidResponseField(fieldTombstone, err)
		return oid.ID{}, err
	}

	return res, nil
}
