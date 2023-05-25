package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// PrmObjectDelete groups optional parameters of ObjectDelete operation.
type PrmObjectDelete struct {
	meta v2session.RequestMetaHeader

	keySet bool
	signer neofscrypto.Signer
}

// WithinSession specifies session within which object should be read.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *PrmObjectDelete) WithinSession(t session.Object) {
	var tv2 v2session.Token
	t.WriteToV2(&tv2)

	x.meta.SetSessionToken(&tv2)
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

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectDelete) UseSigner(signer neofscrypto.Signer) {
	x.keySet = true
	x.signer = signer
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
// Return errors:
//   - global (see Client docs)
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectLocked]
//   - [apistatus.ErrSessionTokenExpired]
func (c *Client) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectDelete) (oid.ID, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.DeleteRequestBody
	)

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return oid.ID{}, err
	}

	// form request body
	body.SetAddress(&addr)

	// form request
	var req v2object.DeleteRequest
	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)

	err = signServiceMessage(signer, &req)
	if err != nil {
		return oid.ID{}, fmt.Errorf("sign request: %w", err)
	}

	resp, err := rpcapi.DeleteObject(&c.c, &req, client.WithContext(ctx))
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
		return oid.ID{}, newErrMissingResponseField(fieldTombstone)
	}

	err = res.ReadFromV2(*idTombV2)
	if err != nil {
		return oid.ID{}, newErrInvalidResponseField(fieldTombstone, err)
	}

	return res, nil
}
