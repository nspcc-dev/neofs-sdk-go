// Package proto contains helper functions for Protocol Buffers
// (https://protobuf.dev) in addition to the ones from
// [google.golang.org/protobuf/encoding/protowire] package.
package proto

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// Various function aliases.
type (
	RepeatedMessageLenFunc   = func(int) int
	WriteMessageFunc         = func(buf []byte) int
	WriteRepeatedMessageFunc = func(buf []byte, i int) int
)

// Common request field numbers.
const (
	FieldRequestBody               = 1
	FieldRequestMetaHeader         = 2
	FieldRequestVerificationHeader = 3
)

// Message is provided by protobuf 'message' types used in NeoFS for so-called
// stable marshaling: protobuf with the order of fields in strict ascending
// order of their numbers.
type Message interface {
	// MarshaledSize returns size of the encoded Message in bytes.
	MarshaledSize() int
	// MarshalStable encodes the Message into b. If the buffer is too small,
	// MarshalStable will panic.
	MarshalStable(b []byte)
}

// MarshalMessage encodes m into a dynamically allocated buffer.
func MarshalMessage(m Message) []byte {
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Marshal encodes v transmitted via NeoFS API protocol into a dynamically
// allocated buffer.
func Marshal[M Message, T interface{ ProtoMessage() M }](v T) []byte {
	return MarshalMessage(v.ProtoMessage())
}

// UnmarshalMessage decodes m from b.
func UnmarshalMessage(b []byte, m proto.Message) error { return proto.Unmarshal(b, m) }

// Unmarshal decodes v transmitted via NeoFS API protocol from b.
func Unmarshal[M_ any, M interface {
	*M_
	proto.Message
	Message
}, T interface{ FromProtoMessage(M) error }](b []byte, v T) error {
	m := M(new(M_))
	if err := UnmarshalMessage(b, m); err != nil {
		return err
	}
	return v.FromProtoMessage(m)
}

// UnmarshalOptional decodes v from [protojson] with ignored missing required
// fields.
func UnmarshalOptional[M_ any, M interface {
	*M_
	proto.Message
	Message
}, T interface{ FromProtoMessage(M) error }](b []byte, v T, dec func(T, M, bool) error) error {
	m := M(new(M_))
	if err := UnmarshalMessage(b, m); err != nil {
		return err
	}
	return dec(v, m, false)
}

var jOpts = protojson.MarshalOptions{EmitUnpopulated: true}

// MarshalMessageJSON encodes m into [protojson] with unpopulated fields'
// emission.
func MarshalMessageJSON(m proto.Message) ([]byte, error) { return jOpts.Marshal(m) }

// MarshalJSON encodes v into [protojson] with unpopulated fields' emission.
func MarshalJSON[M proto.Message, T interface{ ProtoMessage() M }](v T) ([]byte, error) {
	return MarshalMessageJSON(v.ProtoMessage())
}

// UnmarshalMessageJSON decodes m from [protojson].
func UnmarshalMessageJSON(b []byte, m proto.Message) error { return protojson.Unmarshal(b, m) }

// UnmarshalJSON decodes v from [protojson].
func UnmarshalJSON[M_ any, M interface {
	*M_
	proto.Message
	Message
}, T interface{ FromProtoMessage(M) error }](b []byte, v T) error {
	m := M(new(M_))
	if err := UnmarshalMessageJSON(b, m); err != nil {
		return err
	}
	return v.FromProtoMessage(m)
}

// UnmarshalJSONOptional decodes v from [protojson] with ignored missing
// required fields.
func UnmarshalJSONOptional[M_ any, M interface {
	*M_
	proto.Message
	Message
}, T interface{ FromProtoMessage(M) error }](b []byte, v T, dec func(T, M, bool) error) error {
	m := M(new(M_))
	if err := UnmarshalMessageJSON(b, m); err != nil {
		return err
	}
	return dec(v, m, false)
}

// Bytes is a type parameter constraint for any byte arrays.
type Bytes interface{ ~[]byte | ~string }

// Varint is a type parameter constraint for any variable-length protobuf
// integers.
type Varint interface {
	~int32 | int64 | uint32 | uint64 | int // ~int32 for 'enum' fields
}

// SizeVarint returns the encoded size of varint protobuf field with given
// number and value.
func SizeVarint[T Varint](num protowire.Number, v T) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeVarint(uint64(v))
}

// MarshalToVarint encodes varint protobuf field with given number and value into
// b and returns the number of bytes written. If the buffer is too small,
// MarshalToVarint will panic.
func MarshalToVarint[T Varint](b []byte, num protowire.Number, v T) int {
	if v == 0 {
		return 0
	}
	return WriteTagAndVarint(b, num, protowire.VarintType, v)
}

// SizeOptionalVarint returns the encoded size of an optional varint protobuf
// field with given number and value.
func SizeOptionalVarint[T Varint](num protowire.Number, v *T) int {
	if v == nil {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeVarint(uint64(*v))
}

// MarshalToOptionalVarint encodes an optional varint protobuf field with given
// number and value into b and returns the number of bytes written. If the buffer
// is too small, MarshalToOptionalVarint will panic.
func MarshalToOptionalVarint[T Varint](b []byte, num protowire.Number, v *T) int {
	if v == nil {
		return 0
	}
	return WriteTagAndVarint(b, num, protowire.VarintType, *v)
}

// SizeBool returns the encoded size of 'bool' protobuf field with given number
// and value.
func SizeBool(num protowire.Number, v bool) int {
	return SizeVarint(num, protowire.EncodeBool(v))
}

// MarshalToBool encodes 'bool' protobuf field with given number and value into b
// and returns the number of bytes written. If the buffer is too small,
// MarshalToBool will panic.
func MarshalToBool(b []byte, num protowire.Number, v bool) int {
	return MarshalToVarint(b, num, protowire.EncodeBool(v))
}

// SizeBytes returns the encoded size of 'bytes' or 'string' protobuf field with
// given number and value.
func SizeBytes[T Bytes](num protowire.Number, v T) int {
	ln := len(v)
	if ln == 0 {
		return 0
	}
	return sizeEmbeddedLENField(num, ln)
}

// MarshalToBytes encodes 'bytes' or 'string' protobuf field with given number and
// value into b and returns the number of bytes written. If the buffer is too
// small, MarshalToBytes will panic.
func MarshalToBytes[T Bytes](b []byte, num protowire.Number, v T) int {
	if len(v) == 0 {
		return 0
	}
	off := WriteTagAndLength(b, num, len(v))
	if len(b[off:]) < len(v) {
		panic("too short buffer")
	}
	return off + copy(b[off:], v)
}

// SizeFixed32 returns the encoded size of 'fixed32' protobuf field with given
// number and value.
func SizeFixed32(num protowire.Number, v uint32) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeFixed32()
}

// MarshalToFixed32 encodes 'fixed32' protobuf field with given number and value
// into b and returns the number of bytes written. If the buffer is too small,
// MarshalToFixed32 will panic.
func MarshalToFixed32(b []byte, num protowire.Number, v uint32) int {
	if v == 0 {
		return 0
	}
	off := WriteTag(b, num, protowire.Fixed32Type)
	binary.LittleEndian.PutUint32(b[off:], v)
	return off + protowire.SizeFixed32()
}

// SizeFixed64 returns the encoded size of 'fixed64' protobuf field with given
// number and value.
func SizeFixed64(num protowire.Number, v uint64) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeFixed64()
}

// MarshalToFixed64 encodes 'fixed64' protobuf field with given number and value
// into b and returns the number of bytes written. If the buffer is too small,
// MarshalToFixed64 will panic.
func MarshalToFixed64(b []byte, num protowire.Number, v uint64) int {
	if v == 0 {
		return 0
	}
	off := WriteTag(b, num, protowire.Fixed64Type)
	binary.LittleEndian.PutUint64(b[off:], v)
	return off + protowire.SizeFixed64()
}

// SizeFloat returns the encoded size of 'float' protobuf field with given
// number and value.
func SizeFloat(num protowire.Number, v float32) int {
	return SizeFixed32(num, math.Float32bits(v))
}

// MarshalToFloat encodes 'float' protobuf field with given number and value into
// b and returns the number of bytes written. If the buffer is too small,
// MarshalToFloat will panic.
func MarshalToFloat(b []byte, num protowire.Number, v float32) int {
	return MarshalToFixed32(b, num, math.Float32bits(v))
}

// SizeDouble returns the encoded size of 'double' protobuf field with given
// number and value.
func SizeDouble(num protowire.Number, v float64) int {
	return SizeFixed64(num, math.Float64bits(v))
}

// MarshalToDouble encodes 'double' protobuf field with given number and value
// into b and returns the number of bytes written. If the buffer is too small,
// MarshalToDouble will panic.
func MarshalToDouble(b []byte, num protowire.Number, v float64) int {
	return MarshalToFixed64(b, num, math.Float64bits(v))
}

// SizeEmbedded returns the encoded size of embedded message being a protobuf
// field with given number and value.
func SizeEmbedded(num protowire.Number, v Message) int {
	if isMessageNil(v) {
		return 0
	}
	return sizeEmbeddedLENField(num, v.MarshaledSize())
}

// MarshalToEmbedded encodes embedded message being a protobuf field with given
// number and value into b and returns the number of bytes written. If the
// buffer is too small, MarshalToEmbedded will panic.
func MarshalToEmbedded(b []byte, num protowire.Number, v Message) int {
	if isMessageNil(v) {
		return 0
	}
	return marshalToEmbeddedLength(b, num, v.MarshaledSize(), v)
}

// MarshalToEmbeddedLength encodes embedded message being a protobuf field with
// given number, length and value into b and returns the number of bytes
// written. If the buffer is too small, MarshalToEmbeddedLength will panic.
//
// Does not write field if message length is zero.
func MarshalToEmbeddedLength(b []byte, num protowire.Number, sz int, v Message) int {
	if sz == 0 {
		return 0
	}
	return marshalToEmbeddedLength(b, num, sz, v)
}

func marshalToEmbeddedLength(b []byte, num protowire.Number, sz int, v Message) int {
	off := WriteTagAndLength(b, num, sz)
	v.MarshalStable(b[off:])
	return off + sz
}

func sizeRepeatedVarint[T Varint](v []T) int {
	var sz int
	for i := range v {
		// packed (https://protobuf.dev/programming-guides/encoding/#packed)
		sz += protowire.SizeVarint(uint64(v[i]))
	}
	return sz
}

// SizeRepeatedVarint returns the encoded size of 'repeated' varint protobuf
// field with given number and value.
func SizeRepeatedVarint[T Varint](num protowire.Number, v []T) int {
	if len(v) == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeBytes(sizeRepeatedVarint(v))
}

// MarshalToRepeatedVarint encodes 'repeated' varint protobuf field with given
// number and value into b and returns the number of bytes written. If the
// buffer is too small, MarshalToRepeatedVarint will panic.
func MarshalToRepeatedVarint[T Varint](b []byte, num protowire.Number, v []T) int {
	if len(v) == 0 {
		return 0
	}
	off := WriteTagAndLength(b, num, sizeRepeatedVarint(v))
	for i := range v {
		off += binary.PutUvarint(b[off:], uint64(v[i]))
	}
	return off
}

// SizeRepeatedBytes returns the encoded size of 'repeated bytes' or 'repeated
// string' protobuf field with given number and value.
func SizeRepeatedBytes[T Bytes](num protowire.Number, v []T) int {
	if len(v) == 0 {
		return 0
	}
	var sz int
	tagSz := protowire.SizeTag(num)
	for i := range v {
		// non-packed (https://protobuf.dev/programming-guides/encoding/#packed)
		sz += tagSz + protowire.SizeBytes(len(v[i]))
	}
	return sz
}

// MarshalToRepeatedBytes encodes 'repeated bytes' or 'repeated string' protobuf
// field with given number and value into b and returns the number of bytes
// written. If the buffer is too small, MarshalToRepeatedBytes will panic.
func MarshalToRepeatedBytes[T Bytes](b []byte, num protowire.Number, v []T) int {
	if len(v) == 0 {
		return 0
	}
	var off int
	tag := protowire.EncodeTag(num, protowire.BytesType)
	for i := range v {
		off += binary.PutUvarint(b[off:], tag)
		off += binary.PutUvarint(b[off:], uint64(len(v[i])))
		if len(b[off:]) < len(v[i]) {
			panic("too short buffer")
		}
		off += copy(b[off:], v[i])
	}
	return off
}

// SizeRepeatedMessages returns the encoded size of 'repeated M' protobuf field
// with given number and values.
func SizeRepeatedMessages[T any, M interface {
	*T
	Message
}](num protowire.Number, v []M) int {
	var sz int
	var zero T
	for i := range v {
		if !isMessageNil(v[i]) {
			sz += SizeEmbedded(num, v[i])
		} else {
			sz += SizeEmbedded(num, M(&zero))
		}
	}
	return sz
}

// MarshalToRepeatedMessages encodes 'repeated M' protobuf field with given
// number and values into b and returns the number of bytes written. If the
// buffer is too small, MarshalToRepeatedMessages will panic.
func MarshalToRepeatedMessages[T any, M interface {
	*T
	Message
}](b []byte, num protowire.Number, v []M) int {
	if len(v) == 0 {
		return 0
	}
	var off int
	var zero T
	for i := range v {
		if !isMessageNil(v[i]) {
			off += MarshalToEmbedded(b[off:], num, v[i])
		} else {
			off += MarshalToEmbedded(b[off:], num, M(&zero))
		}
	}
	return off
}

// SizeEmbeddedLENField returns the encoded size of embedded LEN protobuf field
// with given number and length. Returns zero for zero length.
func SizeEmbeddedLENField(num protowire.Number, ln int) int {
	if ln < 0 {
		panic(fmt.Sprintf("negative length %d", ln))
	}
	if !num.IsValid() {
		panic(fmt.Sprintf("invalid field number %d", num))
	}
	if ln == 0 {
		return 0
	}
	return sizeEmbeddedLENField(num, ln)
}

func sizeEmbeddedLENField(num protowire.Number, ln int) int {
	return protowire.SizeTag(num) + protowire.SizeBytes(ln)
}

func isMessageNil(m Message) bool {
	return m == nil || reflect.ValueOf(m).IsNil()
}

// EncodeRequest ONLY correctly encodes requests that have strictly ordered two
// fields (excepting verification header): field #1 is body, field #2 is meta
// header. Any other requests must be encoded differently. The second returned
// value means buffer len occupied for req.
func EncodeRequest[B, M Message](buf []byte, reqBody B, reqMetaHeader M) ([]byte, int) {
	var (
		size int
		bLen = reqBody.MarshaledSize()
		mLen = reqMetaHeader.MarshaledSize()
		off  int
	)

	size = CalculateRequestBodyWithMetaHeaderLength(bLen, mLen)
	if len(buf) < size {
		buf = make([]byte, size)
	}

	off = WriteRequestBodyMessage(buf, reqBody)
	off += WriteRequestMetaHeaderMessage(buf[off:], reqMetaHeader)

	return buf, off
}

// WriteTag writes tag for the field of given number and type into buf. Returns
// number of bytes written.
func WriteTag(buf []byte, num protowire.Number, typ protowire.Type) int {
	return binary.PutUvarint(buf, protowire.EncodeTag(num, typ))
}

// WriteTagAndLength writes tag for the LEN field of given number and its length
// into buf. Returns number of bytes written.
func WriteTagAndLength(buf []byte, num protowire.Number, ln int) int {
	return WriteTagAndVarint(buf, num, protowire.BytesType, ln)
}

// WriteTagAndVarint writes tag for the field of given number and type followed
// by varint into buf. Returns number of bytes written.
func WriteTagAndVarint[T Varint](buf []byte, num protowire.Number, typ protowire.Type, v T) int {
	off := WriteTag(buf, num, typ)
	return off + binary.PutUvarint(buf[off:], uint64(v))
}

// WriteStablyMarshalledMessageFunc returns function for m writing.
func WriteStablyMarshalledMessageFunc(m Message) WriteMessageFunc {
	return func(buf []byte) int {
		if isMessageNil(m) {
			return 0
		}
		ln := m.MarshaledSize()
		if ln == 0 {
			return 0
		}
		m.MarshalStable(buf)
		return ln
	}
}

// WriteStablyMarshalledMessageField writes m field with given number into buf.
// Returns number of bytes written.
func WriteStablyMarshalledMessageField(buf []byte, num protowire.Number, m Message) int {
	ln := m.MarshaledSize()
	return WriteMessageField(buf, num, ln, func(buf []byte) int {
		m.MarshalStable(buf)
		return ln
	})
}

// WriteMessageField writes LEN message with given field number into buf.
// Returns number of bytes written.
func WriteMessageField(buf []byte, num protowire.Number, ln int, writeFn WriteMessageFunc) int {
	if ln == 0 {
		return 0
	}
	off := WriteTagAndLength(buf, num, ln)
	off += writeFn(buf[off:])
	return off
}

// CalculateRepeatedFieldsLength returns length of specified number of repeated
// messages with given field number.
func CalculateRepeatedFieldsLength(num protowire.Number, count int, lenFn RepeatedMessageLenFunc) int {
	var ln int
	for i := range count {
		lni := lenFn(i)
		if lni == 0 {
			ln += protowire.SizeTag(num) + protowire.SizeVarint(uint64(lni))
		} else {
			ln += SizeEmbeddedLENField(num, lni)
		}
	}
	return ln
}

// WriteRepeatedFields writes specified number of repeated messages with given
// field number into buf.
func WriteRepeatedFields(buf []byte, num protowire.Number, count int, lenFn RepeatedMessageLenFunc, writeFn WriteRepeatedMessageFunc) int {
	var off int
	for i := range count {
		off += WriteTagAndLength(buf[off:], num, lenFn(i))
		off += writeFn(buf[off:], i)
	}
	return off
}

// CalculateRequestBodyFieldLength returns length of field for the request body
// with given length.
func CalculateRequestBodyFieldLength(ln int) int {
	return SizeEmbeddedLENField(FieldRequestBody, ln)
}

// WriteRequestBodyTagAndLength writes request body field tag and length into
// buf. Returns number of bytes written.
func WriteRequestBodyTagAndLength(buf []byte, ln int) int {
	return WriteTagAndLength(buf, FieldRequestBody, ln)
}

// WriteRequestBodyMessage writes request body field into buf. Returns number of
// bytes written.
func WriteRequestBodyMessage(buf []byte, m Message) int {
	return WriteStablyMarshalledMessageField(buf, FieldRequestBody, m)
}

// CalculateRequestMetaHeaderFieldLength returns length of field for the request meta header
// with given length.
func CalculateRequestMetaHeaderFieldLength(ln int) int {
	return SizeEmbeddedLENField(FieldRequestMetaHeader, ln)
}

// WriteRequestMetaHeaderTagAndLength writes request meta header field tag and
// length into buf. Returns number of bytes written.
func WriteRequestMetaHeaderTagAndLength(buf []byte, metaHdrLen int) int {
	return WriteTagAndLength(buf, FieldRequestMetaHeader, metaHdrLen)
}

// WriteRequestMetaHeaderMessage writes request meta header field into buf.
// Returns number of bytes written.
func WriteRequestMetaHeaderMessage(buf []byte, m Message) int {
	return WriteStablyMarshalledMessageField(buf, FieldRequestMetaHeader, m)
}

// CalculateRequestVerificationHeaderFieldLength returns length of field for the
// request verification header with given length.
func CalculateRequestVerificationHeaderFieldLength(ln int) int {
	return SizeEmbeddedLENField(FieldRequestVerificationHeader, ln)
}

// WriteRequestVerificationHeaderTagAndLength writes request verification header
// field tag and length into buf. Returns number of bytes written.
func WriteRequestVerificationHeaderTagAndLength(buf []byte, ln int) int {
	return WriteTagAndLength(buf, FieldRequestVerificationHeader, ln)
}

// WriteRequestVerificationHeaderMessage writes request verification header
// field into buf. Returns number of bytes written.
func WriteRequestVerificationHeaderMessage(buf []byte, m Message) int {
	return WriteStablyMarshalledMessageField(buf, FieldRequestVerificationHeader, m)
}

// CalculateRequestBodyWithMetaHeaderLength returns length of request body and
// meta header fields with specified lengths.
func CalculateRequestBodyWithMetaHeaderLength(bodyLen int, metaHdrLen int) int {
	ln := CalculateRequestBodyFieldLength(bodyLen)
	ln += CalculateRequestMetaHeaderFieldLength(metaHdrLen)
	return ln
}

// CalculateRequestLength returns length of request body, meta header and
// verification header fields with specified lengths.
func CalculateRequestLength(bodyLen int, metaHdrLen int, verifHdrLen int) int {
	ln := CalculateRequestBodyWithMetaHeaderLength(bodyLen, metaHdrLen)
	ln += CalculateRequestVerificationHeaderFieldLength(verifHdrLen)
	return ln
}
