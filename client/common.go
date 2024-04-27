package client

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Various field numbers in from NeoFS API definitions.
const (
	fieldNumSigPubKey = 1
	fieldNumSigVal    = 2
	fieldNumSigScheme = 3
)

// groups meta parameters shared between all Client operations.
type prmCommonMeta struct {
	// NeoFS request X-Headers
	xHeaders []string
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *prmCommonMeta) WithXHeaders(hs ...string) {
	if len(hs)%2 != 0 {
		panic("slice of X-Headers with odd length")
	}

	x.xHeaders = hs
}

func writeXHeadersToMeta(xHeaders []string, h *v2session.RequestMetaHeader) {
	if len(xHeaders) == 0 {
		return
	}

	if len(xHeaders)%2 != 0 {
		panic("slice of X-Headers with odd length")
	}

	hs := make([]v2session.XHeader, len(xHeaders)/2)
	j := 0

	for i := 0; i < len(xHeaders); i += 2 {
		hs[j].SetKey(xHeaders[i])
		hs[j].SetValue(xHeaders[i+1])
		j++
	}

	h.SetXHeaders(hs)
}

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

	// request signer
	signer neofscrypto.Signer

	// callback prior to processing the response by the client
	callbackResp func(ResponseMetaInfo) error

	// NeoFS network magic
	netMagic uint64

	// Meta parameters
	meta prmCommonMeta

	// ==================================================
	// custom call parameters

	// request to be signed with a signer and sent
	req request

	// function to send a request (unary) and receive a response
	call func() (responseV2, error)

	// function to send the request (req field)
	wReq func() error

	// function to recv the response (resp field)
	rResp func() error

	// function to close the message stream
	closer func() error

	// function of writing response fields to the resulting structure (optional)
	result func(v2 responseV2)

	buf              []byte
	bufCleanCallback func()
}

type request interface {
	GetMetaHeader() *v2session.RequestMetaHeader
	SetMetaHeader(*v2session.RequestMetaHeader)
	SetVerificationHeader(*v2session.RequestVerificationHeader)
}

// sets needed fields of the request meta header.
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
		var verV2 refs.Version
		version.Current().WriteToV2(&verV2)
		meta.SetVersion(&verV2)
	}

	meta.SetNetworkMagic(x.netMagic)

	writeXHeadersToMeta(x.meta.xHeaders, meta)
}

func (c *Client) prepareRequest(req request, meta *v2session.RequestMetaHeader) {
	ttl := meta.GetTTL()
	if ttl == 0 {
		ttl = 2
	}

	verV2 := meta.GetVersion()
	if verV2 == nil {
		verV2 = new(refs.Version)
		version.Current().WriteToV2(verV2)
	}

	meta.SetTTL(ttl)
	meta.SetVersion(verV2)
	meta.SetNetworkMagic(c.prm.netMagic)

	req.SetMetaHeader(meta)
}

// prepares, signs and writes the request. Result means success.
// If failed, contextCall.err contains the reason.
func (x *contextCall) writeRequest() bool {
	x.prepareRequest()

	x.req.SetVerificationHeader(nil)

	// sign the request
	x.err = signServiceMessage(x.signer, x.req, x.buf)
	if x.err != nil {
		x.err = fmt.Errorf("sign request: %w", x.err)
		return false
	}

	x.err = x.wReq()
	if x.err != nil {
		x.err = fmt.Errorf("write request: %w", x.err)
		return false
	}

	return true
}

// performs common actions of response processing and writes any problem as a result status or client error
// (in both cases returns false).
//
// Actions:
//   - verify signature (internal);
//   - call response callback (internal);
//   - unwrap status error (optional).
func (x *contextCall) processResponse() bool {
	// call response callback if set
	if x.callbackResp != nil {
		x.err = x.callbackResp(ResponseMetaInfo{
			key:   x.resp.GetVerificationHeader().GetBodySignature().GetKey(),
			epoch: x.resp.GetMetaHeader().GetEpoch(),
		})
		if x.err != nil {
			x.err = fmt.Errorf("response callback error: %w", x.err)
			return false
		}
	}

	// note that we call response callback before signature check since it is expected more lightweight
	// while verification needs marshaling

	// verify response signature
	x.err = verifyServiceMessage(x.resp)
	if x.err != nil {
		x.err = fmt.Errorf("invalid response signature: %w", x.err)
		return false
	}

	// get result status
	x.err = apistatus.ErrorFromV2(x.resp.GetMetaHeader().GetStatus())
	return x.err == nil
}

// processResponse verifies response signature.
func (c *Client) processResponse(resp responseV2) error {
	if err := verifyServiceMessage(resp); err != nil {
		return fmt.Errorf("invalid response signature: %w", err)
	}

	return apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
}

// reads response (if rResp is set) and processes it. Result means success.
// If failed, contextCall.err contains the reason.
func (x *contextCall) readResponse() bool {
	if x.rResp != nil {
		x.err = x.rResp()
		if x.err != nil {
			x.err = fmt.Errorf("read response: %w", x.err)
			return false
		}
	}

	return x.processResponse()
}

// closes the message stream (if closer is set) and writes the results (if result is set).
// Return means success. If failed, contextCall.err contains the reason.
func (x *contextCall) close() bool {
	if x.closer != nil {
		x.err = x.closer()
		if x.err != nil {
			x.err = fmt.Errorf("close RPC: %w", x.err)
			return false
		}
	}

	// write response to resulting structure
	if x.result != nil {
		x.result(x.resp)
	}

	return x.err == nil
}

// goes through all stages of sending a request and processing a response. Returns true if successful.
// If failed, contextCall.err contains the reason.
func (x *contextCall) processCall() bool {
	// set request writer
	x.wReq = func() error {
		var err error
		x.resp, err = x.call()
		return err
	}

	// write request
	ok := x.writeRequest()
	if x.bufCleanCallback != nil {
		x.bufCleanCallback()
	}

	if !ok {
		return false
	}

	// read response
	ok = x.readResponse()
	if !ok {
		return x.err == nil
	}

	// close and write response to resulting structure
	ok = x.close()
	if !ok {
		return false
	}

	return x.err == nil
}

// initializes static cross-call parameters inherited from client.
func (c *Client) initCallContext(ctx *contextCall) {
	ctx.signer = c.prm.signer
	ctx.callbackResp = c.prm.cbRespInfo
	ctx.netMagic = c.prm.netMagic

	buf := c.buffers.Get().(*[]byte)
	ctx.buf = *buf
	ctx.bufCleanCallback = func() {
		c.buffers.Put(buf)
	}
}

// ExecRaw executes f with underlying github.com/nspcc-dev/neofs-api-go/v2/rpc/client.Client
// instance. Communicate over the Protocol Buffers protocol in a more flexible way:
// most often used to transmit data over a fixed version of the NeoFS protocol, as well
// as to support custom services.
//
// The f must not manipulate the client connection passed into it.
//
// Like all other operations, must be called after connecting to the server and
// before closing the connection.
//
// See also Dial and Close.
// See also github.com/nspcc-dev/neofs-api-go/v2/rpc/client package docs.
func (c *Client) ExecRaw(f func(client *client.Client) error) error {
	return f(&c.c)
}
