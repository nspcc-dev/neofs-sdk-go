package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// PrmObjectHash groups parameters of ObjectHash operation.
type PrmObjectHash struct {
	meta v2session.RequestMetaHeader

	body v2object.GetRangeHashRequestBody

	csAlgo v2refs.ChecksumType

	addr v2refs.Address
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectHash) MarkLocal() {
	x.meta.SetTTL(1)
}

// WithinSession specifies session within which object should be read.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *PrmObjectHash) WithinSession(t session.Object) {
	var tv2 v2session.Token
	t.WriteToV2(&tv2)

	x.meta.SetSessionToken(&tv2)
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectHash) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
}

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *PrmObjectHash) FromContainer(id cid.ID) {
	var cidV2 v2refs.ContainerID
	id.WriteToV2(&cidV2)

	x.addr.SetContainerID(&cidV2)
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *PrmObjectHash) ByID(id oid.ID) {
	var idV2 v2refs.ObjectID
	id.WriteToV2(&idV2)

	x.addr.SetObjectID(&idV2)
}

// SetRangeList sets list of ranges in (offset, length) pair format.
// Required parameter.
//
// If passed as slice, then it must not be mutated before the operation completes.
func (x *PrmObjectHash) SetRangeList(r ...uint64) {
	ln := len(r)
	if ln%2 != 0 {
		panic("odd number of range parameters")
	}

	rs := make([]v2object.Range, ln/2)

	for i := 0; i < ln/2; i++ {
		rs[i].SetOffset(r[2*i])
		rs[i].SetLength(r[2*i+1])
	}

	x.body.SetRanges(rs)
}

// TillichZemorAlgo changes the hash function to Tillich-Zemor
// (https://link.springer.com/content/pdf/10.1007/3-540-48658-5_5.pdf).
//
// By default, SHA256 hash function is used.
func (x *PrmObjectHash) TillichZemorAlgo() {
	x.csAlgo = v2refs.TillichZemor
}

// UseSalt sets the salt to XOR the data range before hashing.
//
// Must not be mutated before the operation completes.
func (x *PrmObjectHash) UseSalt(salt []byte) {
	x.body.SetSalt(salt)
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *PrmObjectHash) WithXHeaders(hs ...string) {
	writeXHeadersToMeta(hs, &x.meta)
}

// ResObjectHash groups resulting values of ObjectHash operation.
type ResObjectHash struct {
	statusRes

	checksums [][]byte
}

// Checksums returns a list of calculated checksums in range order.
func (x ResObjectHash) Checksums() [][]byte {
	return x.checksums
}

// ObjectHash requests checksum of the range list of the object payload using
// NeoFS API protocol.
//
// Returns a list of checksums in raw form: the format of hashes and their number
// is left for the caller to check. Client preserves the order of the server's response.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// If PrmInit.ResolveNeoFSFailures has been called, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectHash docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound;
//   - *apistatus.ObjectNotFound;
//   - *apistatus.ObjectAccessDenied;
//   - *apistatus.ObjectOutOfRange;
//   - *apistatus.SessionTokenExpired.
func (c *Client) ObjectHash(ctx context.Context, prm PrmObjectHash) (*ResObjectHash, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic(panicMsgMissingObject)
	case len(prm.body.GetRanges()) == 0:
		panic("missing ranges")
	}

	prm.body.SetAddress(&prm.addr)
	if prm.csAlgo == v2refs.UnknownChecksum {
		prm.body.SetType(v2refs.SHA256)
	} else {
		prm.body.SetType(prm.csAlgo)
	}

	var req v2object.GetRangeHashRequest
	c.prepareRequest(&req, &prm.meta)
	req.SetBody(&prm.body)

	err := signature.SignServiceMessage(&c.prm.key, req)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}

	resp, err := rpcapi.HashObjectRange(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	var res ResObjectHash
	res.st, err = c.processResponse(resp)
	if err != nil {
		return nil, err
	}

	if !apistatus.IsSuccessful(res.st) {
		return &res, nil
	}

	res.checksums = resp.GetBody().GetHashList()
	if len(res.checksums) == 0 {
		return nil, newErrMissingResponseField("hash list")
	}

	return &res, nil
}
