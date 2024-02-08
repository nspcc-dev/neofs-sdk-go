package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	objectgrpc "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/message"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"google.golang.org/protobuf/encoding/protowire"
)

// TODO: docs
type ReplicatedObject struct {
	s int
	r io.Reader
}

func (x ReplicatedObject) Reset() error {
	if s, ok := x.r.(io.Seeker); ok {
		_, err := s.Seek(0, io.SeekStart)
		return err
	}
	return nil
}

func ReplicateFromReader(s int, r io.Reader) ReplicatedObject {
	return ReplicatedObject{s, r}
}

func ReplicateFromReadSeeker(rs io.ReadSeeker) ReplicatedObject {
	return ReplicatedObject{-1, rs}
}

// TODO: update docs

// ReplicateObject copies binary-encoded NeoFS object from the given
// [io.ReadSeeker] to remote server for local storage. The signer must
// authenticate a storage node that matches the object's storage policy. Since
// this property can change over NeoFS system time, compliance with the policy
// is checked back to foreseeable moment in the past. The server must be a
// storage node compliant with the current object's storage policy.
//
// ReplicateObject is intended for maintaining data storage by NeoFS system
// nodes only, not for regular use.
//
// Object must be encoded in compliance with Protocol Buffers v3 format in
// ascending order of fields.
//
// Source [io.ReadSeeker] must point to the start. Note that ReplicateObject
// does not reset src to start after the call. If it is needed, do not forget to
// Seek.
//
// See also [DemuxReplicatedObject].
//
// Return errors:
//   - [apistatus.ErrServerInternal]: internal server error described in the text message;
//   - [apistatus.ErrObjectAccessDenied]: the signer does not authenticate any
//     NeoFS storage node matching storage policy of the container referenced by the
//     replicated object;
//   - [apistatus.ErrContainerNotFound]: the container to which the replicated
//     object is associated was not found.
func (c *Client) ReplicateObject(ctx context.Context, id oid.ID, ro ReplicatedObject, signer neofscrypto.Signer) error {
	if ro.s < 0 {
		s, ok := ro.r.(io.Seeker)
		if !ok {
			return errors.New("negative size")
		}
		var sz int64
		switch v := s.(type) {
		default:
			var err error
			sz, err = s.Seek(0, io.SeekEnd)
			if err != nil {
				return fmt.Errorf("seek to end: %w", err)
			} else if sz < 0 {
				return fmt.Errorf("seek to end returned negative value %d", sz)
			}
			_, err = s.Seek(-sz, io.SeekCurrent)
			if err != nil {
				return fmt.Errorf("seek back to initial pos: %w", err)
			}
		case *os.File:
			fileInfo, err := v.Stat()
			if err != nil {
				return fmt.Errorf("get file info: %w", err)
			}
			sz = fileInfo.Size()
		case *bytes.Reader:
			sz = v.Size()
			if sz < 0 {
				return fmt.Errorf("negative byte buffer size return %d", sz)
			}
		}
		if sz > math.MaxInt {
			return fmt.Errorf("object size is too big for this OS %d > %d", sz, math.MaxInt)
		}
		ro.s = int(sz)
	}

	const svcName = "neo.fs.v2.object.ObjectService"
	const opName = "Replicate"
	stream, err := c.c.Init(common.CallMethodInfoUnary(svcName, opName),
		client.WithContext(ctx), client.AllowBinarySendingOnly())
	if err != nil {
		return fmt.Errorf("init service=%s/op=%s RPC: %w", svcName, opName, err)
	}

	r, sz, err := prepareReplicateMessage(id, ro, signer)
	if err != nil {
		return err
	}

	err = stream.WriteMessage(client.BinaryMessage{
		R:    r,
		Size: sz,
	})
	if err != nil && !errors.Is(err, io.EOF) { // io.EOF means the server closed the stream on its side
		return fmt.Errorf("send request: %w", err)
	}

	var resp replicateResponse
	err = stream.ReadMessage(&resp)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}

		return fmt.Errorf("recv response: %w", err)
	}

	_ = stream.Close()

	return resp.err
}

func prepareReplicateMessage(id oid.ID, ro ReplicatedObject, signer neofscrypto.Signer) (io.Reader, int, error) {
	if ro.s < 0 {
		panic("negative size")
	}

	var buf [sha256.Size]byte
	id.Encode(buf[:])
	objSig, err := signer.Sign(buf[:])
	if err != nil {
		return nil, 0, fmt.Errorf("sign object ID: %w", err)
	}

	bPubKey := neofscrypto.PublicKeyBytes(signer.Public())
	sigScheme := uint64(signer.Scheme())

	const fieldNumObject = 1
	const fieldNumSignature = 2

	sigSize := protowire.SizeTag(fieldNumSigPubKey) + protowire.SizeBytes(len(bPubKey)) +
		protowire.SizeTag(fieldNumSigVal) + +protowire.SizeBytes(len(objSig)) +
		protowire.SizeTag(fieldNumSigScheme) + protowire.SizeVarint(sigScheme)

	sigFieldSize := protowire.SizeTag(fieldNumSignature) + protowire.SizeBytes(sigSize)

	// TODO(#544): we can reuse such buffers - remember fieldNumSigVal field
	//  offset/len and pool them all
	// TODO: the data below could be read smart or even piped. For the sake of
	//  code simplicity, additional buffer is used for now
	sigField := make([]byte, 0, sigFieldSize)
	sigField = protowire.AppendTag(sigField, fieldNumSignature, protowire.BytesType)
	sigField = protowire.AppendVarint(sigField, uint64(sigSize))
	sigField = protowire.AppendTag(sigField, fieldNumSigPubKey, protowire.BytesType)
	sigField = protowire.AppendBytes(sigField, bPubKey)
	sigField = protowire.AppendTag(sigField, fieldNumSigVal, protowire.BytesType)
	sigField = protowire.AppendBytes(sigField, objSig)
	sigField = protowire.AppendTag(sigField, fieldNumSigScheme, protowire.VarintType)
	sigField = protowire.AppendVarint(sigField, sigScheme)

	objFieldPrefixSize := protowire.SizeTag(fieldNumObject) + protowire.SizeVarint(uint64(ro.s))
	objFieldPrefix := make([]byte, 0, objFieldPrefixSize)
	objFieldPrefix = protowire.AppendTag(objFieldPrefix, fieldNumObject, protowire.BytesType)
	objFieldPrefix = protowire.AppendVarint(objFieldPrefix, uint64(ro.s))

	ro.r = io.LimitReader(ro.r, int64(ro.s))

	if sigFieldSize < objFieldPrefixSize+ro.s { // overflow is not expected in practice
		r := io.MultiReader(bytes.NewReader(sigField), io.MultiReader(bytes.NewReader(objFieldPrefix), ro.r))
		return r, sigFieldSize + objFieldPrefixSize + ro.s, nil
	}

	r := io.MultiReader(io.MultiReader(bytes.NewReader(objFieldPrefix), ro.r), bytes.NewReader(sigField))
	return r, sigFieldSize + objFieldPrefixSize + ro.s, nil
}

type replicateResponse struct {
	err error
}

func (x replicateResponse) ToGRPCMessage() grpc.Message {
	return new(objectgrpc.ReplicateResponse)
}

func (x *replicateResponse) FromGRPCMessage(gm grpc.Message) error {
	m, ok := gm.(*objectgrpc.ReplicateResponse)
	if !ok {
		return message.NewUnexpectedMessageType(gm, m)
	}

	var st *status.Status
	if mst := m.GetStatus(); mst != nil {
		st = new(status.Status)
		err := st.FromGRPCMessage(mst)
		if err != nil {
			return fmt.Errorf("decode response status: %w", err)
		}
	}

	x.err = apistatus.ErrorFromV2(st)

	return nil
}
