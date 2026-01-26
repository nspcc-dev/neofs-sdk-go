package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// PrmContainerPut groups optional parameters of ContainerPut operation.
type PrmContainerPut struct {
	prmCommonMeta

	session   *session.Container
	sessionV2 *sessionv2.Token

	sigSet bool
	sig    neofscrypto.Signature
}

// WithinSession specifies session within which container should be saved.
//
// Creator of the session acquires the authorship of the request. This affects
// the execution of an operation (e.g. access control).
//
// Session is optional, if set the following requirements apply:
//   - session operation MUST be session.VerbContainerPut (ForVerb)
//   - token MUST be signed using private signer of the owner of the container to be saved
//   - MUST NOT be used together with WithinSessionV2
func (x *PrmContainerPut) WithinSession(s session.Container) {
	x.session = &s
}

// WithinSessionV2 specifies session token V2 within which container should be saved.
//
// Creator of the session acquires the authorship of the request. This affects
// the execution of an operation (e.g. access control).
//
// V2 tokens support multiple subjects, delegation chains, and unified contexts.
//
// Must be signed.
// MUST NOT be used together with WithinSession.
func (x *PrmContainerPut) WithinSessionV2(s sessionv2.Token) {
	x.sessionV2 = &s
}

// AttachSignature allows to attach pre-calculated container signature and free
// [Client.ContainerPut] from the calculation. The sig must have
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme.
func (x *PrmContainerPut) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// ContainerPut sends request to save container in NeoFS.
//
// Storage policy is required and limited by:
//   - 256 replica descriptors;
//   - 8 objects in each one;
//   - 64 nodes in any set, i.e. BF * selector count;
//   - 512 total nodes.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is async/await. Deadline is determined from ctx. If ctx has no
// deadline, server waits 15s after submitting the transaction. On timeout,
// [apistatus.ErrContainerAwaitTimeout] is returned. This error indicates a
// delay in transaction execution, meaning the operation may still succeed. If
// necessary, caller can continue the wait via [Client.ContainerGet] with
// returned ID. It is recommended to always use context with timeout. Note that
// the context includes all processing stages incl. network delays.
//
// Success can be verified by reading [Client.ContainerGet] using the returned
// identifier (notice that it needs some time to succeed).
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The account corresponding to the specified Signer will be charged for the operation.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// If signature already exists, use [PrmContainerPut.AttachSignature]:
// then signer will not be used.
//
// Return errors:
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerAwaitTimeout]
func (c *Client) ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm PrmContainerPut) (cid.ID, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerPut, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return cid.ID{}, ErrMissingSigner
	}
	if prm.session != nil && prm.sessionV2 != nil {
		return cid.ID{}, errSessionTokenBothVersionsSet
	}

	if err = cont.PlacementPolicy().Verify(); err != nil {
		err = fmt.Errorf("invalid storage policy: %w", err)
		return cid.ID{}, err
	}

	if !prm.sigSet {
		if err = cont.CalculateSignature(&prm.sig, signer); err != nil {
			err = fmt.Errorf("calculate container signature: %w", err)
			return cid.ID{}, err
		}
	}

	req := &protocontainer.PutRequest{
		Body: &protocontainer.PutRequest_Body{
			Container: cont.ProtoMessage(),
			Signature: &refs.SignatureRFC6979{
				Key:  prm.sig.PublicKeyBytes(),
				Sign: prm.sig.Value(),
			},
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
	if prm.sessionV2 != nil {
		req.MetaHeader.SessionTokenV2 = prm.sessionV2.ProtoMessage()
	}

	var res cid.ID

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.PutRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.container.Put(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return res, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return res, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.PutResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	var statusError error
	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		if !errors.Is(err, apistatus.ErrContainerAwaitTimeout) {
			return res, err
		}
		statusError = err
	}

	const fieldCnrID = "container ID"

	mCID := resp.GetBody().GetContainerId()
	if mCID == nil {
		err = newErrMissingResponseField(fieldCnrID)
		return res, err
	}

	err = res.FromProtoMessage(mCID)
	if err != nil {
		err = newErrInvalidResponseField(fieldCnrID, err)
		return res, err
	}

	return res, statusError
}

// PrmContainerGet groups optional parameters of ContainerGet operation.
type PrmContainerGet struct {
	prmCommonMeta
}

// ContainerGet reads NeoFS container by ID. The ID must not be zero.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ContainerGet(ctx context.Context, id cid.ID, prm PrmContainerGet) (container.Container, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerGet, time.Since(startTime), err)
		}()
	}

	req := &protocontainer.GetRequest{
		Body: &protocontainer.GetRequest_Body{
			ContainerId: id.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     2,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	var res container.Container
	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.GetRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.container.Get(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return res, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return res, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.GetResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
	}

	mc := resp.GetBody().GetContainer()
	if mc == nil {
		err = errors.New("missing container in response")
		return res, err
	}

	err = res.FromProtoMessage(mc)
	if err != nil {
		err = fmt.Errorf("invalid container in response: %w", err)
		return res, err
	}

	return res, nil
}

// PrmContainerList groups optional parameters of ContainerList operation.
type PrmContainerList struct {
	prmCommonMeta
}

// ContainerList requests identifiers of the account-owned containers.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ContainerList(ctx context.Context, ownerID user.ID, prm PrmContainerList) ([]cid.ID, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerList, time.Since(startTime), err)
		}()
	}

	req := &protocontainer.ListRequest{
		Body: &protocontainer.ListRequest_Body{
			OwnerId: ownerID.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.ListRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	resp, err := c.container.List(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return nil, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.ListResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return nil, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, err
	}

	ms := resp.GetBody().GetContainerIds()
	res := make([]cid.ID, len(ms))
	for i := range ms {
		if ms[i] == nil {
			err = newErrInvalidResponseField("ID list", fmt.Errorf("nil element #%d", i))
			return nil, err
		}
		if err = res[i].FromProtoMessage(ms[i]); err != nil {
			err = fmt.Errorf("invalid ID in the response: %w", err)
			return nil, err
		}
	}

	return res, nil
}

// PrmContainerDelete groups optional parameters of ContainerDelete operation.
type PrmContainerDelete struct {
	prmCommonMeta

	tok   *session.Container
	tokV2 *sessionv2.Token

	sigSet bool
	sig    neofscrypto.Signature
}

// WithinSession specifies session within which container should be removed.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
// MUST NOT be used together with WithinSessionV2.
func (x *PrmContainerDelete) WithinSession(tok session.Container) {
	x.tok = &tok
}

// WithinSessionV2 specifies session token V2 within which container should be removed.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// V2 tokens support multiple subjects, delegation chains, and unified contexts.
//
// Must be signed.
// MUST NOT be used together with WithinSession.
func (x *PrmContainerDelete) WithinSessionV2(tok sessionv2.Token) {
	x.tokV2 = &tok
}

// AttachSignature allows to attach pre-calculated container ID signature and
// free [Client.ContainerDelete] from the calculation. The sig must have
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme.
func (x *PrmContainerDelete) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// ContainerDelete sends request to remove the NeoFS container.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is async/await. Deadline is determined from ctx. If ctx has no
// deadline, server waits 15s after submitting the transaction. On timeout,
// [apistatus.ErrContainerAwaitTimeout] is returned. This error indicates a
// delay in transaction execution, meaning the operation may still succeed. If
// necessary, caller can continue the wait via [Client.ContainerGet]. It is
// recommended to always use context with timeout. Note that the context
// includes all processing stages incl. network delays.
//
// Success can be verified by reading by identifier (see GetContainer).
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The account corresponding to the specified Signer will be charged for the operation.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// If signature already exists, use [PrmContainerDelete.AttachSignature]:
// then signer will not be used.
//
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
// Return errors:
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerLocked]
//   - [apistatus.ErrContainerAwaitTimeout]
func (c *Client) ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm PrmContainerDelete) error {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerDelete, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return ErrMissingSigner
	}
	if signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return fmt.Errorf("%w: expected ECDSA_DETERMINISTIC_SHA256 scheme", neofscrypto.ErrIncorrectSigner)
	}
	if prm.tok != nil && prm.tokV2 != nil {
		return errSessionTokenBothVersionsSet
	}

	if !prm.sigSet {
		// container contract expects signature of container ID value
		// don't get confused with stable marshaled protobuf container.ID structure
		if err = prm.sig.Calculate(signer, id[:]); err != nil {
			err = fmt.Errorf("calculate container ID signature: %w", err)
			return err
		}
	}

	req := &protocontainer.DeleteRequest{
		Body: &protocontainer.DeleteRequest_Body{
			ContainerId: id.ProtoMessage(),
			Signature: &refs.SignatureRFC6979{
				Key:  prm.sig.PublicKeyBytes(),
				Sign: prm.sig.Value(),
			},
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.tok != nil {
		req.MetaHeader.SessionToken = prm.tok.ProtoMessage()
	}
	if prm.tokV2 != nil {
		req.MetaHeader.SessionTokenV2 = prm.tokV2.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.DeleteRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	resp, err := c.container.Delete(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.DeleteResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return err
	}

	err = apistatus.ToError(resp.GetMetaHeader().GetStatus())
	return err
}

// PrmContainerEACL groups optional parameters of ContainerEACL operation.
type PrmContainerEACL struct {
	prmCommonMeta
}

// ContainerEACL reads eACL table of the NeoFS container.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ContainerEACL(ctx context.Context, id cid.ID, prm PrmContainerEACL) (eacl.Table, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerEACL, time.Since(startTime), err)
		}()
	}

	req := &protocontainer.GetExtendedACLRequest{
		Body: &protocontainer.GetExtendedACLRequest_Body{
			ContainerId: id.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	var res eacl.Table

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.GetExtendedACLRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.container.GetExtendedACL(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return res, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return res, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.GetExtendedACLResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
	}

	const fieldEACL = "eACL"
	eACL := resp.GetBody().GetEacl()
	if eACL == nil {
		err = newErrMissingResponseField(fieldEACL)
		return res, err
	}
	if err = res.FromProtoMessage(eACL); err != nil {
		err = newErrInvalidResponseField(fieldEACL, err)
		return res, err
	}

	return res, nil
}

// PrmContainerSetEACL groups optional parameters of ContainerSetEACL operation.
type PrmContainerSetEACL struct {
	prmCommonMeta

	session   *session.Container
	sessionV2 *sessionv2.Token

	sigSet bool
	sig    neofscrypto.Signature
}

// WithinSession specifies session within which extended ACL of the container
// should be saved.
//
// Creator of the session acquires the authorship of the request. This affects
// the execution of an operation (e.g. access control).
//
// Session is optional, if set the following requirements apply:
//   - if particular container is specified (ApplyOnlyTo), it MUST equal the container
//     for which extended ACL is going to be set
//   - session operation MUST be session.VerbContainerSetEACL (ForVerb)
//   - token MUST be signed using private signer of the owner of the container to be saved
//   - MUST NOT be used together with WithinSessionV2
func (x *PrmContainerSetEACL) WithinSession(s session.Container) {
	x.session = &s
}

// WithinSessionV2 specifies session token V2 within which extended ACL of the
// container should be saved.
// Creator of the session acquires the authorship of the request. This affects
// the execution of an operation (e.g. access control).
//
// V2 tokens support multiple subjects, delegation chains, and unified contexts.
//
// Must be signed.
// MUST NOT be used together with WithinSession.
func (x *PrmContainerSetEACL) WithinSessionV2(s sessionv2.Token) {
	x.sessionV2 = &s
}

// AttachSignature allows to attach pre-calculated eACL signature and free
// [Client.ContainerSetEACL] from the calculation. The sig must have
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme.
func (x *PrmContainerSetEACL) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// ContainerSetEACL sends request to update eACL table of the NeoFS container.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is async/await. Deadline is determined from ctx. If ctx has no
// deadline, server waits 15s after submitting the transaction. On timeout,
// [apistatus.ErrContainerAwaitTimeout] is returned. This error indicates a
// delay in transaction execution, meaning the operation may still succeed. If
// necessary, caller can continue the wait via [Client.ContainerEACL]. It is
// recommended to always use context with timeout. Note that the context
// includes all processing stages incl. network delays.
//
// Success can be verified by reading by identifier (see EACL).
//
// Signer is required and must not be nil. The account corresponding to the specified Signer will be charged for the operation.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// If signature already exists, use [PrmContainerSetEACL.AttachSignature]:
// then signer will not be used.
//
// Return errors:
//   - [ErrMissingEACLContainer]
//   - [ErrMissingSigner]
//   - [apistatus.ErrContainerAwaitTimeout]
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm PrmContainerSetEACL) error {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerSetEACL, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return ErrMissingSigner
	}
	if prm.session != nil && prm.sessionV2 != nil {
		return errSessionTokenBothVersionsSet
	}

	if table.GetCID().IsZero() {
		err = ErrMissingEACLContainer
		return err
	}
	if signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return fmt.Errorf("%w: expected ECDSA_DETERMINISTIC_SHA256 scheme", neofscrypto.ErrIncorrectSigner)
	}

	// sign the eACL table
	mEACL := table.ProtoMessage()
	if !prm.sigSet {
		if err = prm.sig.Calculate(signer, neofsproto.MarshalMessage(mEACL)); err != nil {
			err = fmt.Errorf("calculate eACL signature: %w", err)
			return err
		}
	}

	req := &protocontainer.SetExtendedACLRequest{
		Body: &protocontainer.SetExtendedACLRequest_Body{
			Eacl: mEACL,
			Signature: &refs.SignatureRFC6979{
				Key:  prm.sig.PublicKeyBytes(),
				Sign: prm.sig.Value(),
			},
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
	if prm.sessionV2 != nil {
		req.MetaHeader.SessionTokenV2 = prm.sessionV2.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.SetExtendedACLRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	resp, err := c.container.SetExtendedACL(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.SetExtendedACLResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return err
	}

	err = apistatus.ToError(resp.GetMetaHeader().GetStatus())
	return err
}

// SyncContainerWithNetwork requests network configuration using passed [NetworkInfoExecutor]
// and applies/rewrites it to the container.
//
// Returns any network/parsing config errors.
//
// See also [client.Client.NetworkInfo], [container.Container.ApplyNetworkConfig].
func SyncContainerWithNetwork(ctx context.Context, cnr *container.Container, c NetworkInfoExecutor) error {
	if cnr == nil {
		return errors.New("empty container")
	}

	res, err := c.NetworkInfo(ctx, PrmNetworkInfo{})
	if err != nil {
		return fmt.Errorf("network info call: %w", err)
	}

	cnr.ApplyNetworkConfig(res)

	return nil
}

// SetContainerAttributeParameters groups signed parameters of
// [Client.SetContainerAttribute].
type SetContainerAttributeParameters struct {
	ID         cid.ID
	Attribute  string
	Value      string
	ValidUntil time.Time
}

// GetSignedSetContainerAttributeParameters returns signed message for prm.
func GetSignedSetContainerAttributeParameters(prm SetContainerAttributeParameters) []byte {
	return neofsproto.MarshalMessage(&protocontainer.SetAttributeRequest_Body_Parameters{
		ContainerId: prm.ID.ProtoMessage(),
		Attribute:   prm.Attribute,
		Value:       prm.Value,
		ValidUntil:  uint64(prm.ValidUntil.Unix()),
	})
}

// SignSetContainerAttributeParameters signs given prm.
func SignSetContainerAttributeParameters(signer neofscrypto.Signer, prm SetContainerAttributeParameters) (neofscrypto.Signature, error) {
	sig, err := signer.Sign(GetSignedSetContainerAttributeParameters(prm))
	if err != nil {
		return neofscrypto.Signature{}, err
	}

	return neofscrypto.NewSignature(signer.Scheme(), signer.Public(), sig), nil
}

// SetContainerAttributeOptions groups optional parameters of
// [Client.SetContainerAttribute].
type SetContainerAttributeOptions struct {
	sessionToken   *sessionv2.Token
	sessionTokenV1 *session.Container
}

// AttachSessionToken makes client to attach specified session token to the
// request.
func (x *SetContainerAttributeOptions) AttachSessionToken(tok sessionv2.Token) {
	x.sessionToken = &tok
}

// AttachSessionTokenV1 makes client to attach specified session token V1 to the
// request. AttachSessionTokenV1 must not be set together with
// [SetContainerAttributeOptions.AttachSessionToken] that is highly recommended
// to be used instead.
func (x *SetContainerAttributeOptions) AttachSessionTokenV1(tok session.Container) {
	x.sessionTokenV1 = &tok
}

// SetContainerAttribute sets container attribute.
//
// If container does not have the attribute, it is added. Otherwise, its value
// is swapped.
//
// [SetContainerAttributeParameters.ValidUntil] must not yet pass.
//
// [SetContainerAttributeParameters.Attribute] must be one of:
//   - CORS
//   - __NEOFS__LOCK_UNTIL
//   - S3_TAGS
//   - S3_SETTINGS
//   - S3_NOTIFICATIONS
//
// In general, requirements for [SetContainerAttributeParameters.Value] are the
// same as for container creation. Attribute-specific requirements:
//   - __NEOFS__LOCK_UNTIL: new timestamp must be after the current one if any
//   - S3_TAGS: must be a valid JSON object
//   - S3_SETTINGS: must be a valid JSON object
//   - S3_NOTIFICATIONS: must be a valid JSON object
//
// The prmSig must be a signature of prm in either [neofscrypto.N3] or
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme (using
// [SignSetContainerAttributeParameters] for example). It must be either
// container owner's or session subject's signature when using
// [SetContainerAttributeOptions.AttachSessionToken]
// ([SetContainerAttributeOptions.AttachSessionTokenV1]). If session is used,
// the token must be issued by the container owner and include
// [SetContainerAttributeParameters.ID] + [sessionv2.VerbContainerSetAttribute]
// ([session.VerbContainerSetAttribute]) context.
//
// Operation is async/await. Deadline is determined from ctx. If ctx has no
// deadline, server waits 15s after submitting the transaction. On timeout,
// [apistatus.ErrContainerAwaitTimeout] is returned. This error indicates a
// delay in transaction execution, meaning the operation may still succeed. If
// necessary, caller can continue the wait via [Client.ContainerGet]. It is
// recommended to always use context with timeout. Note that the context
// includes all processing stages incl. network delays.
func (c *Client) SetContainerAttribute(ctx context.Context, prm SetContainerAttributeParameters, prmSig neofscrypto.Signature, opts SetContainerAttributeOptions) error {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerSetAttribute, time.Since(startTime), err)
		}()
	}

	req := &protocontainer.SetAttributeRequest{
		Body: &protocontainer.SetAttributeRequest_Body{
			Parameters: &protocontainer.SetAttributeRequest_Body_Parameters{
				ContainerId: prm.ID.ProtoMessage(),
				Attribute:   prm.Attribute,
				Value:       prm.Value,
				ValidUntil:  uint64(prm.ValidUntil.Unix()),
			},
			Signature: &refs.SignatureRFC6979{
				Key:  prmSig.PublicKeyBytes(),
				Sign: prmSig.Value(),
			},
		},
	}
	if opts.sessionToken != nil {
		req.Body.SessionToken = opts.sessionToken.ProtoMessage()
	}
	if opts.sessionTokenV1 != nil {
		req.Body.SessionTokenV1 = opts.sessionTokenV1.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	var bodyBuf []byte
	if bodyLen := req.Body.MarshaledSize(); len(*buf) >= bodyLen {
		bodyBuf = (*buf)[:bodyLen]
	} else {
		bodyBuf = make([]byte, bodyLen)
		buf = &bodyBuf
	}
	req.Body.MarshalStable(bodyBuf)

	bodySig, err := c.prm.signer.Sign(bodyBuf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	req.BodySignature = &refs.Signature{
		Key:    neofscrypto.PublicKeyBytes(c.prm.signer.Public()),
		Sign:   bodySig,
		Scheme: refs.SignatureScheme(c.prm.signer.Scheme()),
	}

	resp, err := c.container.SetAttribute(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	err = apistatus.ToError(resp.Status)

	return err
}

// RemoveContainerAttributeParameters groups signed parameters of
// [Client.RemoveContainerAttribute].
type RemoveContainerAttributeParameters struct {
	ID         cid.ID
	Attribute  string
	ValidUntil time.Time
}

// GetSignedRemoveContainerAttributeParameters returns signed message for prm.
func GetSignedRemoveContainerAttributeParameters(prm RemoveContainerAttributeParameters) []byte {
	return neofsproto.MarshalMessage(&protocontainer.RemoveAttributeRequest_Body_Parameters{
		ContainerId: prm.ID.ProtoMessage(),
		Attribute:   prm.Attribute,
		ValidUntil:  uint64(prm.ValidUntil.Unix()),
	})
}

// SignRemoveContainerAttributeParameters signs given prm.
func SignRemoveContainerAttributeParameters(signer neofscrypto.Signer, prm RemoveContainerAttributeParameters) (neofscrypto.Signature, error) {
	sig, err := signer.Sign(GetSignedRemoveContainerAttributeParameters(prm))
	if err != nil {
		return neofscrypto.Signature{}, err
	}

	return neofscrypto.NewSignature(signer.Scheme(), signer.Public(), sig), nil
}

// RemoveContainerAttributeOptions groups optional parameters of
// [Client.RemoveContainerAttribute].
type RemoveContainerAttributeOptions struct {
	sessionToken   *sessionv2.Token
	sessionTokenV1 *session.Container
}

// AttachSessionToken makes client to attach specified session token to the
// request.
func (x *RemoveContainerAttributeOptions) AttachSessionToken(tok sessionv2.Token) {
	x.sessionToken = &tok
}

// AttachSessionTokenV1 makes client to attach specified session token V1 to the
// request. AttachSessionTokenV1 must not be set together with
// [RemoveContainerAttributeOptions.AttachSessionToken] that is highly
// recommended to be used instead.
func (x *RemoveContainerAttributeOptions) AttachSessionTokenV1(tok session.Container) {
	x.sessionTokenV1 = &tok
}

// RemoveContainerAttribute removes container attribute.
//
// If container does not have the attribute, server does nothing and responds
// without an error.
//
// [RemoveContainerAttributeParameters.ValidUntil] must not yet pass.
//
// [RemoveContainerAttributeParameters.Attribute] must be one of:
//   - CORS
//   - __NEOFS__LOCK_UNTIL
//   - S3_TAGS
//   - S3_SETTINGS
//   - S3_NOTIFICATIONS
//
// Attribute-specific requirements:
//   - __NEOFS__LOCK_UNTIL: current timestamp must have already passed if any
//
// The prmSig must be a signature of prm in either [neofscrypto.N3] or
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme (using
// [SignRemoveContainerAttributeParameters] for example). It must be either
// container owner's or session subject's signature when using
// [RemoveContainerAttributeOptions.AttachSessionToken]
// ([RemoveContainerAttributeOptions.AttachSessionTokenV1]). If session is used,
// the token must be issued by the container owner and include
// [RemoveContainerAttributeParameters.ID] +
// [sessionv2.VerbContainerRemoveAttribute]
// ([session.VerbContainerRemoveAttribute]) context.
//
// Operation is async/await. Deadline is determined from ctx. If ctx has no
// deadline, server waits 15s after submitting the transaction. On timeout,
// [apistatus.ErrContainerAwaitTimeout] is returned. This error indicates a
// delay in transaction execution, meaning the operation may still succeed. If
// necessary, caller can continue the wait via [Client.ContainerGet]. It is
// recommended to always use context with timeout. Note that the context
// includes all processing stages incl. network delays.
func (c *Client) RemoveContainerAttribute(ctx context.Context, prm RemoveContainerAttributeParameters, prmSig neofscrypto.Signature, opts RemoveContainerAttributeOptions) error {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerRemoveAttribute, time.Since(startTime), err)
		}()
	}

	req := &protocontainer.RemoveAttributeRequest{
		Body: &protocontainer.RemoveAttributeRequest_Body{
			Parameters: &protocontainer.RemoveAttributeRequest_Body_Parameters{
				ContainerId: prm.ID.ProtoMessage(),
				Attribute:   prm.Attribute,
				ValidUntil:  uint64(prm.ValidUntil.Unix()),
			},
			Signature: &refs.SignatureRFC6979{
				Key:  prmSig.PublicKeyBytes(),
				Sign: prmSig.Value(),
			},
		},
	}
	if opts.sessionToken != nil {
		req.Body.SessionToken = opts.sessionToken.ProtoMessage()
	}
	if opts.sessionTokenV1 != nil {
		req.Body.SessionTokenV1 = opts.sessionTokenV1.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	var bodyBuf []byte
	if bodyLen := req.Body.MarshaledSize(); len(*buf) >= bodyLen {
		bodyBuf = (*buf)[:bodyLen]
	} else {
		bodyBuf = make([]byte, bodyLen)
		buf = &bodyBuf
	}
	req.Body.MarshalStable(bodyBuf)

	bodySig, err := c.prm.signer.Sign(bodyBuf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	req.BodySignature = &refs.Signature{
		Key:    neofscrypto.PublicKeyBytes(c.prm.signer.Public()),
		Sign:   bodySig,
		Scheme: refs.SignatureScheme(c.prm.signer.Scheme()),
	}

	resp, err := c.container.RemoveAttribute(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	err = apistatus.ToError(resp.Status)

	return err
}
