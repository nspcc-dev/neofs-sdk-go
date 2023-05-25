package client

import (
	"context"
	"errors"
	"fmt"

	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmContainerPut groups optional parameters of ContainerPut operation.
type PrmContainerPut struct {
	prmCommonMeta

	sessionSet bool
	session    session.Container

	signer neofscrypto.Signer
}

// SetSigner sets signer to sign request payload.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// Optional parameter: defaults to internal Client signer.
func (x *PrmContainerPut) SetSigner(signer neofscrypto.Signer) {
	x.signer = signer
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

// ResContainerPut groups resulting values of ContainerPut operation.
type ResContainerPut struct {
	id cid.ID
}

// ID returns identifier of the container declared to be stored in the system.
// Used as a link to information about the container (in particular, you can
// asynchronously check if the save was successful).
func (x ResContainerPut) ID() cid.ID {
	return x.id
}

// ContainerPut sends request to save container in NeoFS.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see ResContainerPut.ID).
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ContainerPut(ctx context.Context, cont container.Container, prm PrmContainerPut) (*ResContainerPut, error) {
	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
	}

	var cnr v2container.Container
	cont.WriteToV2(&cnr)

	// TODO: check private signer is set before forming the request
	var sig neofscrypto.Signature
	err = container.CalculateSignature(&sig, cont, signer)
	if err != nil {
		return nil, fmt.Errorf("calculate container signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.PutRequestBody)
	reqBody.SetContainer(&cnr)
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	writeXHeadersToMeta(prm.prmCommonMeta.xHeaders, &meta)

	if prm.sessionSet {
		var tokv2 v2session.Token
		prm.session.WriteToV2(&tokv2)

		meta.SetSessionToken(&tokv2)
	}

	// form request
	var req v2container.PutRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res ResContainerPut
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.PutContainer(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.PutResponse)

		const fieldCnrID = "container ID"

		cidV2 := resp.GetBody().GetContainerID()
		if cidV2 == nil {
			cc.err = newErrMissingResponseField(fieldCnrID)
			return
		}

		cc.err = res.id.ReadFromV2(*cidV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldCnrID, cc.err)
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerGet groups optional parameters of ContainerGet operation.
type PrmContainerGet struct {
	prmCommonMeta
}

// ContainerGet reads NeoFS container by ID.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ContainerGet(ctx context.Context, id cid.ID, prm PrmContainerGet) (container.Container, error) {
	if c.prm.signer == nil {
		return container.Container{}, ErrMissingSigner
	}

	var cidV2 refs.ContainerID
	id.WriteToV2(&cidV2)

	// form request body
	reqBody := new(v2container.GetRequestBody)
	reqBody.SetContainerID(&cidV2)

	// form request
	var req v2container.GetRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res container.Container
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.GetContainer(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetResponse)

		cnrV2 := resp.GetBody().GetContainer()
		if cnrV2 == nil {
			cc.err = errors.New("missing container in response")
			return
		}

		cc.err = res.ReadFromV2(*cnrV2)
		if cc.err != nil {
			cc.err = fmt.Errorf("invalid container in response: %w", cc.err)
		}
	}

	// process call
	if !cc.processCall() {
		return container.Container{}, cc.err
	}

	return res, nil
}

// PrmContainerList groups optional parameters of ContainerList operation.
type PrmContainerList struct {
	prmCommonMeta
}

// ResContainerList groups resulting values of ContainerList operation.
type ResContainerList struct {
	ids []cid.ID
}

// Containers returns list of identifiers of the account-owned containers.
//
// Client doesn't retain value so modification is safe.
func (x ResContainerList) Containers() []cid.ID {
	return x.ids
}

// ContainerList requests identifiers of the account-owned containers.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ContainerList(ctx context.Context, ownerID user.ID, prm PrmContainerList) (*ResContainerList, error) {
	if c.prm.signer == nil {
		return nil, ErrMissingSigner
	}

	// form request body
	var ownerV2 refs.OwnerID
	ownerID.WriteToV2(&ownerV2)

	reqBody := new(v2container.ListRequestBody)
	reqBody.SetOwnerID(&ownerV2)

	// form request
	var req v2container.ListRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ResContainerList
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.ListContainers(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.ListResponse)

		res.ids = make([]cid.ID, len(resp.GetBody().GetContainerIDs()))

		for i, cidV2 := range resp.GetBody().GetContainerIDs() {
			cc.err = res.ids[i].ReadFromV2(cidV2)
			if cc.err != nil {
				cc.err = fmt.Errorf("invalid ID in the response: %w", cc.err)
				return
			}
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerDelete groups optional parameters of ContainerDelete operation.
type PrmContainerDelete struct {
	prmCommonMeta

	tokSet bool
	tok    session.Container

	signer neofscrypto.Signer
}

// SetSigner sets signer to sign request payload.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// Optional parameter: defaults to internal Client signer.
func (x *PrmContainerDelete) SetSigner(signer neofscrypto.Signer) {
	x.signer = signer
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
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
// Return errors:
//   - [ErrMissingSigner]
//   - [neofscrypto.ErrIncorrectSigner]
func (c *Client) ContainerDelete(ctx context.Context, id cid.ID, prm PrmContainerDelete) error {
	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return err
	}

	if signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return errNonNeoSigner
	}

	// sign container ID
	var cidV2 refs.ContainerID
	id.WriteToV2(&cidV2)

	// container contract expects signature of container ID value
	// don't get confused with stable marshaled protobuf container.ID structure
	data := cidV2.GetValue()

	var sig neofscrypto.Signature
	err = sig.Calculate(signer, data)
	if err != nil {
		return fmt.Errorf("calculate signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.DeleteRequestBody)
	reqBody.SetContainerID(&cidV2)
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	writeXHeadersToMeta(prm.prmCommonMeta.xHeaders, &meta)

	if prm.tokSet {
		var tokv2 v2session.Token
		prm.tok.WriteToV2(&tokv2)

		meta.SetSessionToken(&tokv2)
	}

	// form request
	var req v2container.DeleteRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc contextCall
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.DeleteContainer(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return cc.err
	}

	return nil
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
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ContainerEACL(ctx context.Context, id cid.ID, prm PrmContainerEACL) (eacl.Table, error) {
	if c.prm.signer == nil {
		return eacl.Table{}, ErrMissingSigner
	}

	var cidV2 refs.ContainerID
	id.WriteToV2(&cidV2)

	// form request body
	reqBody := new(v2container.GetExtendedACLRequestBody)
	reqBody.SetContainerID(&cidV2)

	// form request
	var req v2container.GetExtendedACLRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res eacl.Table
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.GetEACL(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetExtendedACLResponse)

		eACL := resp.GetBody().GetEACL()
		if eACL == nil {
			cc.err = newErrMissingResponseField("eACL")
			return
		}

		res = *eacl.NewTableFromV2(eACL)
	}

	// process call
	if !cc.processCall() {
		return eacl.Table{}, cc.err
	}

	return res, nil
}

// PrmContainerSetEACL groups optional parameters of ContainerSetEACL operation.
type PrmContainerSetEACL struct {
	prmCommonMeta

	sessionSet bool
	session    session.Container

	signer neofscrypto.Signer
}

// SetSigner sets signer to sign request payload.
// Signer's scheme MUST be neofscrypto.ECDSA_DETERMINISTIC_SHA256. For example, you can use neofsecdsa.SignerRFC6979.
// Optional parameter: defaults to internal Client signer.
func (x *PrmContainerSetEACL) SetSigner(signer neofscrypto.Signer) {
	x.signer = signer
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
// Return errors:
//   - [ErrMissingEACLContainer]
//   - [neofscrypto.ErrIncorrectSigner]
//   - [ErrMissingSigner]
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ContainerSetEACL(ctx context.Context, table eacl.Table, prm PrmContainerSetEACL) error {
	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return err
	}

	if signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return errNonNeoSigner
	}

	_, isCIDSet := table.CID()
	if !isCIDSet {
		return ErrMissingEACLContainer
	}

	// sign the eACL table
	eaclV2 := table.ToV2()

	var sig neofscrypto.Signature
	err = sig.Calculate(signer, eaclV2.StableMarshal(nil))
	if err != nil {
		return fmt.Errorf("calculate signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.SetExtendedACLRequestBody)
	reqBody.SetEACL(eaclV2)
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	writeXHeadersToMeta(prm.prmCommonMeta.xHeaders, &meta)

	if prm.sessionSet {
		var tokv2 v2session.Token
		prm.session.WriteToV2(&tokv2)

		meta.SetSessionToken(&tokv2)
	}

	// form request
	var req v2container.SetExtendedACLRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc contextCall
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.SetEACL(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return cc.err
	}

	return nil
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
//   - [ErrMissingSigner]
func (c *Client) ContainerAnnounceUsedSpace(ctx context.Context, announcements []container.SizeEstimation, prm PrmAnnounceSpace) error {
	// check parameters

	if len(announcements) == 0 {
		return ErrMissingAnnouncements
	}

	if c.prm.signer == nil {
		return ErrMissingSigner
	}

	// convert list of SDK announcement structures into NeoFS-API v2 list
	v2announce := make([]v2container.UsedSpaceAnnouncement, len(announcements))
	for i := range announcements {
		announcements[i].WriteToV2(&v2announce[i])
	}

	// prepare body of the NeoFS-API v2 request and request itself
	reqBody := new(v2container.AnnounceUsedSpaceRequestBody)
	reqBody.SetAnnouncements(v2announce)

	// form request
	var req v2container.AnnounceUsedSpaceRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc contextCall
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceUsedSpace(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return cc.err
	}

	return nil
}

// SyncContainerWithNetwork requests network configuration using passed client
// and applies it to the container. Container MUST not be nil.
//
// Note: if container does not match network configuration, SyncContainerWithNetwork
// changes it.
//
// Returns any network/parsing config errors.
//
// See also NetworkInfo, container.ApplyNetworkConfig.
func SyncContainerWithNetwork(ctx context.Context, cnr *container.Container, c *Client) error {
	res, err := c.NetworkInfo(ctx, PrmNetworkInfo{})
	if err != nil {
		return fmt.Errorf("network info call: %w", err)
	}

	container.ApplyNetworkConfig(cnr, res.Info())

	return nil
}
