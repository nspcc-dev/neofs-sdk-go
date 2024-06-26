// Package proto contains helper functions for Protocol Buffers
// (https://protobuf.dev) in addition to the ones from
// [google.golang.org/protobuf/encoding/protowire] package.
package proto

import (
	"encoding/binary"
	"math"

	"google.golang.org/protobuf/encoding/protowire"
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

// Bytes is a type parameter constraint for any byte arrays.
type Bytes interface{ ~[]byte | ~string }

// Varint is a type parameter constraint for any variable-length protobuf
// integers.
type Varint interface {
	~int32 | int64 | uint32 | uint64 // ~int32 for 'enum' fields
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
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.VarintType))
	return off + binary.PutUvarint(b[off:], uint64(v))
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
	return protowire.SizeTag(num) + protowire.SizeBytes(ln)
}

// MarshalToBytes encodes 'bytes' or 'string' protobuf field with given number and
// value into b and returns the number of bytes written. If the buffer is too
// small, MarshalToBytes will panic.
func MarshalToBytes[T Bytes](b []byte, num protowire.Number, v T) int {
	if len(v) == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.BytesType))
	off += binary.PutUvarint(b[off:], uint64(len(v)))
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
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.Fixed32Type))
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
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.Fixed64Type))
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
	if v == nil {
		return 0
	}
	sz := v.MarshaledSize()
	if sz == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeBytes(sz)
}

// MarshalToEmbedded encodes embedded message being a protobuf field with given
// number and value into b and returns the number of bytes written. If the
// buffer is too small, MarshalToEmbedded will panic.
func MarshalToEmbedded(b []byte, num protowire.Number, v Message) int {
	if v == nil {
		return 0
	}
	sz := v.MarshaledSize()
	if sz == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.BytesType))
	off += binary.PutUvarint(b[off:], uint64(sz))
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
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.BytesType))
	off += binary.PutUvarint(b[off:], uint64(sizeRepeatedVarint(v)))
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
		off += copy(b[off:], v[i])
	}
	return off
}
