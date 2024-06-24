package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PutContainerOptions groups optional parameters of [Client.PutContainer].
type PutContainerOptions struct {
	sessionSet bool
	session    session.Container
}

// WithinSession specifies session within which container should be saved.
// Session tokens grant user-to-user power of attorney: the subject can create
// containers on behalf of the issuer. Session op must be
// [session.VerbContainerPut]. If used, [Client.PutContainer] default ownership
// behavior is replaced with:
//   - session issuer becomes the container's owner;
//   - session must target the subject authenticated by signer passed to
//     [Client.PutContainer].
func (x *PutContainerOptions) WithinSession(s session.Container) {
	x.session = s
	x.sessionSet = true
}

// PutContainer sends request to save given container in NeoFS. If the request
// is accepted, PutContainer returns no error and ID of the new container which
// is going to be saved asynchronously. The completion can be checked by polling
// the container presence using returned ID until it will appear (e.g. call
// [Client.GetContainer] until no error).
//
// Signer must authenticate container's owner. The signature scheme MUST be
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] (e.g. [neofsecdsa.SignerRFC6979]).
// Owner's NeoFS account can be charged for the operation.
func (c *Client) PutContainer(ctx context.Context, cnr container.Container, signer neofscrypto.Signer, opts PutContainerOptions) (cid.ID, error) {
	var res cid.ID
	if signer == nil {
		return res, errMissingSigner
	} else if scheme := signer.Scheme(); scheme != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return res, fmt.Errorf("wrong signature scheme: %v instead of %v", scheme, neofscrypto.ECDSA_DETERMINISTIC_SHA256)
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerPut, time.Since(start), err)
		}(time.Now())
	}

	sig, err := signer.Sign(cnr.Marshal())
	if err != nil {
		err = fmt.Errorf("sign container: %w", err) // for closure above
		return res, err
	}

	// form request
	req := &apicontainer.PutRequest{
		Body: &apicontainer.PutRequest_Body{
			Container: new(apicontainer.Container),
			Signature: &refs.SignatureRFC6979{Key: neofscrypto.PublicKeyBytes(signer.Public()), Sign: sig},
		},
	}
	cnr.WriteToV2(req.Body.Container)
	if opts.sessionSet {
		req.MetaHeader = &apisession.RequestMetaHeader{SessionToken: new(apisession.SessionToken)}
		opts.session.WriteToV2(req.MetaHeader.SessionToken)
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.container.Put(ctx, req)
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
	if resp.Body.ContainerId == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldID) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.ContainerId); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldID, err) // for closure above
		return res, err
	}
	return res, nil
}

// GetContainerOptions groups optional parameters of [Client.GetContainer].
type GetContainerOptions struct{}

// GetContainer reads NeoFS container by ID. Returns
// [apistatus.ErrContainerNotFound] if there is no such container.
func (c *Client) GetContainer(ctx context.Context, id cid.ID, _ GetContainerOptions) (container.Container, error) {
	var res container.Container
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerGet, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apicontainer.GetRequest{
		Body: &apicontainer.GetRequest_Body{ContainerId: new(refs.ContainerID)},
	}
	id.WriteToV2(req.Body.ContainerId)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.container.Get(ctx, req)
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
	const fieldContainer = "container"
	if resp.Body.Container == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldContainer) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.Container); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldContainer, err) // for closure above
		return res, err
	}
	return res, nil
}

// ListContainersOptions groups optional parameters of [Client.ListContainers].
type ListContainersOptions struct{}

// ListContainers requests identifiers of all user-owned containers.
func (c *Client) ListContainers(ctx context.Context, usr user.ID, _ ListContainersOptions) ([]cid.ID, error) {
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerList, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apicontainer.ListRequest{
		Body: &apicontainer.ListRequest_Body{OwnerId: new(refs.OwnerID)},
	}
	usr.WriteToV2(req.Body.OwnerId)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return nil, err
	}

	// send request
	resp, err := c.transport.container.List(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return nil, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return nil, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return nil, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return nil, err
	}
	if sts != nil {
		err = sts // for closure above
		return nil, err
	}

	// decode response payload
	var res []cid.ID
	if resp.Body != nil && len(resp.Body.ContainerIds) > 0 {
		const fieldIDs = "ID list"
		res = make([]cid.ID, len(resp.Body.ContainerIds))
		for i := range resp.Body.ContainerIds {
			if resp.Body.ContainerIds == nil {
				err = fmt.Errorf("%s (%s): nil element #%d", errInvalidResponseBodyField, fieldIDs, i) // for closure above
				return nil, err
			} else if err = res[i].ReadFromV2(resp.Body.ContainerIds[i]); err != nil {
				err = fmt.Errorf("%s (%s): invalid element #%d: %w", errInvalidResponseBodyField, fieldIDs, i, err) // for closure above
				return nil, err
			}
		}
	}
	return res, nil
}

// DeleteContainerOptions groups optional parameters of
// [Client.DeleteContainer].
type DeleteContainerOptions struct {
	sessionSet bool
	session    session.Container
}

// WithinSession specifies session within which container should be removed.
// Session tokens grant user-to-user power of attorney: the subject can remove
// specified issuer's containers. Session op must be
// [session.VerbContainerDelete]. If used, [Client.DeleteContainer] default
// ownership behavior is replaced with:
//   - session must be issued by the container owner;
//   - session must target the subject authenticated by signer passed to
//     [Client.DeleteContainer].
func (x *DeleteContainerOptions) WithinSession(tok session.Container) {
	x.session = tok
	x.sessionSet = true
}

// DeleteContainer sends request to remove the NeoFS container. If the request
// is accepted, DeleteContainer returns no error and the container is going to
// be removed asynchronously. The completion can be checked by polling the
// container presence until it won't be found (e.g. call [Client.GetContainer]
// until [apistatus.ErrContainerNotFound]).
//
// Signer must authenticate container's owner. The signature scheme MUST be
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] (e.g. [neofsecdsa.SignerRFC6979]).
// Corresponding NeoFS account can be charged for the operation.
func (c *Client) DeleteContainer(ctx context.Context, id cid.ID, signer neofscrypto.Signer, opts DeleteContainerOptions) error {
	if signer == nil {
		return errMissingSigner
	} else if scheme := signer.Scheme(); scheme != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return fmt.Errorf("wrong signature scheme: %v instead of %v", scheme, neofscrypto.ECDSA_DETERMINISTIC_SHA256)
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerDelete, time.Since(start), err)
		}(time.Now())
	}

	sig, err := signer.Sign(id[:])
	if err != nil {
		err = fmt.Errorf("sign container ID: %w", err) // for closure above
		return err
	}

	// form request
	req := &apicontainer.DeleteRequest{
		Body: &apicontainer.DeleteRequest_Body{
			ContainerId: new(refs.ContainerID),
			Signature:   &refs.SignatureRFC6979{Key: neofscrypto.PublicKeyBytes(signer.Public()), Sign: sig},
		},
	}
	id.WriteToV2(req.Body.ContainerId)
	if opts.sessionSet {
		req.MetaHeader = &apisession.RequestMetaHeader{SessionToken: new(apisession.SessionToken)}
		opts.session.WriteToV2(req.MetaHeader.SessionToken)
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return err
	}

	// send request
	resp, err := c.transport.container.Delete(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
	} else if sts != nil {
		err = sts // for closure above
	}
	return err
}

// GetEACLOptions groups optional parameters of [Client.GetEACL].
type GetEACLOptions struct{}

// GetEACL reads eACL table of the NeoFS container. Returns
// [apistatus.ErrEACLNotFound] if eACL is unset for this container. Returns
// [apistatus.ErrContainerNotFound] if there is no such container.
func (c *Client) GetEACL(ctx context.Context, id cid.ID, _ GetEACLOptions) (eacl.Table, error) {
	var res eacl.Table
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerEACL, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apicontainer.GetExtendedACLRequest{
		Body: &apicontainer.GetExtendedACLRequest_Body{ContainerId: new(refs.ContainerID)},
	}
	id.WriteToV2(req.Body.ContainerId)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.container.GetExtendedACL(ctx, req)
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
		err = sts
		return res, err // for closure above
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldEACL = "eACL"
	if resp.Body.Eacl == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldEACL) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.Eacl); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldEACL, err) // for closure above
		return res, err
	}
	return res, nil
}

// SetEACLOptions groups optional parameters of [Client.SetEACLOptions].
type SetEACLOptions struct {
	sessionSet bool
	session    session.Container
}

// WithinSession specifies session within which extended ACL of the container
// should be saved. Session tokens grant user-to-user power of attorney: the
// subject can modify eACL rules of specified issuer's containers. Session op
// must be [session.VerbContainerSetEACL]. If used, [Client.SetEACL] default
// ownership behavior is replaced with:
//   - session must be issued by the container owner;
//   - session must target the subject authenticated by signer passed to
//     [Client.SetEACL].
func (x *SetEACLOptions) WithinSession(s session.Container) {
	x.session = s
	x.sessionSet = true
}

// SetEACL sends request to update eACL table of the NeoFS container. If the
// request is accepted, SetEACL returns no error and the eACL is going to be set
// asynchronously. The completion can be checked by eACL polling
// ([Client.GetEACL]) and binary comparison. SetEACL returns
// [apistatus.ErrContainerNotFound] if container is missing.
//
// Signer must authenticate container's owner. The signature scheme MUST be
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] (e.g. [neofsecdsa.SignerRFC6979]).
// Owner's NeoFS account can be charged for the operation.
func (c *Client) SetEACL(ctx context.Context, eACL eacl.Table, signer neofscrypto.Signer, opts SetEACLOptions) error {
	if signer == nil {
		return errMissingSigner
	} else if scheme := signer.Scheme(); scheme != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return fmt.Errorf("wrong signature scheme: %v instead of %v", scheme, neofscrypto.ECDSA_DETERMINISTIC_SHA256)
	} else if eACL.LimitedContainer().IsZero() {
		return errors.New("missing container in the eACL")
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerSetEACL, time.Since(start), err)
		}(time.Now())
	}

	sig, err := signer.Sign(eACL.Marshal())
	if err != nil {
		err = fmt.Errorf("sign eACL: %w", err) // for closure above
		return err
	}

	// form request
	req := &apicontainer.SetExtendedACLRequest{
		Body: &apicontainer.SetExtendedACLRequest_Body{
			Eacl:      new(apiacl.EACLTable),
			Signature: &refs.SignatureRFC6979{Key: neofscrypto.PublicKeyBytes(signer.Public()), Sign: sig},
		},
	}
	eACL.WriteToV2(req.Body.Eacl)
	if opts.sessionSet {
		req.MetaHeader = &apisession.RequestMetaHeader{SessionToken: new(apisession.SessionToken)}
		opts.session.WriteToV2(req.MetaHeader.SessionToken)
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return err
	}

	// send request
	resp, err := c.transport.container.SetExtendedACL(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
	} else {
		err = sts // for closure above
	}
	return err
}

// SendContainerSizeEstimationsOptions groups optional parameters of
// [Client.SendContainerSizeEstimations].
type SendContainerSizeEstimationsOptions struct{}

// SendContainerSizeEstimations sends container size estimations to the remote
// node. The estimation set must not be empty.
//
// SendContainerSizeEstimations is used for system needs and is not intended to
// be called by regular users.
func (c *Client) SendContainerSizeEstimations(ctx context.Context, es []container.SizeEstimation, _ SendContainerSizeEstimationsOptions) error {
	if len(es) == 0 {
		return errors.New("missing estimations")
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodContainerAnnounceUsedSpace, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apicontainer.AnnounceUsedSpaceRequest{
		Body: &apicontainer.AnnounceUsedSpaceRequest_Body{
			Announcements: make([]*apicontainer.AnnounceUsedSpaceRequest_Body_Announcement, len(es)),
		},
	}
	for i := range es {
		req.Body.Announcements[i] = new(apicontainer.AnnounceUsedSpaceRequest_Body_Announcement)
		es[i].WriteToV2(req.Body.Announcements[i])
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return err
	}

	// send request
	resp, err := c.transport.container.AnnounceUsedSpace(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
	}
	if sts != nil {
		err = sts // for closure above
	}
	return err
}
