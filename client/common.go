package client

import (
	"context"
	"fmt"
	"time"

	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// Various field numbers in from NeoFS API definitions.
const (
	fieldNumSigPubKey = 1
	fieldNumSigVal    = 2
	fieldNumSigScheme = 3
)

const (
	localRequestTTL   = 1
	defaultRequestTTL = 2
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

func writeXHeadersToMeta(xHeaders []string, h *protosession.RequestMetaHeader) {
	if len(xHeaders) == 0 {
		return
	}

	if len(xHeaders)%2 != 0 {
		panic("slice of X-Headers with odd length")
	}

	h.XHeaders = make([]*protosession.XHeader, len(xHeaders)/2)
	j := 0

	for i := 0; i < len(xHeaders); i += 2 {
		h.XHeaders[j] = &protosession.XHeader{Key: xHeaders[i], Value: xHeaders[i+1]}
		j++
	}
}

type onlyBinarySendingCodec struct{}

func (x onlyBinarySendingCodec) Name() string {
	// may be any non-empty, conflicts are unlikely to arise
	return "neofs_binary_sender"
}

func (x onlyBinarySendingCodec) Marshal(msg any) ([]byte, error) {
	bMsg, ok := msg.([]byte)
	if ok {
		return bMsg, nil
	}

	return nil, fmt.Errorf("message is not of type %T", bMsg)
}

func (x onlyBinarySendingCodec) Unmarshal(raw []byte, msg any) error {
	return encoding.GetCodec(proto.Name).Unmarshal(raw, msg)
}

// Tries to make an action within given timeout canceling the context at
// expiration.
//
// Copy-pasted from https://github.com/nspcc-dev/neofs-api-go/blob/4d4eaa29436e2b1ce9bcdddd6551133c388a1cdb/rpc/grpc/init.go#L53.
// TODO: https://github.com/nspcc-dev/neofs-sdk-go/issues/640.
func dowithTimeout(timeout time.Duration, cancel context.CancelFunc, action func() error) error {
	ch := make(chan error, 1)
	go func() {
		ch <- action()
		close(ch)
	}()

	tt := time.NewTimer(timeout)

	select {
	case err := <-ch:
		return err
	case <-tt.C:
		cancel()
		return context.DeadlineExceeded
	}
}
