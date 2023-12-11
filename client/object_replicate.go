package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	objectgrpc "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/message"
	"github.com/nspcc-dev/neofs-api-go/v2/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"google.golang.org/protobuf/encoding/protowire"
)

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
// Return errors:
//   - [apistatus.ErrServerInternal]: internal server error described in the text message;
//   - [apistatus.ErrObjectAccessDenied]: the signer does not authenticate any
//     NeoFS storage node matching storage policy of the container referenced by the
//     replicated object;
//   - [apistatus.ErrContainerNotFound]: the container to which the replicated
//     object is associated was not found.
func (c *Client) ReplicateObject(ctx context.Context, src io.ReadSeeker, signer neofscrypto.Signer) error {
	const svcName = "neo.fs.v2.object.ObjectService"
	const opName = "Replicate"
	stream, err := c.c.Init(common.CallMethodInfoUnary(svcName, opName),
		client.WithContext(ctx), client.AllowBinarySendingOnly())
	if err != nil {
		return fmt.Errorf("init service=%s/op=%s RPC: %w", svcName, opName, err)
	}

	msg, err := prepareReplicateMessage(src, signer)
	if err != nil {
		return err
	}

	err = stream.WriteMessage(client.BinaryMessage(msg))
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

func prepareReplicateMessage(src io.ReadSeeker, signer neofscrypto.Signer) ([]byte, error) {
	var objSize uint64
	switch v := src.(type) {
	default:
		n, err := src.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, fmt.Errorf("seek to end: %w", err)
		} else if n < 0 {
			return nil, fmt.Errorf("seek to end returned negative value %d", objSize)
		}

		_, err = src.Seek(-n, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("seek back to initial pos: %w", err)
		}

		objSize = uint64(n)
	case *os.File:
		fileInfo, err := v.Stat()
		if err != nil {
			return nil, fmt.Errorf("get file info: %w", err)
		}

		objSize = uint64(fileInfo.Size())
	case *bytes.Reader:
		n := v.Size()
		if n < 0 {
			return nil, fmt.Errorf("negative byte buffer size return %d", objSize)
		}

		objSize = uint64(n)
	}

	// TODO: limit the objSize?

	// calculate template signature to know its size
	sigTmpl, err := signer.Sign(nil)
	if err != nil {
		return nil, fmt.Errorf("calculate signature of empty ata: %w", err)
	}

	bPubKey := neofscrypto.PublicKeyBytes(signer.Public())
	sigScheme := uint64(signer.Scheme())

	const fieldNumObject = 1
	const fieldNumSignature = 2

	sigSize := protowire.SizeTag(fieldNumSigPubKey) + protowire.SizeBytes(len(bPubKey)) +
		protowire.SizeTag(fieldNumSigVal) + +protowire.SizeBytes(len(sigTmpl)) +
		protowire.SizeTag(fieldNumSigScheme) + protowire.SizeVarint(sigScheme)

	msgSize := protowire.SizeTag(fieldNumObject) + protowire.SizeVarint(objSize) +
		protowire.SizeTag(fieldNumSignature) + sigSize

	// TODO(#544): support external buffers
	msg := make([]byte, 0, uint64(msgSize)+objSize)

	msg = protowire.AppendTag(msg, fieldNumObject, protowire.BytesType)
	msg = protowire.AppendVarint(msg, objSize)
	msg = msg[:uint64(len(msg))+objSize]

	bufObj := msg[uint64(len(msg))-objSize:]
	_, err = io.ReadFull(src, bufObj)
	if err != nil {
		return nil, fmt.Errorf("read full object into the buffer: %w", err)
	}

	objSig, err := signer.Sign(bufObj)
	if err != nil {
		return nil, fmt.Errorf("sign object: %w", err)
	}

	msg = protowire.AppendTag(msg, fieldNumSignature, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(sigSize))
	msg = protowire.AppendTag(msg, fieldNumSigPubKey, protowire.BytesType)
	msg = protowire.AppendBytes(msg, bPubKey)
	msg = protowire.AppendTag(msg, fieldNumSigVal, protowire.BytesType)
	msg = protowire.AppendBytes(msg, objSig)
	msg = protowire.AppendTag(msg, fieldNumSigScheme, protowire.VarintType)
	msg = protowire.AppendVarint(msg, sigScheme)

	return msg, nil
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
