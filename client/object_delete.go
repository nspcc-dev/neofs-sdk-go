package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

var (
	// ErrNoSession indicates that session wasn't set in some Prm* structure.
	ErrNoSession = errors.New("session is not set")
)

// PrmObjectDelete groups optional parameters of ObjectDelete operation.
type PrmObjectDelete struct {
	prmCommonMeta
	sessionContainer
	bearerToken *bearer.Token
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectDelete) WithBearerToken(t bearer.Token) {
	x.bearerToken = &t
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
// If signer implements [neofscrypto.SignerV2], signing is done using it. In
// this case, [neofscrypto.Signer] methods are not called.
// [neofscrypto.OverlapSigner] may be used to pass [neofscrypto.SignerV2] when
// [neofscrypto.Signer] is unimplemented.
//
// Return errors:
//   - global (see Client docs)
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectLocked]
//   - [apistatus.ErrSessionTokenExpired]
func (c *Client) ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectDelete) (oid.ID, error) {
	var err error

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectDelete, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return oid.ID{}, ErrMissingSigner
	}

	req := &protoobject.DeleteRequest{
		Body: &protoobject.DeleteRequest_Body{
			Address: oid.NewAddress(containerID, objectID).ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.session != nil {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.DeleteRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return oid.ID{}, err
	}

	resp, err := c.object.Delete(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return oid.ID{}, err
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protoobject.DeleteResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return oid.ID{}, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return oid.ID{}, err
	}

	const fieldTombstone = "tombstone"

	mt := resp.GetBody().GetTombstone().GetObjectId()
	if mt == nil {
		err = newErrMissingResponseField(fieldTombstone)
		return oid.ID{}, err
	}

	var res oid.ID
	err = res.FromProtoMessage(mt)
	if err != nil {
		err = newErrInvalidResponseField(fieldTombstone, err)
		return oid.ID{}, err
	}

	return res, nil
}
