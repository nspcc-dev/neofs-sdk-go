/*
Package proto contains helper functions in addition to the ones from
[google.golang.org/protobuf/encoding/protowire].
*/

package proto

import (
	"encoding/binary"
	"math"
	"reflect"

	"google.golang.org/protobuf/encoding/protowire"
)

// TODO: docs
type Message interface {
	MarshaledSize() int
	MarshalStable(b []byte)
}

func SizeVarint[integer uint64 | int64 | uint32 | int32](num protowire.Number, v integer) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeVarint(uint64(v))
}

func MarshalVarint[integer uint64 | int64 | uint32 | int32](b []byte, num protowire.Number, v integer) int {
	if v == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.VarintType))
	return off + binary.PutUvarint(b[off:], uint64(v))
}

func SizeBool(num protowire.Number, v bool) int {
	return SizeVarint(num, protowire.EncodeBool(v))
}

func MarshalBool(b []byte, num protowire.Number, v bool) int {
	return MarshalVarint(b, num, protowire.EncodeBool(v))
}

func SizeBytes[bytesOrString []byte | string](num protowire.Number, v bytesOrString) int {
	ln := len(v)
	if ln == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeBytes(ln)
}

func MarshalBytes[bytesOrString []byte | string](b []byte, num protowire.Number, v bytesOrString) int {
	if len(v) == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.BytesType))
	off += binary.PutUvarint(b[off:], uint64(len(v)))
	return off + copy(b[off:], v)
}

func SizeFixed32(num protowire.Number, v uint32) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeFixed32()
}

func MarshalFixed32(b []byte, num protowire.Number, v uint32) int {
	if v == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.Fixed32Type))
	binary.LittleEndian.PutUint32(b[off:], v)
	return off + protowire.SizeFixed32()
}

func SizeFloat32(num protowire.Number, v float32) int {
	return SizeFixed32(num, math.Float32bits(v))
}

func MarshalFloat32(b []byte, num protowire.Number, v float32) int {
	return MarshalFixed32(b, num, math.Float32bits(v))
}

func SizeFixed64(num protowire.Number, v uint64) int {
	if v == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeFixed64()
}

func MarshalFixed64(b []byte, num protowire.Number, v uint64) int {
	if v == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.Fixed64Type))
	binary.LittleEndian.PutUint64(b[off:], v)
	return off + protowire.SizeFixed64()
}

func SizeFloat64(num protowire.Number, v float64) int {
	return SizeFixed64(num, math.Float64bits(v))
}

func MarshalFloat64(b []byte, num protowire.Number, v float64) int {
	return MarshalFixed64(b, num, math.Float64bits(v))
}

func SizeNested(num protowire.Number, v Message) int {
	if v == nil || reflect.ValueOf(v).IsNil() {
		return 0
	}
	sz := v.MarshaledSize()
	if sz == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeBytes(sz)
}

func MarshalNested(b []byte, num protowire.Number, v Message) int {
	if v == nil || reflect.ValueOf(v).IsNil() {
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

func sizeRepeatedVarint(v []uint64) int {
	var sz int
	for i := range v {
		// packed https://protobuf.dev/programming-guides/encoding/#packed
		sz += protowire.SizeVarint(v[i])
	}
	return sz
}

func SizeRepeatedVarint(num protowire.Number, v []uint64) int {
	if len(v) == 0 {
		return 0
	}
	return protowire.SizeTag(num) + protowire.SizeBytes(sizeRepeatedVarint(v))
}

func MarshalRepeatedVarint(b []byte, num protowire.Number, v []uint64) int {
	if len(v) == 0 {
		return 0
	}
	off := binary.PutUvarint(b, protowire.EncodeTag(num, protowire.BytesType))
	off += binary.PutUvarint(b[off:], uint64(sizeRepeatedVarint(v)))
	for i := range v {
		off += binary.PutUvarint(b[off:], v[i])
	}
	return off
}

func sizeRepeatedBytes[bytesOrString []byte | string](num protowire.Number, v []bytesOrString) int {
	var sz int
	tagSz := protowire.SizeTag(num)
	for i := range v {
		// non-packed https://protobuf.dev/programming-guides/encoding/#packed
		sz += tagSz + protowire.SizeBytes(len(v[i]))
	}
	return sz
}

func SizeRepeatedBytes[bytesOrString []byte | string](num protowire.Number, v []bytesOrString) int {
	if len(v) == 0 {
		return 0
	}
	return sizeRepeatedBytes(num, v)
}

func MarshalRepeatedBytes[bytesOrString []byte | string](b []byte, num protowire.Number, v []bytesOrString) int {
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
