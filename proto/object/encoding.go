package object

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [Header_Split] message.
const (
	_ = iota
	FieldHeaderSplitParent
	FieldHeaderSplitPrevious
	FieldHeaderSplitParentSignature
	FieldHeaderSplitParentHeader
	FieldHeaderSplitChildren
	FieldHeaderSplitSplitID
	FieldHeaderSplitFirst
)

// MarshaledSize returns size of the Header_Split in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *Header_Split) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldHeaderSplitParent, x.Parent) +
			proto.SizeEmbedded(FieldHeaderSplitPrevious, x.Previous) +
			proto.SizeEmbedded(FieldHeaderSplitParentSignature, x.ParentSignature) +
			proto.SizeEmbedded(FieldHeaderSplitParentHeader, x.ParentHeader) +
			proto.SizeBytes(FieldHeaderSplitSplitID, x.SplitId) +
			proto.SizeEmbedded(FieldHeaderSplitFirst, x.First) +
			proto.SizeRepeatedMessages(FieldHeaderSplitChildren, x.Children)
	}
	return 0
}

// MarshalStable writes the Header_Split in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Header_Split.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header_Split) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldHeaderSplitParent, x.Parent)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderSplitPrevious, x.Previous)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderSplitParentSignature, x.ParentSignature)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderSplitParentHeader, x.ParentHeader)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldHeaderSplitChildren, x.Children)
		off += proto.MarshalToBytes(b[off:], FieldHeaderSplitSplitID, x.SplitId)
		proto.MarshalToEmbedded(b[off:], FieldHeaderSplitFirst, x.First)
	}
}

// Field numbers of [Header_Attribute] message.
const (
	_ = iota
	FieldAttributeKey
	FieldAttributeValue
)

// MarshaledSize returns size of the Header_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Header_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldAttributeKey, x.Key) +
			proto.SizeBytes(FieldAttributeValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Header_Attribute in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Header_Attribute.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldAttributeKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldAttributeValue, x.Value)
	}
}

// Field numbers of [ShortHeader] message.
const (
	_ = iota
	FieldShortHeaderVersion
	FieldShortHeaderCreationEpoch
	FieldShortHeaderOwnerID
	FieldShortHeaderObjectType
	FieldShortHeaderPayloadLength
	FieldShortHeaderPayloadHash
	FieldShortHeaderHomomorphicHash
)

// MarshaledSize returns size of the ShortHeader in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *ShortHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldShortHeaderVersion, x.Version) +
			proto.SizeVarint(FieldShortHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeEmbedded(FieldShortHeaderOwnerID, x.OwnerId) +
			proto.SizeVarint(FieldShortHeaderObjectType, int32(x.ObjectType)) +
			proto.SizeVarint(FieldShortHeaderPayloadLength, x.PayloadLength) +
			proto.SizeEmbedded(FieldShortHeaderPayloadHash, x.PayloadHash) +
			proto.SizeEmbedded(FieldShortHeaderHomomorphicHash, x.HomomorphicHash)
	}
	return sz
}

// MarshalStable writes the ShortHeader in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ShortHeader.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ShortHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldShortHeaderVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], FieldShortHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalToEmbedded(b[off:], FieldShortHeaderOwnerID, x.OwnerId)
		off += proto.MarshalToVarint(b[off:], FieldShortHeaderObjectType, int32(x.ObjectType))
		off += proto.MarshalToVarint(b[off:], FieldShortHeaderPayloadLength, x.PayloadLength)
		off += proto.MarshalToEmbedded(b[off:], FieldShortHeaderPayloadHash, x.PayloadHash)
		proto.MarshalToEmbedded(b[off:], FieldShortHeaderHomomorphicHash, x.HomomorphicHash)
	}
}

// Field numbers of [Header] message.
const (
	_ = iota
	FieldHeaderVersion
	FieldHeaderContainerID
	FieldHeaderOwnerID
	FieldHeaderCreationEpoch
	FieldHeaderPayloadLength
	FieldHeaderPayloadHash
	FieldHeaderObjectType
	FieldHeaderHomomorphicHash
	FieldHeaderSessionToken
	FieldHeaderAttributes
	FieldHeaderSplit
	FieldHeaderSessionV2
)

// MarshaledSize returns size of the Header in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Header) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldHeaderVersion, x.Version) +
			proto.SizeEmbedded(FieldHeaderContainerID, x.ContainerId) +
			proto.SizeEmbedded(FieldHeaderOwnerID, x.OwnerId) +
			proto.SizeVarint(FieldHeaderCreationEpoch, x.CreationEpoch) +
			proto.SizeVarint(FieldHeaderPayloadLength, x.PayloadLength) +
			proto.SizeEmbedded(FieldHeaderPayloadHash, x.PayloadHash) +
			proto.SizeVarint(FieldHeaderObjectType, int32(x.ObjectType)) +
			proto.SizeEmbedded(FieldHeaderHomomorphicHash, x.HomomorphicHash) +
			proto.SizeEmbedded(FieldHeaderSessionToken, x.SessionToken) +
			proto.SizeEmbedded(FieldHeaderSplit, x.Split) +
			proto.SizeRepeatedMessages(FieldHeaderAttributes, x.Attributes) +
			proto.SizeEmbedded(FieldHeaderSessionV2, x.SessionTokenV2)
	}
	return 0
}

// MarshalStable writes the Header in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Header.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Header) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldHeaderVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderContainerID, x.ContainerId)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderOwnerID, x.OwnerId)
		off += proto.MarshalToVarint(b[off:], FieldHeaderCreationEpoch, x.CreationEpoch)
		off += proto.MarshalToVarint(b[off:], FieldHeaderPayloadLength, x.PayloadLength)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderPayloadHash, x.PayloadHash)
		off += proto.MarshalToVarint(b[off:], FieldHeaderObjectType, int32(x.ObjectType))
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderHomomorphicHash, x.HomomorphicHash)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderSessionToken, x.SessionToken)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldHeaderAttributes, x.Attributes)
		off += proto.MarshalToEmbedded(b[off:], FieldHeaderSplit, x.Split)
		proto.MarshalToEmbedded(b[off:], FieldHeaderSessionV2, x.SessionTokenV2)
	}
}

// Field numbers of [Object] message.
const (
	_ = iota
	FieldObjectID
	FieldObjectSignature
	FieldObjectHeader
	FieldObjectPayload
)

// MarshaledSize returns size of the Object in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Object) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldObjectID, x.ObjectId) +
			proto.SizeEmbedded(FieldObjectSignature, x.Signature) +
			proto.SizeEmbedded(FieldObjectHeader, x.Header) +
			proto.SizeBytes(FieldObjectPayload, x.Payload)
	}
	return sz
}

// MarshalStable writes the Object in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Object.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Object) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldObjectID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], FieldObjectSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], FieldObjectHeader, x.Header)
		proto.MarshalToBytes(b[off:], FieldObjectPayload, x.Payload)
	}
}

// Field numbers of [SplitInfo] message.
const (
	_ = iota
	FieldSplitInfoSplitID
	FieldSplitInfoLastPart
	FieldSplitInfoLink
	FieldSplitInfoFirstPart
)

// MarshaledSize returns size of the SplitInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SplitInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldSplitInfoSplitID, x.SplitId) +
			proto.SizeEmbedded(FieldSplitInfoLastPart, x.LastPart) +
			proto.SizeEmbedded(FieldSplitInfoLink, x.Link) +
			proto.SizeEmbedded(FieldSplitInfoFirstPart, x.FirstPart)
	}
	return sz
}

// MarshalStable writes the SplitInfo in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [SplitInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SplitInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldSplitInfoSplitID, x.SplitId)
		off += proto.MarshalToEmbedded(b[off:], FieldSplitInfoLastPart, x.LastPart)
		off += proto.MarshalToEmbedded(b[off:], FieldSplitInfoLink, x.Link)
		proto.MarshalToEmbedded(b[off:], FieldSplitInfoFirstPart, x.FirstPart)
	}
}

// Field numbers of [GetRequest_Body] message.
const (
	_ = iota
	FieldGetRequestBodyAddress
	FieldGetRequestBodyRaw
)

// MarshaledSize returns size of the GetRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetRequestBodyAddress, x.Address) +
			proto.SizeBool(FieldGetRequestBodyRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGetRequestBodyAddress, x.Address)
		proto.MarshalToBool(b[off:], FieldGetRequestBodyRaw, x.Raw)
	}
}

// Field numbers of [GetResponse_Body_Init] message.
const (
	_ = iota
	FieldGetResponseBodyInitObjectID
	FieldGetResponseBodyInitSignature
	FieldGetResponseBodyInitHeader
)

// MarshaledSize returns size of the GetResponse_Body_Init in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetResponseBodyInitObjectID, x.ObjectId) +
			proto.SizeEmbedded(FieldGetResponseBodyInitSignature, x.Signature) +
			proto.SizeEmbedded(FieldGetResponseBodyInitHeader, x.Header)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body_Init in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body_Init.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *GetResponse_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGetResponseBodyInitObjectID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], FieldGetResponseBodyInitSignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], FieldGetResponseBodyInitHeader, x.Header)
	}
}

// Field numbers of [GetResponse_Body] message.
const (
	_ = iota
	FieldGetResponseBodyInit
	FieldGetResponseBodyChunk
	FieldGetResponseBodySplitInfo
)

// MarshaledSize returns size of the GetResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *GetResponse_Body_Init_:
			if p != nil {
				sz = proto.SizeEmbedded(FieldGetResponseBodyInit, p.Init)
			}
		case *GetResponse_Body_Chunk:
			if p != nil {
				sz = proto.SizeBytes(FieldGetResponseBodyChunk, p.Chunk)
			}
		case *GetResponse_Body_SplitInfo:
			if p != nil {
				sz = proto.SizeEmbedded(FieldGetResponseBodySplitInfo, p.SplitInfo)
			}
		}
	}
	return sz
}

// MarshalStable writes the GetResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		switch p := x.ObjectPart.(type) {
		default:
			panic(fmt.Sprintf("unexpected object part %T", x.ObjectPart))
		case nil:
		case *GetResponse_Body_Init_:
			if p != nil {
				proto.MarshalToEmbedded(b, FieldGetResponseBodyInit, p.Init)
			}
		case *GetResponse_Body_Chunk:
			if p != nil {
				proto.MarshalToBytes(b, FieldGetResponseBodyChunk, p.Chunk)
			}
		case *GetResponse_Body_SplitInfo:
			if p != nil {
				proto.MarshalToEmbedded(b, FieldGetResponseBodySplitInfo, p.SplitInfo)
			}
		}
	}
}

// Field numbers of [HeadRequest_Body] message.
const (
	_ = iota
	FieldHeadRequestBodyAddress
	FieldHeadRequestBodyMainOnly
	FieldHeadRequestBodyRaw
)

// MarshaledSize returns size of the HeadRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeadRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldHeadRequestBodyAddress, x.Address) +
			proto.SizeBool(FieldHeadRequestBodyMainOnly, x.MainOnly) +
			proto.SizeBool(FieldHeadRequestBodyRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the HeadRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [HeadRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *HeadRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldHeadRequestBodyAddress, x.Address)
		off += proto.MarshalToBool(b[off:], FieldHeadRequestBodyMainOnly, x.MainOnly)
		proto.MarshalToBool(b[off:], FieldHeadRequestBodyRaw, x.Raw)
	}
}

// Field numbers of [HeaderWithSignature] message.
const (
	_ = iota
	FieldHeaderWithSignatureHeader
	FieldHeaderWithSignatureSignature
)

// MarshaledSize returns size of the HeaderWithSignature in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeaderWithSignature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldHeaderWithSignatureHeader, x.Header) +
			proto.SizeEmbedded(FieldHeaderWithSignatureSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the HeaderWithSignature in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [HeaderWithSignature.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *HeaderWithSignature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldHeaderWithSignatureHeader, x.Header)
		proto.MarshalToEmbedded(b[off:], FieldHeaderWithSignatureSignature, x.Signature)
	}
}

// Field numbers of [HeadResponse_Body] message.
const (
	_ = iota
	FieldHeadResponseBodyHeader
	FieldHeadResponseBodyShortHeader
	FieldHeadResponseBodySplitInfo
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
				sz = proto.SizeEmbedded(FieldHeadResponseBodyHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				sz = proto.SizeEmbedded(FieldHeadResponseBodyShortHeader, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				sz = proto.SizeEmbedded(FieldHeadResponseBodySplitInfo, h.SplitInfo)
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
				proto.MarshalToEmbedded(b, FieldHeadResponseBodyHeader, h.Header)
			}
		case *HeadResponse_Body_ShortHeader:
			if h != nil {
				proto.MarshalToEmbedded(b, FieldHeadResponseBodyShortHeader, h.ShortHeader)
			}
		case *HeadResponse_Body_SplitInfo:
			if h != nil {
				proto.MarshalToEmbedded(b, FieldHeadResponseBodySplitInfo, h.SplitInfo)
			}
		}
	}
}

// Field numbers of [Range] message.
const (
	_ = iota
	FieldRangeOffset
	FieldRangeLength
)

// MarshaledSize returns size of the Range in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Range) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldRangeOffset, x.Offset) +
			proto.SizeVarint(FieldRangeLength, x.Length)
	}
	return sz
}

// MarshalStable writes the Range in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Range.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Range) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldRangeOffset, x.Offset)
		proto.MarshalToVarint(b[off:], FieldRangeLength, x.Length)
	}
}

// Field numbers of [GetRangeRequest_Body] message.
const (
	_ = iota
	FieldRangeRequestBodyAddress
	FieldRangeRequestBodyRange
	FieldRangeRequestBodyRaw
)

// MarshaledSize returns size of the GetRangeRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldRangeRequestBodyAddress, x.Address) +
			proto.SizeEmbedded(FieldRangeRequestBodyRange, x.Range) +
			proto.SizeBool(FieldRangeRequestBodyRaw, x.Raw)
	}
	return sz
}

// MarshalStable writes the GetRangeRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [GetRangeRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *GetRangeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRangeRequestBodyAddress, x.Address)
		off += proto.MarshalToEmbedded(b[off:], FieldRangeRequestBodyRange, x.Range)
		proto.MarshalToBool(b[off:], FieldRangeRequestBodyRaw, x.Raw)
	}
}

// Field numbers of [GetRangeResponse_Body] message.
const (
	_ = iota
	FieldRangeResponseBodyChunk
	FieldRangeResponseBodySplitInfo
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
				sz = proto.SizeBytes(FieldRangeResponseBodyChunk, p.Chunk)
			}
		case *GetRangeResponse_Body_SplitInfo:
			if p != nil {
				sz = proto.SizeEmbedded(FieldRangeResponseBodySplitInfo, p.SplitInfo)
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
				proto.MarshalToBytes(b, FieldRangeResponseBodyChunk, p.Chunk)
			}
		case *GetRangeResponse_Body_SplitInfo:
			if p != nil {
				proto.MarshalToEmbedded(b, FieldRangeResponseBodySplitInfo, p.SplitInfo)
			}
		}
	}
}

// Field numbers of [GetRangeHashRequest_Body] message.
const (
	_ = iota
	FieldRangeHashRequestBodyAddress
	FieldRangeHashRequestBodyRanges
	FieldRangeHashRequestBodySalt
	FieldRangeHashRequestBodyType
)

// MarshaledSize returns size of the GetRangeHashRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeHashRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldRangeHashRequestBodyAddress, x.Address) +
			proto.SizeBytes(FieldRangeHashRequestBodySalt, x.Salt) +
			proto.SizeVarint(FieldRangeHashRequestBodyType, int32(x.Type)) +
			proto.SizeRepeatedMessages(FieldRangeHashRequestBodyRanges, x.Ranges)
	}
	return 0
}

// MarshalStable writes the GetRangeHashRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetRangeHashRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetRangeHashRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRangeHashRequestBodyAddress, x.Address)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldRangeHashRequestBodyRanges, x.Ranges)
		off += proto.MarshalToBytes(b[off:], FieldRangeHashRequestBodySalt, x.Salt)
		proto.MarshalToVarint(b[off:], FieldRangeHashRequestBodyType, int32(x.Type))
	}
}

// Field numbers of [GetRangeHashResponse_Body] message.
const (
	_ = iota
	FieldRangeHashResponseBodyType
	FieldRangeHashResponseBodyHashes
)

// MarshaledSize returns size of the GetRangeHashResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetRangeHashResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldRangeHashResponseBodyType, int32(x.Type)) +
			proto.SizeRepeatedBytes(FieldRangeHashResponseBodyHashes, x.HashList)
	}
	return sz
}

// MarshalStable writes the GetRangeHashResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetRangeHashResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetRangeHashResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldRangeHashResponseBodyType, int32(x.Type))
		proto.MarshalToRepeatedBytes(b[off:], FieldRangeHashResponseBodyHashes, x.HashList)
	}
}

// Field numbers of [PutRequest_Body_Init] message.
const (
	_ = iota
	FieldPutRequestBodyInitID
	FieldPutRequestBodyInitSignature
	FieldPutRequestBodyInitHeader
	FieldPutRequestBodyInitCopiesNumber
)

// MarshaledSize returns size of the PutRequest_Body_Init in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body_Init) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldPutRequestBodyInitID, x.ObjectId) +
			proto.SizeEmbedded(FieldPutRequestBodyInitSignature, x.Signature) +
			proto.SizeEmbedded(FieldPutRequestBodyInitHeader, x.Header) +
			proto.SizeVarint(FieldPutRequestBodyInitCopiesNumber, x.CopiesNumber)
	}
	return sz
}

// MarshalStable writes the PutRequest_Body_Init in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body_Init.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *PutRequest_Body_Init) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldPutRequestBodyInitID, x.ObjectId)
		off += proto.MarshalToEmbedded(b[off:], FieldPutRequestBodyInitSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], FieldPutRequestBodyInitHeader, x.Header)
		proto.MarshalToVarint(b[off:], FieldPutRequestBodyInitCopiesNumber, x.CopiesNumber)
	}
}

// Field numbers of [PutRequest_Body] message.
const (
	_ = iota
	FieldPutRequestBodyInit
	FieldPutRequestBodyChunk
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
			sz = proto.SizeEmbedded(FieldPutRequestBodyInit, p.Init)
		case *PutRequest_Body_Chunk:
			sz = proto.SizeBytes(FieldPutRequestBodyChunk, p.Chunk)
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
			proto.MarshalToEmbedded(b, FieldPutRequestBodyInit, p.Init)
		case *PutRequest_Body_Chunk:
			proto.MarshalToBytes(b, FieldPutRequestBodyChunk, p.Chunk)
		}
	}
}

// Field numbers of [PutResponse_Body] message.
const (
	_ = iota
	FieldPutResponseBodyObjectID
)

// MarshaledSize returns size of the PutResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldPutResponseBodyObjectID, x.ObjectId)
	}
	return sz
}

// MarshalStable writes the PutResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldPutResponseBodyObjectID, x.ObjectId)
	}
}

// Field numbers of [DeleteRequest_Body] message.
const (
	_ = iota
	FieldDeleteRequestBodyAddress
)

// MarshaledSize returns size of the DeleteRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldDeleteRequestBodyAddress, x.Address)
	}
	return sz
}

// MarshalStable writes the DeleteRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldDeleteRequestBodyAddress, x.Address)
	}
}

// Field numbers of [DeleteResponse_Body] message.
const (
	_ = iota
	FieldDeleteResponseBodyTombstone
)

// MarshaledSize returns size of the DeleteResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldDeleteResponseBodyTombstone, x.Tombstone)
	}
	return sz
}

// MarshalStable writes the DeleteResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldDeleteResponseBodyTombstone, x.Tombstone)
	}
}

// Field numbers of [SearchFilter] message.
const (
	_ = iota
	FieldSearchFilterMatcher
	FieldSearchFilterKey
	FieldSearchFilterValue
)

// MarshaledSize returns size of the SearchFilter in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SearchFilter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldSearchFilterMatcher, int32(x.MatchType))
		sz += proto.SizeBytes(FieldSearchFilterKey, x.Key)
		sz += proto.SizeBytes(FieldSearchFilterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the SearchFilter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SearchFilter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SearchFilter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldSearchFilterMatcher, int32(x.MatchType))
		off += proto.MarshalToBytes(b[off:], FieldSearchFilterKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldSearchFilterValue, x.Value)
	}
}

// Field numbers of [SearchRequest_Body] message.
const (
	_ = iota
	FieldSearchRequestBodyContainerID
	FieldSearchRequestBodyVersion
	FieldSearchRequestBodyFilters
)

// MarshaledSize returns size of the SearchRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SearchRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldSearchRequestBodyContainerID, x.ContainerId) +
			proto.SizeVarint(FieldSearchRequestBodyVersion, x.Version) +
			proto.SizeRepeatedMessages(FieldSearchRequestBodyFilters, x.Filters)
	}
	return 0
}

// MarshalStable writes the SearchRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SearchRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSearchRequestBodyContainerID, x.ContainerId)
		off += proto.MarshalToVarint(b[off:], FieldSearchRequestBodyVersion, x.Version)
		proto.MarshalToRepeatedMessages(b[off:], FieldSearchRequestBodyFilters, x.Filters)
	}
}

// Field numbers of [SearchResponse_Body] message.
const (
	_ = iota
	FieldSearchResponseBodyIDList
)

// MarshaledSize returns size of the SearchResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SearchResponse_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(FieldSearchResponseBodyIDList, x.IdList)
	}
	return 0
}

// MarshalStable writes the SearchResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SearchResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, FieldSearchResponseBodyIDList, x.IdList)
	}
}

// Field numbers of [SearchV2Request_Body] message.
const (
	_ = iota
	FieldSearchV2RequestBodyContainerID
	FieldSearchV2RequestBodyVersion
	FieldSearchV2RequestBodyFilters
	FieldSearchV2RequestBodyCursor
	FieldSearchV2RequestBodyCount
	FieldSearchV2RequestBodyAttributes
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Request_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldSearchV2RequestBodyContainerID, x.ContainerId) +
			proto.SizeVarint(FieldSearchV2RequestBodyVersion, x.Version) +
			proto.SizeRepeatedMessages(FieldSearchV2RequestBodyFilters, x.Filters) +
			proto.SizeBytes(FieldSearchV2RequestBodyCursor, x.Cursor) +
			proto.SizeVarint(FieldSearchV2RequestBodyCount, x.Count) +
			proto.SizeRepeatedBytes(FieldSearchV2RequestBodyAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Request_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchV2Request_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSearchV2RequestBodyContainerID, x.ContainerId)
		off += proto.MarshalToVarint(b[off:], FieldSearchV2RequestBodyVersion, x.Version)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldSearchV2RequestBodyFilters, x.Filters)
		off += proto.MarshalToBytes(b[off:], FieldSearchV2RequestBodyCursor, x.Cursor)
		off += proto.MarshalToVarint(b[off:], FieldSearchV2RequestBodyCount, x.Count)
		proto.MarshalToRepeatedBytes(b[off:], FieldSearchV2RequestBodyAttributes, x.Attributes)
	}
}

// Field numbers of [SearchV2Response_OIDWithMeta] message.
const (
	_ = iota
	FieldSearchV2ResponseBodyOIDWithMetaID
	FieldSearchV2ResponseBodyOIDWithMetaAttributes
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Response_OIDWithMeta) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldSearchV2ResponseBodyOIDWithMetaID, x.Id) +
			proto.SizeRepeatedBytes(FieldSearchV2ResponseBodyOIDWithMetaAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Response_OIDWithMeta.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *SearchV2Response_OIDWithMeta) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSearchV2ResponseBodyOIDWithMetaID, x.Id)
		proto.MarshalToRepeatedBytes(b[off:], FieldSearchV2ResponseBodyOIDWithMetaAttributes, x.Attributes)
	}
}

// Field numbers of [SearchV2Response_Body] message.
const (
	_ = iota
	FieldSearchV2ResponseBodyResult
	FieldSearchV2ResponseBodyCursor
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Response_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(FieldSearchV2ResponseBodyResult, x.Result) +
			proto.SizeBytes(FieldSearchV2ResponseBodyCursor, x.Cursor)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Response_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchV2Response_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToRepeatedMessages(b, FieldSearchV2ResponseBodyResult, x.Result)
		proto.MarshalToBytes(b[off:], FieldSearchV2ResponseBodyCursor, x.Cursor)
	}
}

// Field numbers of [ReplicateRequest] message.
const (
	_ = iota
	FieldReplicateRequestObject
	FieldReplicateRequestSignature
	FieldReplicateRequestSignObject
)

// Field numbers of [ReplicateResponse] message.
const (
	_ = iota
	FieldReplicateResponseStatus
	FieldReplicateResponseObjectSignature
)
