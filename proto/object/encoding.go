package object

import (
	"crypto/sha256"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"google.golang.org/protobuf/encoding/protowire"
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
			proto.SizeVarint(FieldShortHeaderObjectType, x.ObjectType) +
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
		off += proto.MarshalToVarint(b[off:], FieldShortHeaderObjectType, x.ObjectType)
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
			proto.SizeVarint(FieldHeaderObjectType, x.ObjectType) +
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
		off += proto.MarshalToVarint(b[off:], FieldHeaderObjectType, x.ObjectType)
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
	FieldGetRequestBodyRange
	FieldGetRequestBodyPayloadOnly
	FieldGetRequestBodyExtendedRange
)

// CalculateGetRequestBodyLength calculates length of Get request body message
// with static address and given dynamic fields.
func CalculateGetRequestBodyLength(raw bool, rngOff uint64, rngLen uint64, payloadOnly bool, extRngFirst *uint64, extRngLast *uint64) int {
	ln := proto.SizeEmbeddedLENField(FieldGetRequestBodyAddress, refs.ObjectAddressLength)
	ln += calculateDynamicGetRequestBodyFieldsLength(raw, rngOff, rngLen, payloadOnly, extRngFirst, extRngLast)
	return ln
}

func calculateDynamicGetRequestBodyFieldsLength(raw bool, rngOff uint64, rngLen uint64, payloadOnly bool, extRngFirst *uint64, extRngLast *uint64) int {
	ln := proto.SizeBool(FieldGetRequestBodyRaw, raw)
	ln += proto.SizeEmbeddedLENField(FieldGetRequestBodyRange, CalculateRangeLength(rngOff, rngLen))
	ln += proto.SizeBool(FieldGetRequestBodyPayloadOnly, payloadOnly)
	ln += proto.SizeEmbeddedLENField(FieldGetRequestBodyExtendedRange, CalculateExtendedRangeLength(extRngFirst, extRngLast))
	return ln
}

// MarshaledSize returns size of the GetRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		var firstPos, lastPos *uint64
		if x.ExtendedRange != nil {
			firstPos, lastPos = x.ExtendedRange.FirstPos, x.ExtendedRange.LastPos
		}
		sz = proto.SizeEmbedded(FieldGetRequestBodyAddress, x.Address) +
			calculateDynamicGetRequestBodyFieldsLength(x.Raw, x.Range.GetOffset(), x.Range.GetLength(), x.PayloadOnly, firstPos, lastPos)
	}
	return sz
}

// WriteGetRequestBodyToRequest writes Get request body field with given fields
// into buf. Returns number of bytes written.
func WriteGetRequestBodyToRequest(buf []byte, cnr [sha256.Size]byte, obj [sha256.Size]byte, raw bool, rngOff uint64, rngLen uint64, payloadOnly bool, extRngFirst *uint64, extRngLast *uint64) int {
	ln := CalculateGetRequestBodyLength(raw, rngLen, rngOff, payloadOnly, extRngFirst, extRngLast)
	if ln == 0 {
		return 0
	}
	off := proto.WriteRequestBodyTagAndLength(buf, ln)
	off += WriteGetRequestBody(buf[off:], cnr, obj, raw, rngOff, rngLen, payloadOnly, extRngFirst, extRngLast)
	return off
}

// WriteGetRequestBody writes Get request body message with given fields into
// buf. Returns number of bytes written.
func WriteGetRequestBody(buf []byte, cnr [sha256.Size]byte, obj [sha256.Size]byte, raw bool, rngOff uint64, rngLen uint64, payloadOnly bool, extRngFirst *uint64, extRngLast *uint64) int {
	off := refs.WriteObjectAddressField(buf, FieldGetRequestBodyAddress, cnr, obj)
	off += writeDynamicGetRequestBodyFields(buf[off:], raw, rngOff, rngLen, payloadOnly, extRngFirst, extRngLast)
	return off
}

func writeDynamicGetRequestBodyFields(buf []byte, raw bool, rngOff uint64, rngLen uint64, payloadOnly bool, extRngFirst *uint64, extRngLast *uint64) int {
	off := proto.MarshalToBool(buf, FieldGetRequestBodyRaw, raw)
	off += WriteRangeField(buf[off:], FieldGetRequestBodyRange, rngOff, rngLen)
	off += proto.MarshalToBool(buf[off:], FieldGetRequestBodyPayloadOnly, payloadOnly)
	off += WriteExtendedRangeField(buf[off:], FieldGetRequestBodyExtendedRange, extRngFirst, extRngLast)
	return off
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		var firstPos, lastPos *uint64
		if x.ExtendedRange != nil {
			firstPos, lastPos = x.ExtendedRange.FirstPos, x.ExtendedRange.LastPos
		}
		off := proto.MarshalToEmbedded(b, FieldGetRequestBodyAddress, x.Address)
		writeDynamicGetRequestBodyFields(b[off:], x.Raw, x.Range.GetOffset(), x.Range.GetLength(), x.PayloadOnly, firstPos, lastPos)
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

// CalculateHeadRequestBodyLength calculates length of Head request body message
// with static address and given dynamic fields.
func CalculateHeadRequestBodyLength(raw bool) int {
	ln := proto.SizeEmbeddedLENField(FieldHeadRequestBodyAddress, refs.ObjectAddressLength)
	ln += calculateDynamicHeadRequestBodyFieldsLength(false, raw)
	return ln
}

func calculateDynamicHeadRequestBodyFieldsLength(mainOnly bool, raw bool) int {
	ln := proto.SizeBool(FieldHeadRequestBodyMainOnly, mainOnly)
	ln += proto.SizeBool(FieldHeadRequestBodyRaw, raw)
	return ln
}

// MarshaledSize returns size of the HeadRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *HeadRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldHeadRequestBodyAddress, x.Address) +
			calculateDynamicHeadRequestBodyFieldsLength(x.MainOnly, x.Raw)
	}
	return sz
}

// WriteHeadRequestBodyToRequest writes Head request body field with given
// fields into buf. Returns number of bytes written.
func WriteHeadRequestBodyToRequest(buf []byte, cnr [sha256.Size]byte, obj [sha256.Size]byte, raw bool) int {
	ln := CalculateHeadRequestBodyLength(raw)
	if ln == 0 {
		return 0
	}
	off := proto.WriteRequestBodyTagAndLength(buf, ln)
	off += WriteHeadRequestBody(buf[off:], cnr, obj, raw)
	return off
}

// WriteHeadRequestBody writes Head request body message with given fields into
// buf. Returns number of bytes written.
func WriteHeadRequestBody(buf []byte, cnr [sha256.Size]byte, obj [sha256.Size]byte, raw bool) int {
	off := refs.WriteObjectAddressField(buf, FieldHeadRequestBodyAddress, cnr, obj)
	off += writeDynamicHeadRequestBodyFields(buf[off:], false, raw)
	return off
}

func writeDynamicHeadRequestBodyFields(buf []byte, mainOnly bool, raw bool) int {
	off := proto.MarshalToBool(buf, FieldHeadRequestBodyMainOnly, mainOnly)
	off += proto.MarshalToBool(buf[off:], FieldHeadRequestBodyRaw, raw)
	return off
}

// MarshalStable writes the HeadRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [HeadRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *HeadRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldHeadRequestBodyAddress, x.Address)
		writeDynamicHeadRequestBodyFields(b[off:], x.MainOnly, x.Raw)
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

// CalculateRangeLength calculates length of range message with given fields.
func CalculateRangeLength(rngOff uint64, rngLen uint64) int {
	ln := proto.SizeVarint(FieldRangeOffset, rngOff)
	ln += proto.SizeVarint(FieldRangeLength, rngLen)
	return ln
}

// MarshaledSize returns size of the Range in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Range) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateRangeLength(x.Offset, x.Length)
}

// WriteRangeField writes range field with given number and fields into
// buf. Returns number of bytes written.
func WriteRangeField(buf []byte, num protowire.Number, rngOff uint64, rngLen uint64) int {
	ln := CalculateRangeLength(rngOff, rngLen)
	if ln == 0 {
		return 0
	}
	off := proto.WriteTagAndLength(buf, num, ln)
	return off + WriteRange(buf[off:], rngOff, rngLen)
}

// WriteRange writes range message with given fields into buf. Returns number of
// bytes written.
func WriteRange(buf []byte, rngOff uint64, rngLen uint64) int {
	off := proto.MarshalToVarint(buf, FieldRangeOffset, rngOff)
	off += proto.MarshalToVarint(buf[off:], FieldRangeLength, rngLen)
	return off
}

// MarshalStable writes the Range in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Range.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Range) MarshalStable(b []byte) {
	if x != nil {
		WriteRange(b, x.Offset, x.Length)
	}
}

// Field numbers of [ExtendedRange] message.
const (
	_ = iota
	FieldExtendedRangeFirstPos
	FieldExtendedRangeLastPos
)

// CalculateExtendedRangeLength calculates length of extended range message with
// given fields.
func CalculateExtendedRangeLength(first *uint64, last *uint64) int {
	ln := proto.SizeOptionalVarint(FieldExtendedRangeFirstPos, first)
	ln += proto.SizeOptionalVarint(FieldExtendedRangeLastPos, last)
	return ln
}

// MarshaledSize returns size of the ExtendedRange in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *ExtendedRange) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateExtendedRangeLength(x.FirstPos, x.LastPos)
}

// WriteExtendedRangeField writes extended range field with given number and fields into
// buf. Returns number of bytes written.
func WriteExtendedRangeField(buf []byte, num protowire.Number, first *uint64, last *uint64) int {
	ln := CalculateExtendedRangeLength(first, last)
	if ln == 0 {
		return 0
	}
	off := proto.WriteTagAndLength(buf, num, ln)
	return off + WriteExtendedRange(buf[off:], first, last)
}

// WriteExtendedRange writes extended range message with given fields into buf.
// Returns number of bytes written.
func WriteExtendedRange(buf []byte, first *uint64, last *uint64) int {
	off := proto.MarshalToOptionalVarint(buf, FieldExtendedRangeFirstPos, first)
	off += proto.MarshalToOptionalVarint(buf[off:], FieldExtendedRangeLastPos, last)
	return off
}

// MarshalStable writes the ExtendedRange in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ExtendedRange.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ExtendedRange) MarshalStable(b []byte) {
	if x != nil {
		WriteExtendedRange(b, x.FirstPos, x.LastPos)
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
			proto.SizeVarint(FieldRangeHashRequestBodyType, x.Type) +
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
		proto.MarshalToVarint(b[off:], FieldRangeHashRequestBodyType, x.Type)
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
		sz = proto.SizeVarint(FieldRangeHashResponseBodyType, x.Type) +
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
		off := proto.MarshalToVarint(b, FieldRangeHashResponseBodyType, x.Type)
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

// CalculatePutInitRequestBodyLength calculates length of initial Put request
// body message with given fields.
func CalculatePutInitRequestBodyLength(initFldLen int) int {
	return proto.SizeEmbeddedLENField(FieldPutRequestBodyInit, initFldLen)
}

// CalculatePutChunkRequestBodyLength calculates length of chunk Put request body
// message with given fields.
func CalculatePutChunkRequestBodyLength(chunk []byte) int {
	return proto.SizeBytes(FieldPutRequestBodyChunk, chunk)
}

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
			sz = CalculatePutInitRequestBodyLength(p.Init.MarshaledSize())
		case *PutRequest_Body_Chunk:
			sz = CalculatePutChunkRequestBodyLength(p.Chunk)
		}
	}
	return sz
}

// WritePutInitRequestBodyToRequest writes initial Put request body field with
// given length and fields into buf. Returns number of bytes written.
func WritePutInitRequestBodyToRequest(buf []byte, ln int, initFldLen int, writeInitFldFn proto.WriteMessageFunc) int {
	off := proto.WriteRequestBodyTagAndLength(buf, ln)
	off += WritePutInitRequestBody(buf[off:], initFldLen, writeInitFldFn)
	return off
}

// WritePutInitRequestBody writes initial Put request body message with given
// length and fields into buf. Returns number of bytes written.
func WritePutInitRequestBody(buf []byte, initFldLen int, writeInitFldFn proto.WriteMessageFunc) int {
	return proto.WriteMessageField(buf, FieldPutRequestBodyInit, initFldLen, writeInitFldFn)
}

// WritePutChunkRequestBodyToRequest writes chunk Put request body field with
// given fields into buf. Returns number of bytes written.
func WritePutChunkRequestBodyToRequest(buf []byte, chunk []byte) int {
	ln := CalculatePutChunkRequestBodyLength(chunk)
	if ln == 0 {
		return 0
	}
	off := proto.WriteRequestBodyTagAndLength(buf, ln)
	off += WritePutChunkRequestBody(buf[off:], chunk)
	return off
}

// WritePutChunkRequestBody writes chunk Put request body message with given
// length and fields into buf. Returns number of bytes written.
func WritePutChunkRequestBody(buf []byte, chunk []byte) int {
	return proto.MarshalToBytes(buf, FieldPutRequestBodyChunk, chunk)
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
			initFldLen := p.Init.MarshaledSize()
			writeInitFldFn := proto.WriteStablyMarshalledMessageFunc(p.Init)
			WritePutInitRequestBody(b, initFldLen, writeInitFldFn)
		case *PutRequest_Body_Chunk:
			WritePutChunkRequestBody(b, p.Chunk)
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

// CalculateSearchFilterLength calculates length of search filter message with
// given fields.
func CalculateSearchFilterLength[MATCHER ~int32](matcher MATCHER, key string, value string) int {
	ln := proto.SizeVarint(FieldSearchFilterMatcher, matcher)
	ln += proto.SizeBytes(FieldSearchFilterKey, key)
	ln += proto.SizeBytes(FieldSearchFilterValue, value)
	return ln
}

// WriteSearchFilter writes search filter message with given fields into buf.
// Returns number of bytes written.
func WriteSearchFilter[MATCHER ~int32](buf []byte, matcher MATCHER, key string, value string) int {
	off := proto.MarshalToVarint(buf, FieldSearchFilterMatcher, matcher)
	off += proto.MarshalToBytes(buf[off:], FieldSearchFilterKey, key)
	off += proto.MarshalToBytes(buf[off:], FieldSearchFilterValue, value)
	return off
}

// MarshaledSize returns size of the SearchFilter in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SearchFilter) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateSearchFilterLength(x.MatchType, x.Key, x.Value)
}

// MarshalStable writes the SearchFilter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SearchFilter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SearchFilter) MarshalStable(b []byte) {
	if x != nil {
		WriteSearchFilter(b, x.MatchType, x.Key, x.Value)
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

// CalculateSearchV2RequestBodyLength calculates length of SearchV2 request body
// message with static container ID and given dynamic fields.
func CalculateSearchV2RequestBodyLength(version uint32, cursor string, count uint32, attributes []string, filterNum int, filterLenFn proto.RepeatedMessageLenFunc) int {
	ln := proto.SizeEmbeddedLENField(FieldSearchV2RequestBodyContainerID, refs.ContainerIDLength)
	ln += calculateDynamicSearchV2RequestBodyFieldsLength(version, cursor, count, attributes, filterNum, filterLenFn)
	return ln
}

func calculateDynamicSearchV2RequestBodyFieldsLength(version uint32, cursor string, count uint32, attributes []string, filterNum int, filterLenFn proto.RepeatedMessageLenFunc) int {
	ln := proto.SizeVarint(FieldSearchV2RequestBodyVersion, version)
	ln += proto.CalculateRepeatedFieldsLength(FieldSearchV2RequestBodyFilters, filterNum, filterLenFn)
	ln += proto.SizeBytes(FieldSearchV2RequestBodyCursor, cursor)
	ln += proto.SizeVarint(FieldSearchV2RequestBodyCount, count)
	ln += proto.SizeRepeatedBytes(FieldSearchV2RequestBodyAttributes, attributes)
	return ln
}

func (x *SearchV2Request_Body) getFilterLength(i int) int {
	f := x.Filters[i]
	if f == nil {
		return 0
	}
	return CalculateSearchFilterLength(f.MatchType, f.Key, f.Value)
}

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SearchV2Request_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldSearchV2RequestBodyContainerID, x.ContainerId) +
			calculateDynamicSearchV2RequestBodyFieldsLength(x.Version, x.Cursor, x.Count, x.Attributes, len(x.Filters), x.getFilterLength)
	}
	return 0
}

// WriteSearchV2RequestBodyToRequest writes SearchV2 request body field with
// given fields into buf. Returns number of bytes written.
func WriteSearchV2RequestBodyToRequest(buf []byte, cnr [sha256.Size]byte, version uint32, cursor string, count uint32, attributes []string, filterNum int, filterLenFn proto.RepeatedMessageLenFunc, writeFilterFn proto.WriteRepeatedMessageFunc) int {
	ln := CalculateSearchV2RequestBodyLength(version, cursor, count, attributes, filterNum, filterLenFn)
	if ln == 0 {
		return 0
	}
	off := proto.WriteRequestBodyTagAndLength(buf, ln)
	off += WriteSearchV2RequestBody(buf[off:], cnr, version, cursor, count, attributes, filterNum, filterLenFn, writeFilterFn)
	return off
}

// WriteSearchV2RequestBody writes SearchV2 request body message with given
// fields into buf. Returns number of bytes written.
func WriteSearchV2RequestBody(buf []byte, cnr [sha256.Size]byte, version uint32, cursor string, count uint32, attributes []string, filterNum int, filterLenFn proto.RepeatedMessageLenFunc, writeFilterFn proto.WriteRepeatedMessageFunc) int {
	off := refs.WriteContainerIDField(buf, FieldSearchV2RequestBodyContainerID, cnr)
	off += writeDynamicSearchV2RequestBodyFields(buf[off:], version, cursor, count, attributes, filterNum, filterLenFn, writeFilterFn)
	return off
}

func writeDynamicSearchV2RequestBodyFields(buf []byte, version uint32, cursor string, count uint32, attributes []string, filterNum int, filterLenFn proto.RepeatedMessageLenFunc, writeFilterFn proto.WriteRepeatedMessageFunc) int {
	off := proto.MarshalToVarint(buf, FieldSearchV2RequestBodyVersion, version)
	off += proto.WriteRepeatedFields(buf[off:], FieldSearchV2RequestBodyFilters, filterNum, filterLenFn, writeFilterFn)
	off += proto.MarshalToBytes(buf[off:], FieldSearchV2RequestBodyCursor, cursor)
	off += proto.MarshalToVarint(buf[off:], FieldSearchV2RequestBodyCount, count)
	off += proto.MarshalToRepeatedBytes(buf[off:], FieldSearchV2RequestBodyAttributes, attributes)
	return off
}

func (x *SearchV2Request_Body) writeFilter(buf []byte, i int) int {
	f := x.Filters[i]
	if f == nil {
		return 0
	}
	return WriteSearchFilter(buf, f.MatchType, f.Key, f.Value)
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SearchV2Request_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SearchV2Request_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSearchV2RequestBodyContainerID, x.ContainerId)
		writeDynamicSearchV2RequestBodyFields(b[off:], x.Version, x.Cursor, x.Count, x.Attributes, len(x.Filters), x.getFilterLength, x.writeFilter)
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
