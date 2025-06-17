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

// MarshaledSize returns size of the Header_Split in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *Header_Split) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSplitParent, x.Parent) +
			proto.SizeEmbedded(fieldSplitPrevious, x.Previous) +
			proto.SizeEmbedded(fieldSplitParentSignature, x.ParentSignature) +
			proto.SizeEmbedded(fieldSplitParentHeader, x.ParentHeader) +
			proto.SizeBytes(fieldSplitID, x.SplitId) +
			proto.SizeEmbedded(fieldSplitFirst, x.First) +
			proto.SizeRepeatedMessages(fieldSplitChildren, x.Children)
	}
	return 0
}

// MarshalStable writes the Header_Split in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Header_Split.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header_Split) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSplitParent, x.Parent)
		off += proto.MarshalToEmbedded(b[off:], fieldSplitPrevious, x.Previous)
		off += proto.MarshalToEmbedded(b[off:], fieldSplitParentSignature, x.ParentSignature)
		off += proto.MarshalToEmbedded(b[off:], fieldSplitParentHeader, x.ParentHeader)
		off += proto.MarshalToBytes(b[off:], fieldSplitID, x.SplitId)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldSplitChildren, x.Children)
		proto.MarshalToEmbedded(b[off:], fieldSplitFirst, x.First)
	}
}

const (
	_ = iota
	fieldAttributeKey
	fieldAttributeValue
)

// MarshaledSize returns size of the Header_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Header_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldAttributeKey, x.Key) +
			proto.SizeBytes(fieldAttributeValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Header_Attribute in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Header_Attribute.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldAttributeKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldAttributeValue, x.Value)
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

// MarshaledSize returns size of the ShortHeader in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *ShortHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldShortHeaderVersion, x.Version) +
			proto.SizeVarint(fieldShortHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeEmbedded(fieldShortHeaderOwner, x.OwnerId) +
			proto.SizeVarint(fieldShortHeaderType, int32(x.ObjectType)) +
			proto.SizeVarint(fieldShortHeaderLen, x.PayloadLength) +
			proto.SizeEmbedded(fieldShortHeaderChecksum, x.PayloadHash) +
			proto.SizeEmbedded(fieldShortHeaderHomomorphic, x.HomomorphicHash)
	}
	return sz
}

// MarshalStable writes the ShortHeader in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ShortHeader.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ShortHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldShortHeaderVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], fieldShortHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalToEmbedded(b[off:], fieldShortHeaderOwner, x.OwnerId)
		off += proto.MarshalToVarint(b[off:], fieldShortHeaderType, int32(x.ObjectType))
		off += proto.MarshalToVarint(b[off:], fieldShortHeaderLen, x.PayloadLength)
		off += proto.MarshalToEmbedded(b[off:], fieldShortHeaderChecksum, x.PayloadHash)
		proto.MarshalToEmbedded(b[off:], fieldShortHeaderHomomorphic, x.HomomorphicHash)
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

// MarshaledSize returns size of the Header in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Header) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldHeaderVersion, x.Version) +
			proto.SizeEmbedded(fieldHeaderContainer, x.ContainerId) +
			proto.SizeEmbedded(fieldHeaderOwner, x.OwnerId) +
			proto.SizeVarint(fieldHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeVarint(fieldHeaderLen, x.PayloadLength) +
			proto.SizeEmbedded(fieldHeaderChecksum, x.PayloadHash) +
			proto.SizeVarint(fieldHeaderType, int32(x.ObjectType)) +
			proto.SizeEmbedded(fieldHeaderHomomorphic, x.HomomorphicHash) +
			proto.SizeEmbedded(fieldHeaderSession, x.SessionToken) +
			proto.SizeEmbedded(fieldHeaderSplit, x.Split) +
			proto.SizeRepeatedMessages(fieldHeaderAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the Header in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Header.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldHeaderVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], fieldHeaderContainer, x.ContainerId)
		off += proto.MarshalToEmbedded(b[off:], fieldHeaderOwner, x.OwnerId)
		off += proto.MarshalToVarint(b[off:], fieldHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalToVarint(b[off:], fieldHeaderLen, x.PayloadLength)
		off += proto.MarshalToEmbedded(b[off:], fieldHeaderChecksum, x.PayloadHash)
		off += proto.MarshalToVarint(b[off:], fieldHeaderType, int32(x.ObjectType))
		off += proto.MarshalToEmbedded(b[off:], fieldHeaderHomomorphic, x.HomomorphicHash)
		off += proto.MarshalToEmbedded(b[off:], fieldHeaderSession, x.SessionToken)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldHeaderAttributes, x.Attributes)
		proto.MarshalToEmbedded(b[off:], fieldHeaderSplit, x.Split)
	}
}

const (
	_ = iota
	fieldObjectID
	fieldObjectSignature
	fieldObjectHeader
	fieldObjectPayload
)

// MarshaledSize returns size of the Object in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Object) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldObjectID, x.ObjectId) +
			proto.SizeEmbedded(fieldObjectSignature, x.Signature) +
			proto.SizeEmbedded(fieldObjectHeader, x.Header) +
			proto.SizeBytes(fieldObjectPayload, x.Payload)
	}
	return sz
}

// MarshalStable writes the Object in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Object.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Object) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldObjectID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], fieldObjectSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], fieldObjectHeader, x.Header)
		proto.MarshalToBytes(b[off:], fieldObjectPayload, x.Payload)
	}
}

const (
	_ = iota
	fieldSplitInfoID
	fieldSplitInfoLast
	fieldSplitInfoLink
	fieldSplitInfoFirst
)

// MarshaledSize returns size of the SplitInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SplitInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSplitInfoID, x.SplitId) +
			proto.SizeEmbedded(fieldSplitInfoLast, x.LastPart) +
			proto.SizeEmbedded(fieldSplitInfoLink, x.Link) +
			proto.SizeEmbedded(fieldSplitInfoFirst, x.FirstPart)
	}
	return sz
}

// MarshalStable writes the SplitInfo in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [SplitInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SplitInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldSplitInfoID, x.SplitId)
		off += proto.MarshalToEmbedded(b[off:], fieldSplitInfoLast, x.LastPart)
		off += proto.MarshalToEmbedded(b[off:], fieldSplitInfoLink, x.Link)
		proto.MarshalToEmbedded(b[off:], fieldSplitInfoFirst, x.FirstPart)
	}
}

const (
	_ = iota
	fieldGetReqAddress
	fieldGetReqRaw
)

// MarshaledSize returns size of the GetRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetReqAddress, x.Address) +
			proto.SizeBool(fieldGetReqRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGetReqAddress, x.Address)
		proto.MarshalToBool(b[off:], fieldGetReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldGetRespInitID
	fieldGetRespInitSignature
	fieldGetRespInitHeader
)

// MarshaledSize returns size of the GetResponse_Body_Init in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetRespInitID, x.ObjectId) +
			proto.SizeEmbedded(fieldGetRespInitSignature, x.Signature) +
			proto.SizeEmbedded(fieldGetRespInitHeader, x.Header)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body_Init in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body_Init.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *GetResponse_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGetRespInitID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], fieldGetRespInitSignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], fieldGetRespInitHeader, x.Header)
	}
}

const (
	_ = iota
	fieldGetRespInit
	fieldGetRespChunk
	fieldGetRespSplitInfo
)

// MarshaledSize returns size of the GetResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetRespInit, x.Init) +
			proto.SizeBytes(fieldGetRespChunk, x.Chunk) +
			proto.SizeEmbedded(fieldGetRespSplitInfo, x.SplitInfo)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGetRespInit, x.Init)
		off += proto.MarshalToBytes(b[off:], fieldGetRespChunk, x.Chunk)
		proto.MarshalToEmbedded(b[off:], fieldGetRespSplitInfo, x.SplitInfo)
	}
}

const (
	_ = iota
	fieldHeadReqAddress
	fieldHeadReqMain
	fieldHeadReqRaw
)

// MarshaledSize returns size of the HeadRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeadRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldHeadReqAddress, x.Address) +
			proto.SizeBool(fieldHeadReqMain, x.MainOnly) +
			proto.SizeBool(fieldHeadReqRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the HeadRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [HeadRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *HeadRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldHeadReqAddress, x.Address)
		off += proto.MarshalToBool(b[off:], fieldHeadReqMain, x.MainOnly)
		proto.MarshalToBool(b[off:], fieldHeadReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldHeaderSignatureHdr
	fieldHeaderSignatureSig
)

// MarshaledSize returns size of the HeaderWithSignature in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeaderWithSignature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldHeaderSignatureHdr, x.Header) +
			proto.SizeEmbedded(fieldHeaderSignatureSig, x.Signature)
	}
	return sz
}

// MarshalStable writes the HeaderWithSignature in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [HeaderWithSignature.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *HeaderWithSignature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldHeaderSignatureHdr, x.Header)
		proto.MarshalToEmbedded(b[off:], fieldHeaderSignatureSig, x.Signature)
	}
}

const (
	_ = iota
	fieldHeadRespHeader
	fieldHeadRespShort
	fieldHeadRespSplitInfo
)

// MarshaledSize returns size of the HeadResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeadResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch h := x.Head.(type) {
		default:
			panic(fmt.Sprintf("unexpected head part %T", x.Head))
		case nil:
		case *HeadResponse_Body_Header:
			if h != nil {
				sz = proto.SizeEmbedded(fieldHeadRespHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				sz = proto.SizeEmbedded(fieldHeadRespShort, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				sz = proto.SizeEmbedded(fieldHeadRespSplitInfo, h.SplitInfo)
			}
		}
	}
	return sz
}

// MarshalStable writes the HeadResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [HeadResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *HeadResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch h := x.Head.(type) {
		default:
			panic(fmt.Sprintf("unexpected head part %T", x.Head))
		case nil:
		case *HeadResponse_Body_Header:
			if h != nil {
				proto.MarshalToEmbedded(b, fieldHeadRespHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				proto.MarshalToEmbedded(b, fieldHeadRespShort, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				proto.MarshalToEmbedded(b, fieldHeadRespSplitInfo, h.SplitInfo)
			}
		}
	}
}

const (
	_ = iota
	fieldRangeOff
	fieldRangeLen
)

// MarshaledSize returns size of the Range in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Range) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldRangeOff, x.Offset) +
			proto.SizeVarint(fieldRangeLen, x.Length)
	}
	return sz
}

// MarshalStable writes the Range in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Range.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Range) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldRangeOff, x.Offset)
		proto.MarshalToVarint(b[off:], fieldRangeLen, x.Length)
	}
}

const (
	_ = iota
	fieldRangeReqAddress
	fieldRangeReqRange
	fieldRangeReqRaw
)

// MarshaledSize returns size of the GetRangeRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldRangeReqAddress, x.Address) +
			proto.SizeEmbedded(fieldRangeReqRange, x.Range) +
			proto.SizeBool(fieldRangeReqRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the GetRangeRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [GetRangeRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *GetRangeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRangeReqAddress, x.Address)
		off += proto.MarshalToEmbedded(b[off:], fieldRangeReqRange, x.Range)
		proto.MarshalToBool(b[off:], fieldRangeReqRaw, x.Raw)
	}
}

const (
	_ = iota
	fieldRangeRespChunk
	fieldRangeRespSplitInfo
)

// MarshaledSize returns size of the GetRangeResponse_Body in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
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
				sz = proto.SizeEmbedded(fieldRangeRespSplitInfo, p.SplitInfo)
			}
		}
	}
	return sz
}

// MarshalStable writes the GetRangeResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [GetRangeResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *GetRangeResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.RangePart.(type) {
		default:
			panic(fmt.Sprintf("unexpected range part %T", x.RangePart))
		case nil:
		case *GetRangeResponse_Body_Chunk:
			if p != nil {
				proto.MarshalToBytes(b, fieldRangeRespChunk, p.Chunk)
			}
		case *GetRangeResponse_Body_SplitInfo:
			if p != nil {
				proto.MarshalToEmbedded(b, fieldRangeRespSplitInfo, p.SplitInfo)
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

// MarshaledSize returns size of the GetRangeHashRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeHashRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldRangeHashReqAddress, x.Address) +
			proto.SizeBytes(fieldRangeHashReqSalt, x.Salt) +
			proto.SizeVarint(fieldRangeHashReqType, int32(x.Type)) +
			proto.SizeRepeatedMessages(fieldRangeHashReqRanges, x.Ranges)
	}
	return 0
}

// MarshalStable writes the GetRangeHashRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetRangeHashRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetRangeHashRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRangeHashReqAddress, x.Address)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldRangeHashReqRanges, x.Ranges)
		off += proto.MarshalToBytes(b[off:], fieldRangeHashReqSalt, x.Salt)
		proto.MarshalToVarint(b[off:], fieldRangeHashReqType, int32(x.Type))
	}
}

const (
	_ = iota
	fieldRangeHashRespType
	fieldRangeHashRespHashes
)

// MarshaledSize returns size of the GetRangeHashResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeHashResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldRangeHashRespType, int32(x.Type)) +
			proto.SizeRepeatedBytes(fieldRangeHashRespHashes, x.HashList)
	}
	return sz
}

// MarshalStable writes the GetRangeHashResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetRangeHashResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetRangeHashResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldRangeHashRespType, int32(x.Type))
		proto.MarshalToRepeatedBytes(b[off:], fieldRangeHashRespHashes, x.HashList)
	}
}

const (
	_ = iota
	fieldPutReqInitID
	fieldPutReqInitSignature
	fieldPutReqInitHeader
	fieldPutReqInitCopies
)

// MarshaledSize returns size of the PutRequest_Body_Init in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldPutReqInitID, x.ObjectId) +
			proto.SizeEmbedded(fieldPutReqInitSignature, x.Signature) +
			proto.SizeEmbedded(fieldPutReqInitHeader, x.Header) +
			proto.SizeVarint(fieldPutReqInitCopies, x.CopiesNumber)
	}
	return sz
}

// MarshalStable writes the PutRequest_Body_Init in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body_Init.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *PutRequest_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldPutReqInitID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], fieldPutReqInitSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], fieldPutReqInitHeader, x.Header)
		proto.MarshalToVarint(b[off:], fieldPutReqInitCopies, x.CopiesNumber)
	}
}

const (
	_ = iota
	fieldPutReqInit
	fieldPutReqChunk
)

// MarshaledSize returns size of the PutRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *PutRequest_Body_Init_:
			sz = proto.SizeEmbedded(fieldPutReqInit, p.Init)
		case *PutRequest_Body_Chunk:
			sz = proto.SizeBytes(fieldPutReqChunk, p.Chunk)
		}
	}
	return sz
}

// MarshalStable writes the PutRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *PutRequest_Body_Init_:
			proto.MarshalToEmbedded(b, fieldPutReqInit, p.Init)
		case *PutRequest_Body_Chunk:
			proto.MarshalToBytes(b, fieldPutReqChunk, p.Chunk)
		}
	}
}

const (
	_ = iota
	fieldPutRespID
)

// MarshaledSize returns size of the PutResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldPutRespID, x.ObjectId)
	}
	return sz
}

// MarshalStable writes the PutResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldPutRespID, x.ObjectId)
	}
}

const (
	_ = iota
	fieldDeleteReqAddress
)

// MarshaledSize returns size of the DeleteRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldDeleteReqAddress, x.Address)
	}
	return sz
}

// MarshalStable writes the DeleteRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldDeleteReqAddress, x.Address)
	}
}

const (
	_ = iota
	fieldDeleteRespTombstone
)

// MarshaledSize returns size of the DeleteResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldDeleteRespTombstone, x.Tombstone)
	}
	return sz
}

// MarshalStable writes the DeleteResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldDeleteRespTombstone, x.Tombstone)
	}
}

const (
	_ = iota
	fieldSearchFilterMatcher
	fieldSearchFilterKey
	fieldSearchFilterValue
)

// MarshaledSize returns size of the SearchFilter in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SearchFilter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldSearchFilterMatcher, int32(x.MatchType))
		sz += proto.SizeBytes(fieldSearchFilterKey, x.Key)
		sz += proto.SizeBytes(fieldSearchFilterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the SearchFilter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SearchFilter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SearchFilter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldSearchFilterMatcher, int32(x.MatchType))
		off += proto.MarshalToBytes(b[off:], fieldSearchFilterKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldSearchFilterValue, x.Value)
	}
}

const (
	_ = iota
	fieldSearchReqContainer
	fieldSearchReqVersion
	fieldSearchReqFilters
)

// MarshaledSize returns size of the SearchRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SearchRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSearchReqContainer, x.ContainerId) +
			proto.SizeVarint(fieldSearchReqVersion, x.Version) +
			proto.SizeRepeatedMessages(fieldSearchReqFilters, x.Filters)
	}
	return 0
}

// MarshalStable writes the SearchRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SearchRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSearchReqContainer, x.ContainerId)
		off += proto.MarshalToVarint(b[off:], fieldSearchReqVersion, x.Version)
		proto.MarshalToRepeatedMessages(b[off:], fieldSearchReqFilters, x.Filters)
	}
}

const (
	_ = iota
	fieldSearchRespIDList
)

// MarshaledSize returns size of the SearchResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SearchResponse_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(fieldSearchRespIDList, x.IdList)
	}
	return 0
}

// MarshalStable writes the SearchResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SearchResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, fieldSearchRespIDList, x.IdList)
	}
}

const (
	_ = iota
	fieldSearchV2ReqContainer
	fieldSearchV2ReqVersion
	fieldSearchV2ReqFilters
	fieldSearchV2ReqCursor
	fieldSearchV2ReqCount
	fieldSearchV2ReqAttrs
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Request_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSearchV2ReqContainer, x.ContainerId) +
			proto.SizeVarint(fieldSearchV2ReqVersion, x.Version) +
			proto.SizeRepeatedMessages(fieldSearchV2ReqFilters, x.Filters) +
			proto.SizeBytes(fieldSearchV2ReqCursor, x.Cursor) +
			proto.SizeVarint(fieldSearchV2ReqCount, x.Count) +
			proto.SizeRepeatedBytes(fieldSearchV2ReqAttrs, x.Attributes)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Request_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchV2Request_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSearchV2ReqContainer, x.ContainerId)
		off += proto.MarshalToVarint(b[off:], fieldSearchV2ReqVersion, x.Version)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldSearchV2ReqFilters, x.Filters)
		off += proto.MarshalToBytes(b[off:], fieldSearchV2ReqCursor, x.Cursor)
		off += proto.MarshalToVarint(b[off:], fieldSearchV2ReqCount, x.Count)
		proto.MarshalToRepeatedBytes(b[off:], fieldSearchV2ReqAttrs, x.Attributes)
	}
}

const (
	_ = iota
	fieldSearchV2RespItemID
	fieldSearchV2RespItemAttrs
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Response_OIDWithMeta) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSearchV2RespItemID, x.Id) +
			proto.SizeRepeatedBytes(fieldSearchV2RespItemAttrs, x.Attributes)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Response_OIDWithMeta.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *SearchV2Response_OIDWithMeta) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSearchV2RespItemID, x.Id)
		proto.MarshalToRepeatedBytes(b[off:], fieldSearchV2RespItemAttrs, x.Attributes)
	}
}

const (
	_ = iota
	fieldSearchV2RespItems
	fieldSearchV2RespCursor
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Response_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(fieldSearchV2RespItems, x.Result) +
			proto.SizeBytes(fieldSearchV2RespCursor, x.Cursor)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Response_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchV2Response_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToRepeatedMessages(b, fieldSearchV2RespItems, x.Result)
		proto.MarshalToBytes(b[off:], fieldSearchV2RespCursor, x.Cursor)
	}
}
