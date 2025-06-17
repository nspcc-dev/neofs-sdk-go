package object_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

// returns random object.Range with all non-zero fields.
func randRange() *object.Range {
	return &object.Range{Offset: prototest.RandUint64(), Length: prototest.RandUint64()}
}

// returns non-empty list of object.Range up to 10 elements. Each element may be
// nil and pointer to zero.
func randRanges() []*object.Range { return prototest.RandRepeated(randRange) }

// returns random object.Header_Attribute with all non-zero fields.
func randAttribute() *object.Header_Attribute {
	return &object.Header_Attribute{
		Key: prototest.RandString(), Value: prototest.RandString(),
	}
}

// returns non-empty list of object.Header_Attribute up to 10 elements. Each
// element may be nil and pointer to zero.
func randAttributes() []*object.Header_Attribute { return prototest.RandRepeated(randAttribute) }

// returns random object.ShortHeader with all non-zero fields.
func randShortHeader() *object.ShortHeader {
	return &object.ShortHeader{
		Version:         prototest.RandVersion(),
		CreationEpoch:   prototest.RandUint64(),
		OwnerId:         prototest.RandOwnerID(),
		ObjectType:      prototest.RandInteger[object.ObjectType](),
		PayloadLength:   prototest.RandUint64(),
		PayloadHash:     prototest.RandChecksum(),
		HomomorphicHash: prototest.RandChecksum(),
	}
}

func _randHeader(withParentHdr bool) *object.Header {
	return &object.Header{
		Version:         prototest.RandVersion(),
		ContainerId:     prototest.RandContainerID(),
		OwnerId:         prototest.RandOwnerID(),
		CreationEpoch:   prototest.RandUint64(),
		PayloadLength:   prototest.RandUint64(),
		PayloadHash:     prototest.RandChecksum(),
		ObjectType:      prototest.RandInteger[object.ObjectType](),
		HomomorphicHash: prototest.RandChecksum(),
		SessionToken:    prototest.RandSessionToken(),
		Attributes:      randAttributes(),
		Split:           _randSplitHeader(withParentHdr),
	}
}

// returns random object.Header with all non-zero fields.
func randHeader() *object.Header { return _randHeader(true) }

// returns random object.HeaderWithSignature with all non-zero fields.
func randHeaderWithSignature() *object.HeaderWithSignature {
	return &object.HeaderWithSignature{
		Header:    randHeader(),
		Signature: prototest.RandSignature(),
	}
}

func _randSplitHeader(withParentHdr bool) *object.Header_Split {
	v := &object.Header_Split{
		Parent:          prototest.RandObjectID(),
		Previous:        prototest.RandObjectID(),
		ParentSignature: prototest.RandSignature(),
		Children:        prototest.RandObjectIDs(),
		SplitId:         prototest.RandBytes(),
		First:           prototest.RandObjectID(),
	}
	if withParentHdr {
		v.ParentHeader = _randHeader(false)
	}
	return v
}

// returns random object.Header_Split with all non-zero fields.
func randSplitHeader() *object.Header_Split { return _randSplitHeader(true) }

// returns random object.SplitInfo with all non-zero fields.
func randSplitInfo() *object.SplitInfo {
	return &object.SplitInfo{
		SplitId:   prototest.RandBytes(),
		LastPart:  prototest.RandObjectID(),
		Link:      prototest.RandObjectID(),
		FirstPart: prototest.RandObjectID(),
	}
}

// returns random object.GetResponse_Body_Init with all non-zero fields.
func randGetResponseInit() *object.GetResponse_Body_Init {
	return &object.GetResponse_Body_Init{
		ObjectId:  prototest.RandObjectID(),
		Signature: prototest.RandSignature(),
		Header:    randHeader(),
	}
}

// returns random object.PutRequest_Body_Init with all non-zero fields.
func randPutRequestInit() *object.PutRequest_Body_Init {
	return &object.PutRequest_Body_Init{
		ObjectId:     prototest.RandObjectID(),
		Signature:    prototest.RandSignature(),
		Header:       randHeader(),
		CopiesNumber: prototest.RandUint32(),
	}
}

// returns random object.SearchFilter with all non-zero fields.
func randSearchFilter() *object.SearchFilter {
	return &object.SearchFilter{
		MatchType: prototest.RandInteger[object.MatchType](),
		Key:       prototest.RandString(),
		Value:     prototest.RandString(),
	}
}

// returns non-empty list of object.SearchFilter up to 10 elements.
// Each element may be nil and pointer to zero.
func randSearchFilters() []*object.SearchFilter {
	return prototest.RandRepeated(randSearchFilter)
}

func TestHeader_Attribute_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.Header_Attribute{
		randAttribute(),
	})
}

func TestRange_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.Range{
		randRange(),
	})
}

func TestShortHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.ShortHeader{
		randShortHeader(),
	})
}

func TestHeader_Split_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.Header_Split{
			Children: []*refs.ObjectID{nil, {}},
		}

		var dst object.Header_Split
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		cs := dst.GetChildren()
		require.Len(t, cs, 2)
		require.Equal(t, cs[0], new(refs.ObjectID))
		require.Equal(t, cs[1], new(refs.ObjectID))
	})

	prototest.TestMarshalStable(t, []*object.Header_Split{
		randSplitHeader(),
	})
}

func TestHeader_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.Header{
			Attributes: []*object.Header_Attribute{nil, {}},
		}

		var dst object.Header
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		as := dst.GetAttributes()
		require.Len(t, as, 2)
		require.Equal(t, as[0], new(object.Header_Attribute))
		require.Equal(t, as[1], new(object.Header_Attribute))
	})

	prototest.TestMarshalStable(t, []*object.Header{
		randHeader(),
	})
}

func TestObject_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.Object{
		{
			ObjectId:  prototest.RandObjectID(),
			Signature: prototest.RandSignature(),
			Header:    randHeader(),
			Payload:   prototest.RandBytes(),
		},
	})
}

func TestSplitInfo_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.SplitInfo{
		randSplitInfo(),
	})
}

func TestGetRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.GetRequest_Body{
		{Address: prototest.RandObjectAddress(), Raw: false},
		{Address: prototest.RandObjectAddress(), Raw: true},
	})
}

func TestGetResponse_Body_Init_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.GetResponse_Body_Init{
		randGetResponseInit(),
	})
}

func TestGetResponse_Body_MarshalStable(t *testing.T) {
	var vEmpty object.GetResponse_Body
	var vInit object.GetResponse_Body
	vInit.Init = randGetResponseInit()
	var vChunk object.GetResponse_Body
	vChunk.Chunk = prototest.RandBytes()
	var vSplitInfo object.GetResponse_Body
	vSplitInfo.SplitInfo = randSplitInfo()
	prototest.TestMarshalStable(t, []*object.GetResponse_Body{
		&vEmpty,
		&vInit,
		&vChunk,
		&vSplitInfo,
	})
}

func TestHeadRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.HeadRequest_Body{
		{Address: prototest.RandObjectAddress(), MainOnly: false, Raw: false},
		{Address: prototest.RandObjectAddress(), MainOnly: true, Raw: false},
		{Address: prototest.RandObjectAddress(), MainOnly: false, Raw: true},
		{Address: prototest.RandObjectAddress(), MainOnly: true, Raw: true},
	})
}

func TestHeaderWithSignature_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.HeaderWithSignature{
		randHeaderWithSignature(),
	})
}

func TestHeadResponse_Body_MarshalStable(t *testing.T) {
	var vEmpty object.HeadResponse_Body
	var vFull object.HeadResponse_Body
	vFull.Head = &object.HeadResponse_Body_Header{Header: randHeaderWithSignature()}
	var vShort object.HeadResponse_Body
	vShort.Head = &object.HeadResponse_Body_ShortHeader{ShortHeader: randShortHeader()}
	var vSplitInfo object.HeadResponse_Body
	vSplitInfo.Head = &object.HeadResponse_Body_SplitInfo{SplitInfo: randSplitInfo()}
	prototest.TestMarshalStable(t, []*object.HeadResponse_Body{
		&vEmpty,
		&vFull,
		&vShort,
		&vSplitInfo,
	})
}

func TestGetRangeRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.GetRangeRequest_Body{
		{Address: prototest.RandObjectAddress(), Range: randRange(), Raw: false},
		{Address: prototest.RandObjectAddress(), Range: randRange(), Raw: true},
	})
}

func TestGetRangeResponse_Body_MarshalStable(t *testing.T) {
	var vEmpty object.GetRangeResponse_Body
	var vChunk object.GetRangeResponse_Body
	vChunk.RangePart = &object.GetRangeResponse_Body_Chunk{Chunk: prototest.RandBytes()}
	var vSplitInfo object.GetRangeResponse_Body
	vSplitInfo.RangePart = &object.GetRangeResponse_Body_SplitInfo{SplitInfo: randSplitInfo()}
	prototest.TestMarshalStable(t, []*object.GetRangeResponse_Body{
		&vEmpty,
		&vChunk,
		&vSplitInfo,
	})
}

func TestGetRangeHashRequest_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.GetRangeHashRequest_Body{
			Ranges: []*object.Range{nil, {}},
		}

		var dst object.GetRangeHashRequest_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		rs := dst.GetRanges()
		require.Len(t, rs, 2)
		require.Equal(t, rs[0], new(object.Range))
		require.Equal(t, rs[1], new(object.Range))
	})

	prototest.TestMarshalStable(t, []*object.GetRangeHashRequest_Body{
		{
			Address: prototest.RandObjectAddress(),
			Ranges:  randRanges(),
			Salt:    prototest.RandBytes(),
			Type:    prototest.RandInteger[refs.ChecksumType](),
		},
	})
}

func TestGetRangeHashResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.GetRangeHashResponse_Body{
		{
			Type:     prototest.RandInteger[refs.ChecksumType](),
			HashList: prototest.RandRepeatedBytes(),
		},
	})
}

func TestPutRequest_Body_Init_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.PutRequest_Body_Init{
		randPutRequestInit(),
	})
}

func TestPutRequest_Body_MarshalStable(t *testing.T) {
	var vEmpty object.PutRequest_Body
	var vInit object.PutRequest_Body
	vInit.ObjectPart = &object.PutRequest_Body_Init_{Init: randPutRequestInit()}
	var vChunk object.PutRequest_Body
	vChunk.ObjectPart = &object.PutRequest_Body_Chunk{Chunk: prototest.RandBytes()}
	prototest.TestMarshalStable(t, []*object.PutRequest_Body{
		&vEmpty,
		&vInit,
		&vChunk,
	})
}

func TestPutResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.PutResponse_Body{
		{ObjectId: prototest.RandObjectID()},
	})
}

func TestDeleteRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.DeleteRequest_Body{
		{Address: prototest.RandObjectAddress()},
	})
}

func TestDeleteResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.DeleteResponse_Body{
		{Tombstone: prototest.RandObjectAddress()},
	})
}

func TestSearchFilter_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.SearchRequest_Body{
			Filters: []*object.SearchFilter{nil, {}},
		}

		var dst object.SearchRequest_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		fs := dst.GetFilters()
		require.Len(t, fs, 2)
		require.Equal(t, fs[0], new(object.SearchFilter))
		require.Equal(t, fs[1], new(object.SearchFilter))
	})

	prototest.TestMarshalStable(t, []*object.SearchFilter{
		randSearchFilter(),
	})
}

func TestSearchRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.SearchRequest_Body{
		{
			ContainerId: prototest.RandContainerID(),
			Version:     prototest.RandUint32(),
			Filters:     randSearchFilters(),
		},
	})
}

func TestSearchResponse_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.SearchResponse_Body{
			IdList: []*refs.ObjectID{nil, {}},
		}

		var dst object.SearchResponse_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ids := dst.GetIdList()
		require.Len(t, ids, 2)
		require.Equal(t, ids[0], new(refs.ObjectID))
		require.Equal(t, ids[1], new(refs.ObjectID))
	})

	prototest.TestMarshalStable(t, []*object.SearchResponse_Body{
		{IdList: prototest.RandObjectIDs()},
	})
}

func TestSearchV2Request_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*object.SearchV2Request_Body{
		{
			ContainerId: prototest.RandContainerID(),
			Version:     prototest.RandUint32(),
			Filters:     randSearchFilters(),
			Cursor:      prototest.RandString(),
			Count:       prototest.RandUint32(),
			Attributes:  prototest.RandStrings(),
		},
	})
}

func TestSearchV2Response_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &object.SearchV2Response_Body{
			Result: []*object.SearchV2Response_OIDWithMeta{nil, {}},
		}

		var dst object.SearchV2Response_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ids := dst.GetResult()
		require.Len(t, ids, 2)
		require.Equal(t, ids[0], new(object.SearchV2Response_OIDWithMeta))
		require.Equal(t, ids[1], new(object.SearchV2Response_OIDWithMeta))
	})

	prototest.TestMarshalStable(t, []*object.SearchV2Response_Body{
		{
			Result: []*object.SearchV2Response_OIDWithMeta{
				{
					Id:         prototest.RandObjectID(),
					Attributes: prototest.RandStrings(),
				},
				{
					Id:         prototest.RandObjectID(),
					Attributes: prototest.RandStrings(),
				},
			},
			Cursor: prototest.RandString(),
		},
	})
}
