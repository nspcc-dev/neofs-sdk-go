package client

import (
	"context"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// PrmObjectHash groups parameters of ObjectHash operation.
type PrmObjectHash struct {
	prmCommonMeta
	sessionContainer
	bearerToken *bearer.Token
	local       bool

	tz   bool
	rs   []uint64
	salt []byte
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectHash) MarkLocal() {
	x.local = true
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectHash) WithBearerToken(t bearer.Token) {
	x.bearerToken = &t
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

	x.rs = r
}

// TillichZemorAlgo changes the hash function to Tillich-Zemor
// (https://link.springer.com/content/pdf/10.1007/3-540-48658-5_5.pdf).
//
// By default, SHA256 hash function is used.
func (x *PrmObjectHash) TillichZemorAlgo() {
	x.tz = true
}

// UseSalt sets the salt to XOR the data range before hashing.
//
// Must not be mutated before the operation completes.
func (x *PrmObjectHash) UseSalt(salt []byte) {
	x.salt = salt
}

// ObjectHash requests checksum of the range list of the object payload using
// NeoFS API protocol.
//
// To hash full payload, set both offset and length to zero. Otherwise, length
// must not be zero.
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
// If signer implements [neofscrypto.SignerV2], signing is done using it. In
// this case, [neofscrypto.Signer] methods are not called.
// [neofscrypto.OverlapSigner] may be used to pass [neofscrypto.SignerV2] when
// [neofscrypto.Signer] is unimplemented.
//
// Return errors:
//   - [ErrMissingRanges]
//   - [ErrMissingSigner]
func (c *Client) ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectHash) ([][]byte, error) {
	var err error

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectHash, time.Since(startTime), err)
		}()
	}

	if len(prm.rs) == 0 {
		err = ErrMissingRanges
		return nil, err
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}
	if prm.session != nil && prm.sessionV2 != nil {
		return nil, errSessionTokenBothVersionsSet
	}

	req := &protoobject.GetRangeHashRequest{
		Body: &protoobject.GetRangeHashRequest_Body{
			Address: oid.NewAddress(containerID, objectID).ProtoMessage(),
			Ranges:  make([]*protoobject.Range, len(prm.rs)/2),
			Salt:    prm.salt,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
		},
	}
	if prm.tz {
		req.Body.Type = refs.ChecksumType_TZ
	} else {
		req.Body.Type = refs.ChecksumType_SHA256
	}
	for i := range len(prm.rs) / 2 {
		req.Body.Ranges[i] = &protoobject.Range{
			Offset: prm.rs[2*i],
			Length: prm.rs[2*i+1],
		}
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.local {
		req.MetaHeader.Ttl = localRequestTTL
	} else {
		req.MetaHeader.Ttl = defaultRequestTTL
	}
	if prm.session != nil {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
	}
	if prm.sessionV2 != nil {
		req.MetaHeader.SessionTokenV2 = prm.sessionV2.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.GetRangeHashRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	resp, err := c.object.GetRangeHash(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, err
	}

	res := resp.GetBody().GetHashList()
	if len(res) == 0 {
		err = newErrMissingResponseField("hash list")
		return nil, err
	}

	return res, nil
}
