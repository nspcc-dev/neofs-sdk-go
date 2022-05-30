package client

import (
	"context"
	"fmt"

	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmContainerPut groups parameters of ContainerPut operation.
type PrmContainerPut struct {
	prmCommonMeta

	cnrSet bool
	cnr    container.Container
}

// SetContainer sets structured information about new NeoFS container.
// Required parameter.
func (x *PrmContainerPut) SetContainer(cnr container.Container) {
	x.cnr = cnr
	x.cnrSet = true
}

// ResContainerPut groups resulting values of ContainerPut operation.
type ResContainerPut struct {
	statusRes

	id *cid.ID
}

// ID returns identifier of the container declared to be stored in the system.
// Used as a link to information about the container (in particular, you can
// asynchronously check if the save was successful).
//
// Client doesn't retain value so modification is safe.
func (x ResContainerPut) ID() *cid.ID {
	return x.id
}

func (x *ResContainerPut) setID(id *cid.ID) {
	x.id = id
}

// ContainerPut sends request to save container in NeoFS.
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
// Success can be verified by reading by identifier (see ResContainerPut.ID).
//
// Immediately panics if parameters are set incorrectly (see PrmContainerPut docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ContainerPut(ctx context.Context, prm PrmContainerPut) (*ResContainerPut, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.cnrSet:
		panic(panicMsgMissingContainer)
	}

	// TODO: check private key is set before forming the request
	// sign container
	cnr := prm.cnr.ToV2()

	var sig neofscrypto.Signature

	err := sig.Calculate(neofsecdsa.SignerRFC6979(c.prm.key), cnr.StableMarshal(nil))
	if err != nil {
		return nil, fmt.Errorf("calculate signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.PutRequestBody)
	reqBody.SetContainer(prm.cnr.ToV2())
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	prm.prmCommonMeta.writeToMetaHeader(&meta)

	if tok := prm.cnr.SessionToken(); tok != nil {
		var tokv2 v2session.Token
		tok.WriteToV2(&tokv2)

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
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.PutContainer(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.PutResponse)
		var cID *cid.ID

		cidV2 := resp.GetBody().GetContainerID()
		if cidV2 != nil {
			var c cid.ID
			_ = c.ReadFromV2(*cidV2)

			cID = &c
		}

		res.setID(cID)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerGet groups parameters of ContainerGet operation.
type PrmContainerGet struct {
	prmCommonMeta

	idSet bool
	id    cid.ID
}

// SetContainer sets identifier of the container to be read.
// Required parameter.
func (x *PrmContainerGet) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
}

// ResContainerGet groups resulting values of ContainerGet operation.
type ResContainerGet struct {
	statusRes

	cnr *container.Container
}

// Container returns structured information about the requested container.
//
// Client doesn't retain value so modification is safe.
func (x ResContainerGet) Container() *container.Container {
	return x.cnr
}

func (x *ResContainerGet) setContainer(cnr *container.Container) {
	x.cnr = cnr
}

// ContainerGet reads NeoFS container by ID.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmContainerGet docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound.
func (c *Client) ContainerGet(ctx context.Context, prm PrmContainerGet) (*ResContainerGet, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	var cidV2 refs.ContainerID
	prm.id.WriteToV2(&cidV2)

	// form request body
	reqBody := new(v2container.GetRequestBody)
	reqBody.SetContainerID(&cidV2)

	// form request
	var req v2container.GetRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ResContainerGet
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.GetContainer(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetResponse)

		body := resp.GetBody()

		cnr := container.NewContainerFromV2(body.GetContainer())

		tokv2 := body.GetSessionToken()
		if tokv2 != nil {
			var tok session.Container

			// FIXME: (neofs-sdk-go#221) need to handle the error
			err := tok.ReadFromV2(*tokv2)
			if err == nil {
				cnr.SetSessionToken(&tok)
			}
		}

		var sig *neofscrypto.Signature

		if sigv2 := body.GetSignature(); sigv2 != nil {
			sig = new(neofscrypto.Signature)
			sig.ReadFromV2(*sigv2)
		}

		cnr.SetSignature(sig)

		res.setContainer(cnr)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerList groups parameters of ContainerList operation.
type PrmContainerList struct {
	prmCommonMeta

	ownerSet bool
	ownerID  user.ID
}

// SetAccount sets identifier of the NeoFS account to list the containers.
// Required parameter.
func (x *PrmContainerList) SetAccount(id user.ID) {
	x.ownerID = id
	x.ownerSet = true
}

// ResContainerList groups resulting values of ContainerList operation.
type ResContainerList struct {
	statusRes

	ids []cid.ID
}

// Containers returns list of identifiers of the account-owned containers.
//
// Client doesn't retain value so modification is safe.
func (x ResContainerList) Containers() []cid.ID {
	return x.ids
}

func (x *ResContainerList) setContainers(ids []cid.ID) {
	x.ids = ids
}

// ContainerList requests identifiers of the account-owned containers.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmContainerList docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ContainerList(ctx context.Context, prm PrmContainerList) (*ResContainerList, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.ownerSet:
		panic("account not set")
	}

	// form request body
	var ownerV2 refs.OwnerID
	prm.ownerID.WriteToV2(&ownerV2)

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
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.ListContainers(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.ListResponse)

		ids := make([]cid.ID, len(resp.GetBody().GetContainerIDs()))

		for i, cidV2 := range resp.GetBody().GetContainerIDs() {
			_ = ids[i].ReadFromV2(cidV2)
		}

		res.setContainers(ids)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerDelete groups parameters of ContainerDelete operation.
type PrmContainerDelete struct {
	prmCommonMeta

	idSet bool
	id    cid.ID

	tokSet bool
	tok    session.Container
}

// SetContainer sets identifier of the NeoFS container to be removed.
// Required parameter.
func (x *PrmContainerDelete) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
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

// ResContainerDelete groups resulting values of ContainerDelete operation.
type ResContainerDelete struct {
	statusRes
}

// ContainerDelete sends request to remove the NeoFS container.
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
// Immediately panics if parameters are set incorrectly (see PrmContainerDelete docs).
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in ResContainerDelete.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ContainerDelete(ctx context.Context, prm PrmContainerDelete) (*ResContainerDelete, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	// sign container ID
	var cidV2 refs.ContainerID
	prm.id.WriteToV2(&cidV2)

	// container contract expects signature of container ID value
	// don't get confused with stable marshaled protobuf container.ID structure
	data := cidV2.GetValue()

	var sig neofscrypto.Signature

	err := sig.Calculate(neofsecdsa.SignerRFC6979(c.prm.key), data)
	if err != nil {
		return nil, fmt.Errorf("calculate signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.DeleteRequestBody)
	reqBody.SetContainerID(&cidV2)
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	prm.prmCommonMeta.writeToMetaHeader(&meta)

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
		cc  contextCall
		res ResContainerDelete
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.DeleteContainer(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerEACL groups parameters of ContainerEACL operation.
type PrmContainerEACL struct {
	prmCommonMeta

	idSet bool
	id    cid.ID
}

// SetContainer sets identifier of the NeoFS container to read the eACL table.
// Required parameter.
func (x *PrmContainerEACL) SetContainer(id cid.ID) {
	x.id = id
	x.idSet = true
}

// ResContainerEACL groups resulting values of ContainerEACL operation.
type ResContainerEACL struct {
	statusRes

	table *eacl.Table
}

// Table returns eACL table of the requested container.
//
// Client doesn't retain value so modification is safe.
func (x ResContainerEACL) Table() *eacl.Table {
	return x.table
}

func (x *ResContainerEACL) setTable(table *eacl.Table) {
	x.table = table
}

// ContainerEACL reads eACL table of the NeoFS container.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmContainerEACL docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound.
func (c *Client) ContainerEACL(ctx context.Context, prm PrmContainerEACL) (*ResContainerEACL, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.idSet:
		panic(panicMsgMissingContainer)
	}

	var cidV2 refs.ContainerID
	prm.id.WriteToV2(&cidV2)

	// form request body
	reqBody := new(v2container.GetExtendedACLRequestBody)
	reqBody.SetContainerID(&cidV2)

	// form request
	var req v2container.GetExtendedACLRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res ResContainerEACL
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.GetEACL(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2container.GetExtendedACLResponse)

		body := resp.GetBody()

		table := eacl.NewTableFromV2(body.GetEACL())

		tokv2 := body.GetSessionToken()
		if tokv2 != nil {
			var tok session.Container

			// FIXME: (neofs-sdk-go#221) need to handle the error
			err := tok.ReadFromV2(*tokv2)
			if err == nil {
				table.SetSessionToken(&tok)
			}
		}

		var sig *neofscrypto.Signature

		if sigv2 := body.GetSignature(); sigv2 != nil {
			sig = new(neofscrypto.Signature)
			sig.ReadFromV2(*sigv2)
		}

		table.SetSignature(sig)

		res.setTable(table)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmContainerSetEACL groups parameters of ContainerSetEACL operation.
type PrmContainerSetEACL struct {
	prmCommonMeta

	tableSet bool
	table    eacl.Table
}

// SetTable sets eACL table structure to be set for the container.
// Required parameter.
func (x *PrmContainerSetEACL) SetTable(table eacl.Table) {
	x.table = table
	x.tableSet = true
}

// ResContainerSetEACL groups resulting values of ContainerSetEACL operation.
type ResContainerSetEACL struct {
	statusRes
}

// ContainerSetEACL sends request to update eACL table of the NeoFS container.
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
// Immediately panics if parameters are set incorrectly (see PrmContainerSetEACL docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ContainerSetEACL(ctx context.Context, prm PrmContainerSetEACL) (*ResContainerSetEACL, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.tableSet:
		panic("eACL table not set")
	}

	// sign the eACL table
	eaclV2 := prm.table.ToV2()

	var sig neofscrypto.Signature

	err := sig.Calculate(neofsecdsa.SignerRFC6979(c.prm.key), eaclV2.StableMarshal(nil))
	if err != nil {
		return nil, fmt.Errorf("calculate signature: %w", err)
	}

	var sigv2 refs.Signature

	sig.WriteToV2(&sigv2)

	// form request body
	reqBody := new(v2container.SetExtendedACLRequestBody)
	reqBody.SetEACL(eaclV2)
	reqBody.SetSignature(&sigv2)

	// form meta header
	var meta v2session.RequestMetaHeader
	prm.prmCommonMeta.writeToMetaHeader(&meta)

	if tok := prm.table.SessionToken(); tok != nil {
		var tokv2 v2session.Token
		tok.WriteToV2(&tokv2)

		meta.SetSessionToken(&tokv2)
	}

	// form request
	var req v2container.SetExtendedACLRequest

	req.SetBody(reqBody)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res ResContainerSetEACL
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.SetEACL(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmAnnounceSpace groups parameters of ContainerAnnounceUsedSpace operation.
type PrmAnnounceSpace struct {
	prmCommonMeta

	announcements []container.UsedSpaceAnnouncement
}

// SetValues sets values describing volume of space that is used for the container objects.
// Required parameter. Must not be empty.
//
// Must not be mutated before the end of the operation.
func (x *PrmAnnounceSpace) SetValues(announcements []container.UsedSpaceAnnouncement) {
	x.announcements = announcements
}

// ResAnnounceSpace groups resulting values of ContainerAnnounceUsedSpace operation.
type ResAnnounceSpace struct {
	statusRes
}

// ContainerAnnounceUsedSpace sends request to announce volume of the space used for the container objects.
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
// Immediately panics if parameters are set incorrectly (see PrmAnnounceSpace docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ContainerAnnounceUsedSpace(ctx context.Context, prm PrmAnnounceSpace) (*ResAnnounceSpace, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case len(prm.announcements) == 0:
		panic("missing announcements")
	}

	// convert list of SDK announcement structures into NeoFS-API v2 list
	v2announce := make([]v2container.UsedSpaceAnnouncement, len(prm.announcements))
	for i := range prm.announcements {
		v2announce[i] = *prm.announcements[i].ToV2()
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
		res ResAnnounceSpace
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceUsedSpace(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
