package client

import (
	"context"
	"fmt"

	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

type delContainerSignWrapper struct {
	body *v2container.DeleteRequestBody
}

// EACLWithSignature represents eACL table/signature pair.
type EACLWithSignature struct {
	table *eacl.Table
}

func (c delContainerSignWrapper) ReadSignedData(bytes []byte) ([]byte, error) {
	return c.body.GetContainerID().GetValue(), nil
}

func (c delContainerSignWrapper) SignedDataSize() int {
	return len(c.body.GetContainerID().GetValue())
}

// EACL returns eACL table.
func (e EACLWithSignature) EACL() *eacl.Table {
	return e.table
}

// Signature returns table signature.
//
// Deprecated: use EACL().Signature() instead.
func (e EACLWithSignature) Signature() *signature.Signature {
	return e.table.Signature()
}

type ContainerPutRes struct {
	statusRes

	id *cid.ID
}

func (x ContainerPutRes) ID() *cid.ID {
	return x.id
}

func (x *ContainerPutRes) setID(id *cid.ID) {
	x.id = id
}

// PutContainer puts container through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) PutContainer(ctx context.Context, cnr *container.Container, opts ...CallOption) (*ContainerPutRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	// set transport version
	cnr.SetVersion(version.Current())

	// if container owner is not set, then use client key as owner
	if cnr.OwnerID() == nil {
		ownerID := owner.NewIDFromPublicKey(&callOptions.key.PublicKey)

		cnr.SetOwnerID(ownerID)
	}

	reqBody := new(v2container.PutRequestBody)
	reqBody.SetContainer(cnr.ToV2())

	// sign container
	signWrapper := v2signature.StableMarshalerWrapper{SM: reqBody.GetContainer()}

	err := sigutil.SignDataWithHandler(callOptions.key, signWrapper, func(key []byte, sig []byte) {
		containerSignature := new(refs.Signature)
		containerSignature.SetKey(key)
		containerSignature.SetSign(sig)
		reqBody.SetSignature(containerSignature)
	}, sigutil.SignWithRFC6979())
	if err != nil {
		return nil, err
	}

	req := new(v2container.PutRequest)
	req.SetBody(reqBody)

	meta := v2MetaHeaderFromOpts(callOptions)
	meta.SetSessionToken(cnr.SessionToken().ToV2())

	req.SetMetaHeader(meta)

	err = v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.PutContainer(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var (
		res     = new(ContainerPutRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	// sets result status
	st := apistatus.FromStatusV2(resp.GetMetaHeader().GetStatus())

	res.setStatus(st)

	if apistatus.IsSuccessful(st) {
		res.setID(cid.NewFromV2(resp.GetBody().GetContainerID()))
	}

	return res, nil
}

type ContainerGetRes struct {
	statusRes

	cnr *container.Container
}

func (x ContainerGetRes) Container() *container.Container {
	return x.cnr
}

func (x *ContainerGetRes) setContainer(cnr *container.Container) {
	x.cnr = cnr
}

// GetContainer receives container structure through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) GetContainer(ctx context.Context, id *cid.ID, opts ...CallOption) (*ContainerGetRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2container.GetRequestBody)
	reqBody.SetContainerID(id.ToV2())

	req := new(v2container.GetRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.GetContainer(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(ContainerGetRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	body := resp.GetBody()

	cnr := container.NewContainerFromV2(body.GetContainer())

	cnr.SetSessionToken(
		session.NewTokenFromV2(body.GetSessionToken()),
	)

	cnr.SetSignature(
		signature.NewFromV2(body.GetSignature()),
	)

	res.setContainer(cnr)

	return res, nil
}

type ContainerListRes struct {
	statusRes

	ids []*cid.ID
}

func (x ContainerListRes) IDList() []*cid.ID {
	return x.ids
}

func (x *ContainerListRes) setIDList(ids []*cid.ID) {
	x.ids = ids
}

// ListContainers receives all owner's containers through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) ListContainers(ctx context.Context, ownerID *owner.ID, opts ...CallOption) (*ContainerListRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	if ownerID == nil {
		ownerID = owner.NewIDFromPublicKey(&callOptions.key.PublicKey)
	}

	reqBody := new(v2container.ListRequestBody)
	reqBody.SetOwnerID(ownerID.ToV2())

	req := new(v2container.ListRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.ListContainers(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(ContainerListRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	ids := make([]*cid.ID, 0, len(resp.GetBody().GetContainerIDs()))

	for _, cidV2 := range resp.GetBody().GetContainerIDs() {
		ids = append(ids, cid.NewFromV2(cidV2))
	}

	res.setIDList(ids)

	return res, nil
}

type ContainerDeleteRes struct {
	statusRes
}

// DeleteContainer deletes specified container through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) DeleteContainer(ctx context.Context, id *cid.ID, opts ...CallOption) (*ContainerDeleteRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2container.DeleteRequestBody)
	reqBody.SetContainerID(id.ToV2())

	// sign container
	err := sigutil.SignDataWithHandler(callOptions.key,
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

	req := new(v2container.DeleteRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err = v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.DeleteContainer(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(ContainerDeleteRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	return res, nil
}

type EACLRes struct {
	statusRes

	table *eacl.Table
}

func (x EACLRes) Table() *eacl.Table {
	return x.table
}

func (x *EACLRes) SetTable(table *eacl.Table) {
	x.table = table
}

// EACL receives eACL of the specified container through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) EACL(ctx context.Context, id *cid.ID, opts ...CallOption) (*EACLRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2container.GetExtendedACLRequestBody)
	reqBody.SetContainerID(id.ToV2())

	req := new(v2container.GetExtendedACLRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.GetEACL(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(EACLRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	body := resp.GetBody()

	table := eacl.NewTableFromV2(body.GetEACL())

	table.SetSessionToken(
		session.NewTokenFromV2(body.GetSessionToken()),
	)

	table.SetSignature(
		signature.NewFromV2(body.GetSignature()),
	)

	res.SetTable(table)

	return res, nil
}

type SetEACLRes struct {
	statusRes
}

// SetEACL sets eACL through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) SetEACL(ctx context.Context, eacl *eacl.Table, opts ...CallOption) (*SetEACLRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2container.SetExtendedACLRequestBody)
	reqBody.SetEACL(eacl.ToV2())

	signWrapper := v2signature.StableMarshalerWrapper{SM: reqBody.GetEACL()}

	err := sigutil.SignDataWithHandler(callOptions.key, signWrapper, func(key []byte, sig []byte) {
		eaclSignature := new(refs.Signature)
		eaclSignature.SetKey(key)
		eaclSignature.SetSign(sig)
		reqBody.SetSignature(eaclSignature)
	}, sigutil.SignWithRFC6979())
	if err != nil {
		return nil, err
	}

	req := new(v2container.SetExtendedACLRequest)
	req.SetBody(reqBody)

	meta := v2MetaHeaderFromOpts(callOptions)
	meta.SetSessionToken(eacl.SessionToken().ToV2())

	req.SetMetaHeader(meta)

	err = v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.SetEACL(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(SetEACLRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	return res, nil
}

type AnnounceSpaceRes struct {
	statusRes
}

// AnnounceContainerUsedSpace used by storage nodes to estimate their container
// sizes during lifetime. Use it only in storage node applications.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) AnnounceContainerUsedSpace(
	ctx context.Context,
	announce []container.UsedSpaceAnnouncement,
	opts ...CallOption,
) (*AnnounceSpaceRes, error) {
	callOptions := c.defaultCallOptions() // apply all available options

	for i := range opts {
		opts[i](callOptions)
	}

	// convert list of SDK announcement structures into NeoFS-API v2 list
	v2announce := make([]*v2container.UsedSpaceAnnouncement, 0, len(announce))
	for i := range announce {
		v2announce = append(v2announce, announce[i].ToV2())
	}

	// prepare body of the NeoFS-API v2 request and request itself
	reqBody := new(v2container.AnnounceUsedSpaceRequestBody)
	reqBody.SetAnnouncements(v2announce)

	req := new(v2container.AnnounceUsedSpaceRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	// sign the request
	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.AnnounceUsedSpace(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(AnnounceSpaceRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	return res, nil
}
