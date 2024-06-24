package object

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldSplitParent
	fieldSplitPrevious
	fieldSplitParentSignature
	fieldSplitParentHeader
	fieldSplitChildren
	fieldSplitID
	fieldSplitFirst
)

func (x *Header_Split) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldSplitParent, x.Parent) +
			proto.SizeNested(fieldSplitPrevious, x.Previous) +
			proto.SizeNested(fieldSplitParentSignature, x.ParentSignature) +
			proto.SizeNested(fieldSplitParentHeader, x.ParentHeader) +
			proto.SizeBytes(fieldSplitID, x.SplitId) +
			proto.SizeNested(fieldSplitFirst, x.First)
		for i := range x.Children {
			sz += proto.SizeNested(fieldSplitChildren, x.Children[i])
		}
	}
	return sz
}

func (x *Header_Split) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldSplitParent, x.Parent)
		off += proto.MarshalNested(b[off:], fieldSplitPrevious, x.Previous)
		off += proto.MarshalNested(b[off:], fieldSplitParentSignature, x.ParentSignature)
		off += proto.MarshalNested(b[off:], fieldSplitParentHeader, x.ParentHeader)
		off += proto.MarshalBytes(b[off:], fieldSplitID, x.SplitId)
		for i := range x.Children {
			off += proto.MarshalNested(b[off:], fieldSplitChildren, x.Children[i])
		}
		proto.MarshalNested(b[off:], fieldSplitFirst, x.First)
	}
}

const (
	_ = iota
	fieldAttributeKey
	fieldAttributeValue
)

func (x *Header_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldAttributeKey, x.Key) +
			proto.SizeBytes(fieldAttributeValue, x.Value)
	}
	return sz
}

func (x *Header_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldAttributeKey, x.Key)
		proto.MarshalBytes(b[off:], fieldAttributeValue, x.Value)
	}
}

const (
	_ = iota
	fieldShortHeaderVersion
	fieldShortHeaderCreationEpoch
	fieldShortHeaderOwner
	fieldShortHeaderType
	fieldShortHeaderLen
	fieldShortHeaderChecksum
	fieldShortHeaderHomomorphic
)

func (x *ShortHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldShortHeaderVersion, x.Version) +
			proto.SizeVarint(fieldShortHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeNested(fieldShortHeaderOwner, x.OwnerId) +
			proto.SizeVarint(fieldShortHeaderType, int32(x.ObjectType)) +
			proto.SizeVarint(fieldShortHeaderLen, x.PayloadLength) +
			proto.SizeNested(fieldShortHeaderChecksum, x.PayloadHash) +
			proto.SizeNested(fieldShortHeaderHomomorphic, x.HomomorphicHash)
	}
	return sz
}

func (x *ShortHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldShortHeaderVersion, x.Version)
		off += proto.MarshalVarint(b[off:], fieldShortHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalNested(b[off:], fieldShortHeaderOwner, x.OwnerId)
		off += proto.MarshalVarint(b[off:], fieldShortHeaderType, int32(x.ObjectType))
		off += proto.MarshalVarint(b[off:], fieldShortHeaderLen, x.PayloadLength)
		off += proto.MarshalNested(b[off:], fieldShortHeaderChecksum, x.PayloadHash)
		off += proto.MarshalNested(b[off:], fieldShortHeaderHomomorphic, x.HomomorphicHash)
	}
}

const (
	_ = iota
	fieldHeaderVersion
	fieldHeaderContainer
	fieldHeaderOwner
	fieldHeaderCreationEpoch
	fieldHeaderLen
	fieldHeaderChecksum
	fieldHeaderType
	fieldHeaderHomomorphic
	fieldHeaderSession
	fieldHeaderAttributes
	fieldHeaderSplit
)

func (x *Header) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldHeaderVersion, x.Version) +
			proto.SizeNested(fieldHeaderContainer, x.ContainerId) +
			proto.SizeNested(fieldHeaderOwner, x.OwnerId) +
			proto.SizeVarint(fieldHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeVarint(fieldHeaderLen, x.PayloadLength) +
			proto.SizeNested(fieldHeaderChecksum, x.PayloadHash) +
			proto.SizeVarint(fieldHeaderType, int32(x.ObjectType)) +
			proto.SizeNested(fieldHeaderHomomorphic, x.HomomorphicHash) +
			proto.SizeNested(fieldHeaderSession, x.SessionToken) +
			proto.SizeNested(fieldHeaderSplit, x.Split)
		for i := range x.Attributes {
			sz += proto.SizeNested(fieldHeaderAttributes, x.Attributes[i])
		}
	}
	return sz
}

func (x *Header) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldHeaderVersion, x.Version)
		off += proto.MarshalNested(b[off:], fieldHeaderContainer, x.ContainerId)
		off += proto.MarshalNested(b[off:], fieldHeaderOwner, x.OwnerId)
		off += proto.MarshalVarint(b[off:], fieldHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalVarint(b[off:], fieldHeaderLen, x.PayloadLength)
		off += proto.MarshalNested(b[off:], fieldHeaderChecksum, x.PayloadHash)
		off += proto.MarshalVarint(b[off:], fieldHeaderType, int32(x.ObjectType))
		off += proto.MarshalNested(b[off:], fieldHeaderHomomorphic, x.HomomorphicHash)
		off += proto.MarshalNested(b[off:], fieldHeaderSession, x.SessionToken)
		for i := range x.Attributes {
			off += proto.MarshalNested(b[off:], fieldHeaderAttributes, x.Attributes[i])
		}
		proto.MarshalNested(b[off:], fieldHeaderSplit, x.Split)
	}
}

const (
	_ = iota
	fieldObjectID
	fieldObjectSignature
	fieldObjectHeader
	fieldObjectPayload
)

func (x *Object) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldObjectID, x.ObjectId) +
			proto.SizeNested(fieldObjectSignature, x.Signature) +
			proto.SizeNested(fieldObjectHeader, x.Header) +
			proto.SizeBytes(fieldObjectPayload, x.Payload)
	}
	return sz
}

func (x *Object) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldObjectID, x.ObjectId)
		off += proto.MarshalNested(b[off:], fieldObjectSignature, x.Signature)
		off += proto.MarshalNested(b[off:], fieldObjectHeader, x.Header)
		proto.MarshalBytes(b[off:], fieldObjectPayload, x.Payload)
	}
}

const (
	_ = iota
	fieldSplitInfoID
	fieldSplitInfoLast
	fieldSplitInfoLink
	fieldSplitInfoFirst
)

func (x *SplitInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSplitInfoID, x.SplitId) +
			proto.SizeNested(fieldSplitInfoLast, x.LastPart) +
			proto.SizeNested(fieldSplitInfoLink, x.Link) +
			proto.SizeNested(fieldSplitInfoFirst, x.FirstPart)
	}
	return sz
}

func (x *SplitInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldSplitInfoID, x.SplitId)
		off += proto.MarshalNested(b[off:], fieldSplitInfoLast, x.LastPart)
		off += proto.MarshalNested(b[off:], fieldSplitInfoLink, x.Link)
		proto.MarshalNested(b[off:], fieldSplitInfoFirst, x.FirstPart)
	}
}

const (
	_ = iota
	fieldGetReqAddress
	fieldGetReqRaw
)

func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetReqAddress, x.Address) +
			proto.SizeBool(fieldGetReqRaw, x.Raw)
	}
	return sz
}

func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldGetReqAddress, x.Address)
		proto.MarshalBool(b[off:], fieldGetReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldGetRespInitID
	fieldGetRespInitSignature
	fieldGetRespInitHeader
)

func (x *GetResponse_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetRespInitID, x.ObjectId) +
			proto.SizeNested(fieldGetRespInitSignature, x.Signature) +
			proto.SizeNested(fieldGetRespInitHeader, x.Header)
	}
	return sz
}

func (x *GetResponse_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldGetRespInitID, x.ObjectId)
		off += proto.MarshalNested(b[off:], fieldGetRespInitSignature, x.Signature)
		proto.MarshalNested(b[off:], fieldGetRespInitHeader, x.Header)
	}
}

const (
	_ = iota
	fieldGetRespInit
	fieldGetRespChunk
	fieldGetRespSplitInfo
)

func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *GetResponse_Body_Init_:
			if p != nil {
				sz = proto.SizeNested(fieldGetRespInit, p.Init)
			}
		case *GetResponse_Body_Chunk:
			if p != nil {
				sz = proto.SizeBytes(fieldGetRespChunk, p.Chunk)
			}
		case *GetResponse_Body_SplitInfo:
			if p != nil {
				sz = proto.SizeNested(fieldGetRespSplitInfo, p.SplitInfo)
			}
		}
	}
	return sz
}

func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *GetResponse_Body_Init_:
			if p != nil {
				proto.MarshalNested(b, fieldGetRespInit, p.Init)
			}
		case *GetResponse_Body_Chunk:
			if p != nil {
				proto.MarshalBytes(b, fieldGetRespChunk, p.Chunk)
			}
		case *GetResponse_Body_SplitInfo:
			if p != nil {
				proto.MarshalNested(b, fieldGetRespSplitInfo, p.SplitInfo)
			}
		}
	}
}

const (
	_ = iota
	fieldHeadReqAddress
	fieldHeadReqMain
	fieldHeadReqRaw
)

func (x *HeadRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldHeadReqAddress, x.Address) +
			proto.SizeBool(fieldHeadReqMain, x.MainOnly) +
			proto.SizeBool(fieldHeadReqRaw, x.Raw)
	}
	return sz
}

func (x *HeadRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldHeadReqAddress, x.Address)
		off += proto.MarshalBool(b[off:], fieldHeadReqMain, x.MainOnly)
		proto.MarshalBool(b[off:], fieldHeadReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldHeaderSignatureHdr
	fieldHeaderSignatureSig
)

func (x *HeaderWithSignature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldHeaderSignatureHdr, x.Header) +
			proto.SizeNested(fieldHeaderSignatureSig, x.Signature)
	}
	return sz
}

func (x *HeaderWithSignature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldHeaderSignatureHdr, x.Header)
		proto.MarshalNested(b[off:], fieldHeaderSignatureSig, x.Signature)
	}
}

const (
	_ = iota
	fieldHeadRespHeader
	fieldHeadRespShort
	fieldHeadRespSplitInfo
)

func (x *HeadResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch h := x.Head.(type) {
		default:
			panic(fmt.Sprintf("unexpected head part %T", x.Head))
		case nil:
		case *HeadResponse_Body_Header:
			if h != nil {
				sz = proto.SizeNested(fieldHeadRespHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				sz = proto.SizeNested(fieldHeadRespShort, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				sz = proto.SizeNested(fieldHeadRespSplitInfo, h.SplitInfo)
			}
		}
	}
	return sz
}

func (x *HeadResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch h := x.Head.(type) {
		default:
			panic(fmt.Sprintf("unexpected head part %T", x.Head))
		case nil:
		case *HeadResponse_Body_Header:
			if h != nil {
				proto.MarshalNested(b, fieldHeadRespHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				proto.MarshalNested(b, fieldHeadRespShort, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				proto.MarshalNested(b, fieldHeadRespSplitInfo, h.SplitInfo)
			}
		}
	}
}

const (
	_ = iota
	fieldRangeOff
	fieldRangeLen
)

func (x *Range) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldRangeOff, x.Offset) +
			proto.SizeVarint(fieldRangeLen, x.Length)
	}
	return sz
}

func (x *Range) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldRangeOff, x.Offset)
		proto.MarshalVarint(b[off:], fieldRangeLen, x.Length)
	}
}

const (
	_ = iota
	fieldRangeReqAddress
	fieldRangeReqRange
	fieldRangeReqRaw
)

func (x *GetRangeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldRangeReqAddress, x.Address) +
			proto.SizeNested(fieldRangeReqRange, x.Range) +
			proto.SizeBool(fieldRangeReqRaw, x.Raw)
	}
	return sz
}

func (x *GetRangeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldRangeReqAddress, x.Address)
		off += proto.MarshalNested(b[off:], fieldRangeReqRange, x.Range)
		proto.MarshalBool(b[off:], fieldRangeReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldPutReqInitID
	fieldPutReqInitSignature
	fieldPutReqInitHeader
	fieldPutReqInitCopies
)

const (
	_ = iota
	fieldRangeRespChunk
	fieldRangeRespSplitInfo
)

func (x *GetRangeResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch p := x.RangePart.(type) {
		default:
			panic(fmt.Sprintf("unexpected range part %T", x.RangePart))
		case nil:
		case *GetRangeResponse_Body_Chunk:
			if p != nil {
				sz = proto.SizeBytes(fieldRangeRespChunk, p.Chunk)
			}
		case *GetRangeResponse_Body_SplitInfo:
			if p != nil {
				sz = proto.SizeNested(fieldRangeRespSplitInfo, p.SplitInfo)
			}
		}
	}
	return sz
}

func (x *GetRangeResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.RangePart.(type) {
		default:
			panic(fmt.Sprintf("unexpected range part %T", x.RangePart))
		case nil:
		case *GetRangeResponse_Body_Chunk:
			if p != nil {
				proto.MarshalBytes(b, fieldRangeRespChunk, p.Chunk)
			}
		case *GetRangeResponse_Body_SplitInfo:
			if p != nil {
				proto.MarshalNested(b, fieldRangeRespSplitInfo, p.SplitInfo)
			}
		}
	}
}

const (
	_ = iota
	fieldRangeHashReqAddress
	fieldRangeHashReqRanges
	fieldRangeHashReqSalt
	fieldRangeHashReqType
)

func (x *GetRangeHashRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldRangeHashReqAddress, x.Address) +
			proto.SizeBytes(fieldRangeHashReqSalt, x.Salt) +
			proto.SizeVarint(fieldRangeHashReqType, int32(x.Type))
		for i := range x.Ranges {
			sz += proto.SizeNested(fieldRangeHashReqRanges, x.Ranges[i])
		}
	}
	return sz
}

func (x *GetRangeHashRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldRangeHashReqAddress, x.Address)
		for i := range x.Ranges {
			off += proto.MarshalNested(b[off:], fieldRangeHashReqRanges, x.Ranges[i])
		}
		off += proto.MarshalBytes(b[off:], fieldRangeHashReqSalt, x.Salt)
		proto.MarshalVarint(b[off:], fieldRangeHashReqType, int32(x.Type))
	}
}

const (
	_ = iota
	fieldRangeHashRespType
	fieldRangeHashRespHashes
)

func (x *GetRangeHashResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldRangeHashRespType, int32(x.Type)) +
			proto.SizeRepeatedBytes(fieldRangeHashRespHashes, x.HashList)
	}
	return sz
}

func (x *GetRangeHashResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldRangeHashRespType, int32(x.Type))
		proto.MarshalRepeatedBytes(b[off:], fieldRangeHashRespHashes, x.HashList)
	}
}

func (x *PutRequest_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldPutReqInitID, x.ObjectId) +
			proto.SizeNested(fieldPutReqInitSignature, x.Signature) +
			proto.SizeNested(fieldPutReqInitHeader, x.Header) +
			proto.SizeVarint(fieldPutReqInitCopies, x.CopiesNumber)
	}
	return sz
}

func (x *PutRequest_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldPutReqInitID, x.ObjectId)
		off += proto.MarshalNested(b[off:], fieldPutReqInitSignature, x.Signature)
		off += proto.MarshalNested(b[off:], fieldPutReqInitHeader, x.Header)
		proto.MarshalVarint(b[off:], fieldPutReqInitCopies, x.CopiesNumber)
	}
}

const (
	_ = iota
	fieldPutReqInit
	fieldPutReqChunk
)

func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *PutRequest_Body_Init_:
			sz = proto.SizeNested(fieldPutReqInit, p.Init)
		case *PutRequest_Body_Chunk:
			sz = proto.SizeBytes(fieldPutReqChunk, p.Chunk)
		}
	}
	return sz
}

func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *PutRequest_Body_Init_:
			proto.MarshalNested(b, fieldPutReqInit, p.Init)
		case *PutRequest_Body_Chunk:
			proto.MarshalBytes(b, fieldPutReqChunk, p.Chunk)
		}
	}
}

const (
	_ = iota
	fieldPutRespID
)

func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldPutRespID, x.ObjectId)
	}
	return sz
}

func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldPutRespID, x.ObjectId)
	}
}

const (
	_ = iota
	fieldDeleteReqAddress
)

func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldDeleteReqAddress, x.Address)
	}
	return sz
}

func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldDeleteReqAddress, x.Address)
	}
}

const (
	_ = iota
	fieldDeleteRespTombstone
)

func (x *DeleteResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldDeleteRespTombstone, x.Tombstone)
	}
	return sz
}

func (x *DeleteResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldDeleteRespTombstone, x.Tombstone)
	}
}

const (
	_ = iota
	fieldSearchFilterMatcher
	fieldSearchFilterKey
	fieldSearchFilterValue
)

func (x *SearchRequest_Body_Filter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldSearchFilterMatcher, int32(x.MatchType))
		sz += proto.SizeBytes(fieldSearchFilterKey, x.Key)
		sz += proto.SizeBytes(fieldSearchFilterValue, x.Value)
	}
	return sz
}

func (x *SearchRequest_Body_Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldSearchFilterMatcher, int32(x.MatchType))
		off += proto.MarshalBytes(b[off:], fieldSearchFilterKey, x.Key)
		proto.MarshalBytes(b[off:], fieldSearchFilterValue, x.Value)
	}
}

const (
	_ = iota
	fieldSearchReqContainer
	fieldSearchReqVersion
	fieldSearchReqFilters
)

func (x *SearchRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldSearchReqContainer, x.ContainerId)
		sz += proto.SizeVarint(fieldSearchReqVersion, x.Version)
		for i := range x.Filters {
			sz += proto.SizeNested(fieldSearchReqFilters, x.Filters[i])
		}
	}
	return sz
}

func (x *SearchRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldSearchReqContainer, x.ContainerId)
		off += proto.MarshalVarint(b[off:], fieldSearchReqVersion, x.Version)
		for i := range x.Filters {
			off += proto.MarshalNested(b[off:], fieldSearchReqFilters, x.Filters[i])
		}
	}
}

const (
	_ = iota
	fieldSearchRespIDList
)

func (x *SearchResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.IdList {
			sz += proto.SizeNested(fieldSearchRespIDList, x.IdList[i])
		}
	}
	return sz
}

func (x *SearchResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.IdList {
			off += proto.MarshalNested(b[off:], fieldSearchRespIDList, x.IdList[i])
		}
	}
}
