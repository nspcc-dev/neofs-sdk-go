package client

import (
	"context"
	"crypto/ecdsa"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// PrmObjectDelete groups parameters of ObjectDelete operation.
type PrmObjectDelete struct {
	meta v2session.RequestMetaHeader

	body v2object.DeleteRequestBody

	addr v2refs.Address

	keySet bool
	key    ecdsa.PrivateKey
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

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *PrmObjectDelete) FromContainer(id cid.ID) {
	var cidV2 v2refs.ContainerID
	id.WriteToV2(&cidV2)

	x.addr.SetContainerID(&cidV2)
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *PrmObjectDelete) ByID(id oid.ID) {
	var idV2 v2refs.ObjectID
	id.WriteToV2(&idV2)

	x.addr.SetObjectID(&idV2)
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *PrmObjectDelete) UseKey(key ecdsa.PrivateKey) {
	x.keySet = true
	x.key = key
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *PrmObjectDelete) WithXHeaders(hs ...string) {
	if len(hs)%2 != 0 {
		panic("slice of X-Headers with odd length")
	}

	prmCommonMeta{xHeaders: hs}.writeToMetaHeader(&x.meta)
}

// ResObjectDelete groups resulting values of ObjectDelete operation.
type ResObjectDelete struct {
	statusRes

	tomb oid.ID
}

// Tombstone returns identifier of the created tombstone object.
func (x ResObjectDelete) Tombstone() oid.ID {
	return x.tomb
}

// ObjectDelete marks an object for deletion from the container using NeoFS API protocol.
// As a marker, a special unit called a tombstone is placed in the container.
// It confirms the user's intent to delete the object, and is itself a container object.
// Explicit deletion is done asynchronously, and is generally not guaranteed.
//
// Returns a list of checksums in raw form: the format of hashes and their number
// is left for the caller to check. Client preserves the order of the server's response.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// If PrmInit.ResolveNeoFSFailures has been called, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectDelete docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//   - global (see Client docs)
//   - *apistatus.ContainerNotFound;
//   - *apistatus.ObjectAccessDenied;
//   - *apistatus.ObjectLocked;
//   - *apistatus.SessionTokenExpired.
func (c *Client) ObjectDelete(ctx context.Context, prm PrmObjectDelete) (*ResObjectDelete, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic("missing object")
	}

	// form request body
	prm.body.SetAddress(&prm.addr)

	// form request
	var req v2object.DeleteRequest
	req.SetBody(&prm.body)
	req.SetMetaHeader(&prm.meta)

	// init call context
	var (
		cc  contextCall
		res ResObjectDelete
	)

	c.initCallContext(&cc)
	if prm.keySet {
		cc.key = prm.key
	}

	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.DeleteObject(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		const fieldTombstone = "tombstone"

		idTombV2 := r.(*v2object.DeleteResponse).GetBody().GetTombstone().GetObjectID()
		if idTombV2 == nil {
			cc.err = newErrMissingResponseField(fieldTombstone)
			return
		}

		cc.err = res.tomb.ReadFromV2(*idTombV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldTombstone, cc.err)
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
