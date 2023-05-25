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
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// PrmObjectHash groups parameters of ObjectHash operation.
type PrmObjectHash struct {
	meta v2session.RequestMetaHeader

	body v2object.GetRangeHashRequestBody

	csAlgo v2refs.ChecksumType

	signer neofscrypto.Signer
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectHash) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
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
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingRanges]
//   - [ErrMissingSigner]
func (c *Client) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectHash) (*ResObjectHash, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
	)

	if len(prm.body.GetRanges()) == 0 {
		return nil, ErrMissingRanges
	}

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
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

	err = signServiceMessage(signer, &req)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}

	resp, err := rpcapi.HashObjectRange(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	var res ResObjectHash
	if err = c.processResponse(resp); err != nil {
		return nil, err
	}

	res.checksums = resp.GetBody().GetHashList()
	if len(res.checksums) == 0 {
		return nil, newErrMissingResponseField("hash list")
	}

	return &res, nil
}
