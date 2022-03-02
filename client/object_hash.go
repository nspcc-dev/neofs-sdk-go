package client

import (
	"context"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

// PrmObjectHash groups parameters of ObjectHash operation.
type PrmObjectHash struct {
	meta v2session.RequestMetaHeader

	body v2object.GetRangeHashRequestBody

	tillichZemor bool

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
func (x *PrmObjectHash) WithinSession(t session.Token) {
	x.meta.SetSessionToken(t.ToV2())
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectHash) WithBearerToken(t token.BearerToken) {
	x.meta.SetBearerToken(t.ToV2())
}

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *PrmObjectHash) FromContainer(id cid.ID) {
	x.addr.SetContainerID(id.ToV2())
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *PrmObjectHash) ByID(id oid.ID) {
	x.addr.SetObjectID(id.ToV2())
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

	rs := make([]*v2object.Range, ln/2)

	for i := 0; i < ln/2; i++ {
		rs[i] = new(v2object.Range)
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
	x.tillichZemor = true
}

// UseSalt sets the salt to XOR the data range before hashing.
//
// Must not be mutated before the operation completes.
func (x *PrmObjectHash) UseSalt(salt []byte) {
	x.body.SetSalt(salt)
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
// If WithNeoFSErrorParsing option has been provided, unsuccessful
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
//   - *apistatus.SessionTokenExpired.
func (c *Client) ObjectHash(ctx context.Context, prm PrmObjectHash) (*ResObjectHash, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic("missing object")
	case len(prm.body.GetRanges()) == 0:
		panic("missing ranges")
	}

	// form request body
	prm.body.SetAddress(&prm.addr)
	// ranges and salt are already by prm setters

	if prm.tillichZemor {
		prm.body.SetType(v2refs.TillichZemor)
	} else {
		prm.body.SetType(v2refs.SHA256)
	}

	// form request
	var req v2object.GetRangeHashRequest
	req.SetBody(&prm.body)
	req.SetMetaHeader(&prm.meta)

	// init call context
	var (
		cc  contextCall
		res ResObjectHash
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.HashObjectRange(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		res.checksums = r.(*v2object.GetRangeHashResponse).GetBody().GetHashList()
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
