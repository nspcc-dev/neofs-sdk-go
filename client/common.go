package client

import (
	"crypto/ecdsa"
	"fmt"

	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/signature"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// common interface of resulting structures with API status.
type resCommon interface {
	setStatus(apistatus.Status)
}

// structure is embedded to all resulting types in order to inherit status-related methods.
type statusRes struct {
	st apistatus.Status
}

// setStatus implements resCommon interface method.
func (x *statusRes) setStatus(st apistatus.Status) {
	x.st = st
}

// Status returns server's status return.
//
// Use apistatus package functionality to handle the status.
func (x statusRes) Status() apistatus.Status {
	return x.st
}

// checks response signature and write client error if it is not correct (in this case returns true).
func isInvalidSignatureV2(res *processResponseV2Res, resp responseV2) bool {
	err := signature.VerifyServiceMessage(resp)

	isErr := err != nil
	if isErr {
		res.cliErr = fmt.Errorf("invalid response signature: %w", err)
	}

	return isErr
}

type processResponseV2Prm struct {
	callOpts *callOptions

	resp responseV2
}

type processResponseV2Res struct {
	statusRes resCommon

	cliErr error
}

// performs common actions of response processing and writes any problem as a result status or client error
// (in both cases returns true).
//
// Actions:
//  * verify signature (internal);
//  * call response callback (internal);
//  * unwrap status error (optional).
func (c *Client) processResponseV2(res *processResponseV2Res, prm processResponseV2Prm) bool {
	// verify response structure
	if isInvalidSignatureV2(res, prm.resp) {
		return true
	}

	// handle response meta info
	if err := c.handleResponseInfoV2(prm.callOpts, prm.resp); err != nil {
		res.cliErr = err
		return true
	}

	// get result status
	st := apistatus.FromStatusV2(prm.resp.GetMetaHeader().GetStatus())

	// unwrap unsuccessful status and return it
	// as error if client has been configured so
	unsuccessfulStatus := !apistatus.IsSuccessful(st)
	if unsuccessfulStatus && c.opts.parseNeoFSErrors {
		res.cliErr = apistatus.ErrFromStatus(st)
		return true
	}

	res.statusRes.setStatus(st)

	return unsuccessfulStatus
}

type prmSession struct {
	tokenSessionSet bool
	tokenSession    session.Token
}

// SetSessionToken sets token of the session within which request should be sent.
func (x *prmSession) SetSessionToken(tok session.Token) {
	x.tokenSession = tok
	x.tokenSessionSet = true
}

func (x prmSession) writeToMetaHeader(meta *v2session.RequestMetaHeader) {
	if x.tokenSessionSet {
		meta.SetSessionToken(x.tokenSession.ToV2())
	}
}

// panic messages.
const (
	panicMsgMissingContext   = "missing context"
	panicMsgMissingContainer = "missing container"
)

// groups all the details required to send a single request and process a response to it.
type contextCall struct {
	// ==================================================
	// state vars that do not require explicit initialization

	// final error to be returned from client method
	err error

	// received response
	resp responseV2

	// ==================================================
	// shared parameters which are set uniformly on all calls

	// request signing key
	key ecdsa.PrivateKey

	// callback prior to processing the response by the client
	callbackResp func(ResponseMetaInfo) error

	// if set, protocol errors will be expanded into a final error
	resolveAPIFailures bool

	// NeoFS network magic
	netMagic uint64

	// ==================================================
	// custom call parameters

	// structure of the call result
	statusRes resCommon

	// request to be signed with a key and sent
	req interface {
		GetMetaHeader() *v2session.RequestMetaHeader
		SetMetaHeader(*v2session.RequestMetaHeader)
	}

	// function to send a request (unary) and receive a response
	call func() (responseV2, error)

	// function of writing response fields to the resulting structure (optional)
	result func(v2 responseV2)
}

func (x contextCall) prepareRequest() {
	meta := x.req.GetMetaHeader()
	if meta == nil {
		meta = new(v2session.RequestMetaHeader)
		x.req.SetMetaHeader(meta)
	}

	if meta.GetTTL() == 0 {
		meta.SetTTL(2)
	}

	if meta.GetVersion() == nil {
		meta.SetVersion(version.Current().ToV2())
	}

	meta.SetNetworkMagic(x.netMagic)
}

// performs common actions of response processing and writes any problem as a result status or client error
// (in both cases returns false).
//
// Actions:
//  * verify signature (internal);
//  * call response callback (internal);
//  * unwrap status error (optional).
func (x *contextCall) processResponse() bool {
	// call response callback if set
	if x.callbackResp != nil {
		x.err = x.callbackResp(ResponseMetaInfo{
			key: x.resp.GetVerificationHeader().GetBodySignature().GetKey(),
		})
		if x.err != nil {
			x.err = fmt.Errorf("response callback error: %w", x.err)
			return false
		}
	}

	// note that we call response callback before signature check since it is expected more lightweight
	// while verification needs marshaling

	// verify response signature
	x.err = signature.VerifyServiceMessage(x.resp)
	if x.err != nil {
		x.err = fmt.Errorf("invalid response signature: %w", x.err)
		return false
	}

	// get result status
	st := apistatus.FromStatusV2(x.resp.GetMetaHeader().GetStatus())

	// unwrap unsuccessful status and return it
	// as error if client has been configured so
	successfulStatus := apistatus.IsSuccessful(st)
	if !successfulStatus && x.resolveAPIFailures {
		x.err = apistatus.ErrFromStatus(st)
		return false
	}

	x.statusRes.setStatus(st)

	return successfulStatus
}

// goes through all stages of sending a request and processing a response. Returns true if successful.
func (x *contextCall) processCall() bool {
	// prepare the request
	x.prepareRequest()

	// sign the request
	x.err = signature.SignServiceMessage(&x.key, x.req)
	if x.err != nil {
		x.err = fmt.Errorf("sign request: %w", x.err)
		return false
	}

	// perform RPC
	x.resp, x.err = x.call()
	if x.err != nil {
		x.err = fmt.Errorf("transport error: %w", x.err)
		return false
	}

	// process the response
	ok := x.processResponse()
	if !ok {
		return false
	}

	// write response to resulting structure
	if x.result != nil {
		x.result(x.resp)
	}

	return true
}

// initializes static cross-call parameters inherited from client.
func (c *Client) initCallContext(ctx *contextCall) {
	ctx.key = *c.opts.key
	ctx.resolveAPIFailures = c.opts.parseNeoFSErrors
	ctx.callbackResp = c.opts.cbRespInfo
	ctx.netMagic = c.opts.netMagic
}
