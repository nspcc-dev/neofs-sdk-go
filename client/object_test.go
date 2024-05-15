package client

import (
	"context"
	"math"

	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
)

type noOtherObjectCalls struct{}

func (noOtherObjectCalls) Get(*apiobject.GetRequest, apiobject.ObjectService_GetServer) error {
	panic("must not be called")
}

func (noOtherObjectCalls) Put(apiobject.ObjectService_PutServer) error {
	panic("must not be called")
}

func (noOtherObjectCalls) Delete(context.Context, *apiobject.DeleteRequest) (*apiobject.DeleteResponse, error) {
	panic("must not be called")
}

func (noOtherObjectCalls) Head(context.Context, *apiobject.HeadRequest) (*apiobject.HeadResponse, error) {
	panic("must not be called")
}

func (noOtherObjectCalls) Search(*apiobject.SearchRequest, apiobject.ObjectService_SearchServer) error {
	panic("must not be called")
}

func (noOtherObjectCalls) GetRange(*apiobject.GetRangeRequest, apiobject.ObjectService_GetRangeServer) error {
	panic("must not be called")
}

func (noOtherObjectCalls) GetRangeHash(context.Context, *apiobject.GetRangeHashRequest) (*apiobject.GetRangeHashResponse, error) {
	panic("must not be called")
}

func (noOtherObjectCalls) Replicate(context.Context, *apiobject.ReplicateRequest) (*apiobject.ReplicateResponse, error) {
	panic("must not be called")
}

var invalidObjectHeaderTestCases = []struct {
	err     string
	corrupt func(*apiobject.Header)
}{
	{err: "invalid type field -1", corrupt: func(h *apiobject.Header) { h.ObjectType = -1 }},
	{err: "missing container", corrupt: func(h *apiobject.Header) { h.ContainerId = nil }},
	{err: "invalid container: missing value field", corrupt: func(h *apiobject.Header) { h.ContainerId.Value = nil }},
	{err: "invalid container: invalid value length 31", corrupt: func(h *apiobject.Header) { h.ContainerId.Value = make([]byte, 31) }},
	{err: "missing owner", corrupt: func(h *apiobject.Header) { h.OwnerId = nil }},
	{err: "invalid owner: missing value field", corrupt: func(h *apiobject.Header) { h.OwnerId.Value = nil }},
	{err: "invalid owner: invalid value length 24", corrupt: func(h *apiobject.Header) { h.OwnerId.Value = make([]byte, 24) }},
	{err: "invalid owner: invalid prefix byte 0x42, expected 0x35", corrupt: func(h *apiobject.Header) { h.OwnerId.Value[0] = 0x42 }},
	{err: "invalid owner: value checksum mismatch", corrupt: func(h *apiobject.Header) { h.OwnerId.Value[len(h.OwnerId.Value)-1]++ }},
	{err: "invalid payload checksum: missing value", corrupt: func(h *apiobject.Header) { h.PayloadHash.Sum = nil }},
	{err: "invalid payload homomorphic checksum: missing value", corrupt: func(h *apiobject.Header) { h.HomomorphicHash.Sum = nil }},
	{err: "invalid session: missing token body", corrupt: func(h *apiobject.Header) { h.SessionToken.Body = nil }},
	{err: "invalid session: missing session ID", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Id = nil }},
	{err: "invalid session: missing session ID", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Id = []byte{} }},
	{err: "invalid session: invalid session ID: invalid UUID (got 15 bytes)", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Id = make([]byte, 15) }},
	{err: "invalid session: invalid session UUID version 3", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Id[6] = 3 << 4 }},
	{err: "invalid session: missing session issuer", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.OwnerId = nil }},
	{err: "invalid session: invalid session issuer: missing value field", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.OwnerId.Value = nil }},
	{err: "invalid session: invalid session issuer: missing value field", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.OwnerId.Value = []byte{} }},
	{err: "invalid session: invalid session issuer: invalid value length 26", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.OwnerId.Value = make([]byte, 26) }},
	{err: "invalid session: invalid session issuer: invalid prefix byte 0x43, expected 0x35", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.OwnerId.Value[0] = 0x43 }},
	{err: "invalid session: invalid session issuer: value checksum mismatch", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.OwnerId.Value[len(h.SessionToken.Body.OwnerId.Value)-1]++
	}},
	{err: "invalid session: missing token lifetime", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Lifetime = nil }},
	{err: "invalid session: missing session public key", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.SessionKey = nil }},
	{err: "invalid session: missing session public key", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.SessionKey = []byte{} }},
	{err: "invalid session: invalid body signature: missing public key", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Key = nil }},
	{err: "invalid session: invalid body signature: missing public key", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Key = []byte{} }},
	{err: "invalid session: invalid body signature: decode public key from binary", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Key = make([]byte, 32) }},
	{err: "invalid session: invalid body signature: missing signature", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Sign = nil }},
	{err: "invalid session: invalid body signature: missing signature", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Sign = []byte{} }},
	{err: "invalid session: invalid body signature: unsupported scheme 2147483647", corrupt: func(h *apiobject.Header) { h.SessionToken.Signature.Scheme = math.MaxInt32 }},
	{err: "invalid session: missing session context", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Context = nil }},
	{err: "invalid session: wrong context field", corrupt: func(h *apiobject.Header) { h.SessionToken.Body.Context = new(apisession.SessionToken_Body_Container) }},
	{err: "invalid session: invalid context: invalid target container: missing value field", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = nil
	}},
	{err: "invalid session: invalid context: invalid target container: missing value field", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = []byte{}
	}},
	{err: "invalid session: invalid context: invalid target container: invalid value length 31", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = make([]byte, 31)
	}},
	{err: "invalid session: invalid context: invalid target object #1: missing value field", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
			{Value: make([]byte, 32)}, {Value: nil}}
	}},
	{err: "invalid session: invalid context: invalid target object #1: missing value field", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
			{Value: make([]byte, 32)}, {Value: nil}}
	}},
	{err: "invalid session: invalid context: invalid target object #1: invalid value length 31", corrupt: func(h *apiobject.Header) {
		h.SessionToken.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
			{Value: make([]byte, 32)}, {Value: make([]byte, 31)}}
	}},
	{err: "invalid split-chain ID: wrong length 15", corrupt: func(h *apiobject.Header) { h.Split.SplitId = make([]byte, 15) }},
	{err: "invalid split-chain ID: wrong version #3", corrupt: func(h *apiobject.Header) {
		h.Split.SplitId = make([]byte, 16)
		h.Split.SplitId[6] = 3 << 4
	}},
	{err: "invalid parent ID: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Parent.Value = nil }},
	{err: "invalid parent ID: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Parent.Value = []byte{} }},
	{err: "invalid parent ID: invalid value length 31", corrupt: func(h *apiobject.Header) { h.Split.Parent.Value = make([]byte, 31) }},
	{err: "invalid previous split-chain element: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Previous.Value = nil }},
	{err: "invalid previous split-chain element: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Previous.Value = []byte{} }},
	{err: "invalid previous split-chain element: invalid value length 31", corrupt: func(h *apiobject.Header) { h.Split.Previous.Value = make([]byte, 31) }},
	{err: "invalid parent signature: missing public key", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Key = nil }},
	{err: "invalid parent signature: missing public key", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Key = []byte{} }},
	{err: "invalid parent signature: decode public key from binary", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Key = make([]byte, 32) }},
	{err: "invalid parent signature: missing signature", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Sign = nil }},
	{err: "invalid parent signature: missing signature", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Sign = []byte{} }},
	{err: "invalid parent signature: unsupported scheme 2147483647", corrupt: func(h *apiobject.Header) { h.Split.ParentSignature.Scheme = math.MaxInt32 }},
	{err: "invalid child split-chain element #1: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Children = []*refs.ObjectID{{Value: make([]byte, 32)}, nil} }},
	{err: "invalid child split-chain element #1: missing value field", corrupt: func(h *apiobject.Header) { h.Split.Children = []*refs.ObjectID{{Value: make([]byte, 32)}, {}} }},
	{err: "invalid child split-chain element #1: invalid value length 31", corrupt: func(h *apiobject.Header) {
		h.Split.Children = []*refs.ObjectID{{Value: make([]byte, 32)}, {Value: make([]byte, 31)}}
	}},
	{err: "invalid first split-chain element: missing value field", corrupt: func(h *apiobject.Header) { h.Split.First.Value = nil }},
	{err: "invalid first split-chain element: missing value field", corrupt: func(h *apiobject.Header) { h.Split.First.Value = []byte{} }},
	{err: "invalid first split-chain element: invalid value length 31", corrupt: func(h *apiobject.Header) { h.Split.First.Value = make([]byte, 31) }},
	{err: "invalid attribute #1: missing key", corrupt: func(h *apiobject.Header) {
		h.Attributes = []*apiobject.Header_Attribute{{Key: "k1", Value: "v1"}, {Key: "", Value: "v2"}}
	}},
	{err: "invalid attribute #1 (k2): missing value", corrupt: func(h *apiobject.Header) {
		h.Attributes = []*apiobject.Header_Attribute{{Key: "k1", Value: "v1"}, {Key: "k2", Value: ""}}
	}},
	{err: "multiple attributes with key=k2", corrupt: func(h *apiobject.Header) {
		h.Attributes = []*apiobject.Header_Attribute{{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v2"}, {Key: "k3", Value: "v3"}, {Key: "k2", Value: "v4"}}
	}},
	{err: "invalid expiration attribute (#1): invalid integer", corrupt: func(h *apiobject.Header) {
		h.Attributes = []*apiobject.Header_Attribute{{Key: "k1", Value: "v1"}, {Key: "__NEOFS__EXPIRATION_EPOCH", Value: "not a number"}}
	}},
	{err: "invalid timestamp attribute (#1): invalid integer", corrupt: func(h *apiobject.Header) {
		h.Attributes = []*apiobject.Header_Attribute{{Key: "k1", Value: "v1"}, {Key: "Timestamp", Value: "not a number"}}
	}},
}
