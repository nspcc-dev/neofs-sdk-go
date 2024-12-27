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
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// PrmContainerPut groups optional parameters of ContainerPut operation.
type PrmContainerPut struct {
	prmCommonMeta

	sessionSet bool
	session    session.Container

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
func (x *PrmContainerPut) WithinSession(s session.Container) {
	x.session = s
	x.sessionSet = true
}

// AttachSignature allows to attach pre-calculated container signature and free
// [Client.ContainerPut] from the calculation. The sig must have
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme.
func (x *PrmContainerPut) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// ContainerPut sends request to save container in NeoFS.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
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
	if prm.sessionSet {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
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

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
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

	return res, nil
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

	tokSet bool
	tok    session.Container

	sigSet bool
	sig    neofscrypto.Signature
}

// WithinSession specifies session within which container should be removed.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *PrmContainerDelete) WithinSession(tok session.Container) {
	x.tok = tok
	x.tokSet = true
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
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
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
	writeXHeadersToMeta(prm.prmCommonMeta.xHeaders, req.MetaHeader)
	if prm.tokSet {
		req.MetaHeader.SessionToken = prm.tok.ProtoMessage()
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

	sessionSet bool
	session    session.Container

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
func (x *PrmContainerSetEACL) WithinSession(s session.Container) {
	x.session = s
	x.sessionSet = true
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
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
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
	writeXHeadersToMeta(prm.prmCommonMeta.xHeaders, req.MetaHeader)
	if prm.sessionSet {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
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

// PrmAnnounceSpace groups optional parameters of ContainerAnnounceUsedSpace operation.
type PrmAnnounceSpace struct {
	prmCommonMeta
}

// ContainerAnnounceUsedSpace sends request to announce volume of the space used for the container objects.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// At this moment success can not be checked.
//
// Context is required and must not be nil. It is used for network communication.
//
// Announcements parameter MUST NOT be empty slice.
//
// Return errors:
//   - [ErrMissingAnnouncements]
func (c *Client) ContainerAnnounceUsedSpace(ctx context.Context, announcements []container.SizeEstimation, prm PrmAnnounceSpace) error {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodContainerAnnounceUsedSpace, time.Since(startTime), err)
		}()
	}

	if len(announcements) == 0 {
		err = ErrMissingAnnouncements
		return err
	}

	req := &protocontainer.AnnounceUsedSpaceRequest{
		Body: &protocontainer.AnnounceUsedSpaceRequest_Body{
			Announcements: make([]*protocontainer.AnnounceUsedSpaceRequest_Body_Announcement, len(announcements)),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	for i := range announcements {
		req.Body.Announcements[i] = announcements[i].ProtoMessage()
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protocontainer.AnnounceUsedSpaceRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	resp, err := c.container.AnnounceUsedSpace(ctx, req)
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

	if err = neofscrypto.VerifyResponseWithBuffer[*protocontainer.AnnounceUsedSpaceResponse_Body](resp, *buf); err != nil {
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
