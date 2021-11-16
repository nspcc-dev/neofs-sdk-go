package client

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/signature"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
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
//  * call response callback (internal).
func (c *clientImpl) processResponseV2(res *processResponseV2Res, prm processResponseV2Prm) bool {
	// verify response structure
	if isInvalidSignatureV2(res, prm.resp) {
		return true
	}

	// handle response meta info
	if err := c.handleResponseInfoV2(prm.callOpts, prm.resp); err != nil {
		res.cliErr = err
		return true
	}

	// set result status
	st := apistatus.FromStatusV2(prm.resp.GetMetaHeader().GetStatus())

	res.statusRes.setStatus(st)

	return !apistatus.IsSuccessful(st)
}
