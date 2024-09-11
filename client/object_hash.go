package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var (
	// special variable for test purposes only, to overwrite real RPC calls.
	rpcAPIHashObjectRange = rpcapi.HashObjectRange
)

// PrmObjectHash groups parameters of ObjectHash operation.
type PrmObjectHash struct {
	sessionContainer

	body v2object.GetRangeHashRequestBody

	csAlgo v2refs.ChecksumType
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectHash) MarkLocal() {
	x.meta.SetTTL(1)
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

	for i := range ln / 2 {
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

// ObjectHash requests checksum of the range list of the object payload using
// NeoFS API protocol.
//
// Returns a list of checksums in raw form: the format of hashes and their number
// is left for the caller to check. Client preserves the order of the server's response.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
//
// Return errors:
//   - [ErrMissingRanges]
//   - [ErrMissingSigner]
func (c *Client) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectHash) ([][]byte, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		err   error
	)

	defer func() {
		c.sendStatistic(stat.MethodObjectHash, err)()
	}()

	if len(prm.body.GetRanges()) == 0 {
		err = ErrMissingRanges
		return nil, err
	}

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	if signer == nil {
		return nil, ErrMissingSigner
	}

	prm.body.SetAddress(&addr)
	if prm.csAlgo == v2refs.UnknownChecksum {
		prm.body.SetType(v2refs.SHA256)
	} else {
		prm.body.SetType(prm.csAlgo)
	}

	var req v2object.GetRangeHashRequest
	c.prepareRequest(&req, &prm.meta)
	req.SetBody(&prm.body)

	buf := c.buffers.Get().(*[]byte)
	err = signServiceMessage(signer, &req, *buf)
	c.buffers.Put(buf)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return nil, err
	}

	resp, err := rpcAPIHashObjectRange(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		err = fmt.Errorf("write request: %w", err)
		return nil, err
	}

	var res [][]byte
	if err = c.processResponse(resp); err != nil {
		return nil, err
	}

	res = resp.GetBody().GetHashList()
	if len(res) == 0 {
		err = newErrMissingResponseField("hash list")
		return nil, err
	}

	return res, nil
}
