package client

import (
	"context"

	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

// ContainerPutPrm groups parameters of PutContainer operation.
type ContainerPutPrm struct {
	prmSession

	cnrSet bool
	cnr    container.Container
}

// SetContainer sets structured information about new NeoFS container.
// Required parameter.
func (x *ContainerPutPrm) SetContainer(cnr container.Container) {
	x.cnr = cnr
	x.cnrSet = true
}

// ContainerPutRes groups resulting values of PutContainer operation.
type ContainerPutRes struct {
	statusRes

	id *cid.ID
}

// ID returns identifier of the container declared to be stored in the system.
// Used as a link to information about the container (in particular, you can
// asynchronously check if the save was successful).
//
// Client doesn't retain value so modification is safe.
func (x ContainerPutRes) ID() *cid.ID {
	return x.id
}

func (x *ContainerPutRes) setID(id *cid.ID) {
	x.id = id
}

// PutContainer sends request to save container in NeoFS.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see ContainerPutRes.ID).
//
// Immediately panics if parameters are set incorrectly (see ContainerPutPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) PutContainer(ctx context.Context, prm ContainerPutPrm) (*ContainerPutRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.cnrSet:
		panic(panicMsgMissingContainer)
	}

	// TODO: check private key is set before forming the request

	// form request body
	reqBody := new(v2container.PutRequestBody)
	reqBody.SetContainer(prm.cnr.ToV2())

	// sign container
	signWrapper := v2signature.StableMarshalerWrapper{SM: reqBody.GetContainer()}

	err := sigutil.SignDataWithHandler(c.opts.key, signWrapper, func(key []byte, sig []byte) {
		containerSignature := new(refs.Signature)
		containerSignature.SetKey(key)
		containerSignature.SetSign(sig)
		reqBody.SetSignature(containerSignature)
	}, sigutil.SignWithRFC6979())
	if err != nil {
		return nil, err
	}

	// form meta header
	var meta v2session.RequestMetaHeader

	prm.prmSession.writeToMetaHeader(&meta)

	// form request
	var req v2container.PutRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res ContainerPutRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.PutContainer(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.PutResponse)
		res.setID(cid.NewFromV2(resp.GetBody().GetContainerID()))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// ContainerGetPrm groups parameters of GetContainer operation.
type ContainerGetPrm struct {
	idSet bool
	id    cid.ID
}

// SetContainer sets identifier of the container to be read.
// Required parameter.
func (x *ContainerGetPrm) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
}

// ContainerGetRes groups resulting values of GetContainer operation.
type ContainerGetRes struct {
	statusRes

	cnr *container.Container
}

// Container returns structured information about the requested container.
//
// Client doesn't retain value so modification is safe.
func (x ContainerGetRes) Container() *container.Container {
	return x.cnr
}

func (x *ContainerGetRes) setContainer(cnr *container.Container) {
	x.cnr = cnr
}

// GetContainer reads NeoFS container by ID.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see ContainerGetPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) GetContainer(ctx context.Context, prm ContainerGetPrm) (*ContainerGetRes, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	// form request body
	reqBody := new(v2container.GetRequestBody)
	reqBody.SetContainerID(prm.id.ToV2())

	// form request
	var req v2container.GetRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ContainerGetRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.GetContainer(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetResponse)

		body := resp.GetBody()

		cnr := container.NewContainerFromV2(body.GetContainer())

		cnr.SetSessionToken(
			session.NewTokenFromV2(body.GetSessionToken()),
		)

		cnr.SetSignature(
			signature.NewFromV2(body.GetSignature()),
		)

		res.setContainer(cnr)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// ContainerListPrm groups parameters of ListContainers operation.
type ContainerListPrm struct {
	ownerSet bool
	ownerID  owner.ID
}

// SetAccount sets identifier of the NeoFS account to list the containers.
// Required parameter. Must be a valid ID according to NeoFS API protocol.
func (x *ContainerListPrm) SetAccount(id owner.ID) {
	x.ownerID = id
	x.ownerSet = true
}

// ContainerListRes groups resulting values of ListContainers operation.
type ContainerListRes struct {
	statusRes

	ids []*cid.ID
}

// Containers returns list of identifiers of the account-owned containers.
//
// Client doesn't retain value so modification is safe.
func (x ContainerListRes) Containers() []*cid.ID {
	return x.ids
}

func (x *ContainerListRes) setContainers(ids []*cid.ID) {
	x.ids = ids
}

// ListContainers requests identifiers of the account-owned containers.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see ContainerListPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ListContainers(ctx context.Context, prm ContainerListPrm) (*ContainerListRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.ownerSet:
		panic("account not set")
	case !prm.ownerID.Valid():
		panic("invalid account")
	}

	// form request body
	reqBody := new(v2container.ListRequestBody)
	reqBody.SetOwnerID(prm.ownerID.ToV2())

	// form request
	var req v2container.ListRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ContainerListRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.ListContainers(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.ListResponse)

		ids := make([]*cid.ID, 0, len(resp.GetBody().GetContainerIDs()))

		for _, cidV2 := range resp.GetBody().GetContainerIDs() {
			ids = append(ids, cid.NewFromV2(cidV2))
		}

		res.setContainers(ids)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// ContainerDeletePrm groups parameters of DeleteContainer operation.
type ContainerDeletePrm struct {
	prmSession

	idSet bool
	id    cid.ID
}

// SetContainer sets identifier of the NeoFS container to be removed.
// Required parameter.
func (x *ContainerDeletePrm) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
}

// ContainerDeleteRes groups resulting values of DeleteContainer operation.
type ContainerDeleteRes struct {
	statusRes
}

// implements github.com/nspcc-dev/neofs-sdk-go/util/signature.DataSource.
type delContainerSignWrapper struct {
	body *v2container.DeleteRequestBody
}

func (c delContainerSignWrapper) ReadSignedData([]byte) ([]byte, error) {
	return c.body.GetContainerID().GetValue(), nil
}

func (c delContainerSignWrapper) SignedDataSize() int {
	return len(c.body.GetContainerID().GetValue())
}

// DeleteContainer sends request to remove the NeoFS container.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see GetContainer).
//
// Immediately panics if parameters are set incorrectly (see ContainerDeletePrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in ContainerDeleteRes.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) DeleteContainer(ctx context.Context, prm ContainerDeletePrm) (*ContainerDeleteRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	// form request body
	reqBody := new(v2container.DeleteRequestBody)
	reqBody.SetContainerID(prm.id.ToV2())

	// sign container
	err := sigutil.SignDataWithHandler(c.opts.key,
		delContainerSignWrapper{
			body: reqBody,
		},
		func(key []byte, sig []byte) {
			containerSignature := new(refs.Signature)
			containerSignature.SetKey(key)
			containerSignature.SetSign(sig)
			reqBody.SetSignature(containerSignature)
		},
		sigutil.SignWithRFC6979())
	if err != nil {
		return nil, err
	}

	// form meta header
	var meta v2session.RequestMetaHeader

	prm.prmSession.writeToMetaHeader(&meta)

	// form request
	var req v2container.DeleteRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res ContainerDeleteRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.DeleteContainer(c.Raw(), &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// EACLPrm groups parameters of EACL operation.
type EACLPrm struct {
	idSet bool
	id    cid.ID
}

// SetContainer sets identifier of the NeoFS container to read the eACL table.
// Required parameter.
func (x *EACLPrm) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
}

// EACLRes groups resulting values of EACL operation.
type EACLRes struct {
	statusRes

	table *eacl.Table
}

// Table returns eACL table of the requested container.
//
// Client doesn't retain value so modification is safe.
func (x EACLRes) Table() *eacl.Table {
	return x.table
}

func (x *EACLRes) setTable(table *eacl.Table) {
	x.table = table
}

// EACL reads eACL table of the NeoFS container.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see EACLPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) EACL(ctx context.Context, prm EACLPrm) (*EACLRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	// form request body
	reqBody := new(v2container.GetExtendedACLRequestBody)
	reqBody.SetContainerID(prm.id.ToV2())

	// form request
	var req v2container.GetExtendedACLRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res EACLRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.GetEACL(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetExtendedACLResponse)

		body := resp.GetBody()

		table := eacl.NewTableFromV2(body.GetEACL())

		table.SetSessionToken(
			session.NewTokenFromV2(body.GetSessionToken()),
		)

		table.SetSignature(
			signature.NewFromV2(body.GetSignature()),
		)

		res.setTable(table)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// SetEACLPrm groups parameters of SetEACL operation.
type SetEACLPrm struct {
	prmSession

	tableSet bool
	table    eacl.Table
}

// SetTable sets eACL table structure to be set for the container.
// Required parameter.
func (x *SetEACLPrm) SetTable(table eacl.Table) {
	x.table = table
	x.tableSet = true
}

// SetEACLRes groups resulting values of SetEACL operation.
type SetEACLRes struct {
	statusRes
}

// SetEACL sends request to update eACL table of the NeoFS container.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see EACL).
//
// Immediately panics if parameters are set incorrectly (see SetEACLPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) SetEACL(ctx context.Context, prm SetEACLPrm) (*SetEACLRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.tableSet:
		panic("eACL table not set")
	}

	// form request body
	reqBody := new(v2container.SetExtendedACLRequestBody)
	reqBody.SetEACL(prm.table.ToV2())

	// sign the eACL table
	signWrapper := v2signature.StableMarshalerWrapper{SM: reqBody.GetEACL()}

	err := sigutil.SignDataWithHandler(c.opts.key, signWrapper, func(key []byte, sig []byte) {
		eaclSignature := new(refs.Signature)
		eaclSignature.SetKey(key)
		eaclSignature.SetSign(sig)
		reqBody.SetSignature(eaclSignature)
	}, sigutil.SignWithRFC6979())
	if err != nil {
		return nil, err
	}

	// form meta header
	var meta v2session.RequestMetaHeader

	prm.prmSession.writeToMetaHeader(&meta)

	// form request
	var req v2container.SetExtendedACLRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res SetEACLRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.SetEACL(c.Raw(), &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// AnnounceSpacePrm groups parameters of AnnounceContainerUsedSpace operation.
type AnnounceSpacePrm struct {
	announcements []container.UsedSpaceAnnouncement
}

// SetValues sets values describing volume of space that is used for the container objects.
// Required parameter. Must not be empty.
//
// Must not be mutated before the end of the operation.
func (x *AnnounceSpacePrm) SetValues(announcements []container.UsedSpaceAnnouncement) {
	x.announcements = announcements
}

// AnnounceSpaceRes groups resulting values of AnnounceContainerUsedSpace operation.
type AnnounceSpaceRes struct {
	statusRes
}

// AnnounceContainerUsedSpace sends request to announce volume of the space used for the container objects.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// At this moment success can not be checked.
//
// Immediately panics if parameters are set incorrectly (see AnnounceSpacePrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) AnnounceContainerUsedSpace(ctx context.Context, prm AnnounceSpacePrm) (*AnnounceSpaceRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case len(prm.announcements) == 0:
		panic("missing announcements")
	}

	// convert list of SDK announcement structures into NeoFS-API v2 list
	v2announce := make([]*v2container.UsedSpaceAnnouncement, 0, len(prm.announcements))
	for i := range prm.announcements {
		v2announce = append(v2announce, prm.announcements[i].ToV2())
	}

	// prepare body of the NeoFS-API v2 request and request itself
	reqBody := new(v2container.AnnounceUsedSpaceRequestBody)
	reqBody.SetAnnouncements(v2announce)

	// form request
	var req v2container.AnnounceUsedSpaceRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res AnnounceSpaceRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceUsedSpace(c.Raw(), &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
