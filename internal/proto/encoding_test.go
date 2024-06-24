package proto_test

import (
	"math"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestVarint(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeVarint(anyFieldNum, int32(0)))
	require.Zero(t, proto.MarshalVarint(nil, anyFieldNum, int32(0)))

	const v = int32(42)
	sz := proto.SizeVarint(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalVarint(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.VarintType, typ)

	res, _ := protowire.ConsumeVarint(b[tagLn:])
	require.EqualValues(t, v, res)
}

func TestBool(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeBool(anyFieldNum, false))
	require.Zero(t, proto.MarshalBool(nil, anyFieldNum, false))

	const v = true
	sz := proto.SizeBool(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalBool(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.VarintType, typ)

	res, _ := protowire.ConsumeVarint(b[tagLn:])
	require.EqualValues(t, 1, res)
}

func TestFixed32(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeFixed32(anyFieldNum, 0))
	require.Zero(t, proto.MarshalFixed32(nil, anyFieldNum, 0))

	const v = 42
	sz := proto.SizeFixed32(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalFixed32(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.Fixed32Type, typ)

	res, _ := protowire.ConsumeFixed32(b[tagLn:])
	require.EqualValues(t, v, res)
}

func TestFloat32(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeFloat32(anyFieldNum, 0.0))
	require.Zero(t, proto.MarshalFloat32(nil, anyFieldNum, 0.0))

	const v = float32(1234.5678)
	sz := proto.SizeFloat32(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalFloat32(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.Fixed32Type, typ)

	res, _ := protowire.ConsumeFixed32(b[tagLn:])
	require.EqualValues(t, v, math.Float32frombits(res))
}

func TestFixed64(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeFixed64(anyFieldNum, 0))
	require.Zero(t, proto.MarshalFixed64(nil, anyFieldNum, 0))

	const v = 42
	sz := proto.SizeFixed64(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalFixed64(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.Fixed64Type, typ)

	res, _ := protowire.ConsumeFixed64(b[tagLn:])
	require.EqualValues(t, v, res)
}

func TestFloat64(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeFloat64(anyFieldNum, 0.0))
	require.Zero(t, proto.MarshalFloat64(nil, anyFieldNum, 0.0))

	const v = 1234.5678
	sz := proto.SizeFloat64(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalFloat64(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.Fixed64Type, typ)

	res, _ := protowire.ConsumeFixed64(b[tagLn:])
	require.EqualValues(t, v, math.Float64frombits(res))
}

func TestBytes(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeBytes(anyFieldNum, []byte(nil)))
	require.Zero(t, proto.MarshalBytes(nil, anyFieldNum, []byte(nil)))
	require.Zero(t, proto.SizeBytes(anyFieldNum, []byte{}))
	require.Zero(t, proto.MarshalBytes(nil, anyFieldNum, []byte{}))
	require.Zero(t, proto.SizeBytes(anyFieldNum, ""))
	require.Zero(t, proto.MarshalBytes(nil, anyFieldNum, ""))

	const v = "Hello, world!"
	sz := proto.SizeBytes(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalBytes(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.BytesType, typ)

	res, _ := protowire.ConsumeBytes(b[tagLn:])
	require.EqualValues(t, v, res)
}

type nestedMessageStub string

func (x *nestedMessageStub) MarshaledSize() int     { return len(*x) }
func (x *nestedMessageStub) MarshalStable(b []byte) { copy(b, *x) }

func TestNested(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeNested(anyFieldNum, nil))
	require.Zero(t, proto.MarshalNested(nil, anyFieldNum, nil))
	require.Zero(t, proto.SizeNested(anyFieldNum, (*nestedMessageStub)(nil)))
	require.Zero(t, proto.MarshalNested(nil, anyFieldNum, (*nestedMessageStub)(nil)))

	v := nestedMessageStub("Hello, world!")
	sz := proto.SizeNested(anyFieldNum, &v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalNested(b, anyFieldNum, &v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.BytesType, typ)

	res, _ := protowire.ConsumeBytes(b[tagLn:])
	require.EqualValues(t, v, res)
}

func TestRepeatedVarint(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeRepeatedVarint(anyFieldNum, nil))
	require.Zero(t, proto.MarshalRepeatedVarint(nil, anyFieldNum, nil))
	require.Zero(t, proto.SizeRepeatedVarint(anyFieldNum, []uint64{}))
	require.Zero(t, proto.MarshalRepeatedVarint(nil, anyFieldNum, []uint64{}))

	v := []uint64{12, 345, 0, 67890} // unlike single varint, zero must be explicitly encoded
	sz := proto.SizeRepeatedVarint(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalRepeatedVarint(b, anyFieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, anyFieldNum, num)
	require.EqualValues(t, protowire.BytesType, typ)

	b, _ = protowire.ConsumeBytes(b[tagLn:])
	var res []uint64
	for len(b) > 0 {
		i, ln := protowire.ConsumeVarint(b)
		require.Positive(t, tagLn, protowire.ParseError(ln))
		res = append(res, i)
		b = b[ln:]
	}
	require.Equal(t, v, res)
}

func TestRepeatedBytes(t *testing.T) {
	const anyFieldNum = 123
	require.Zero(t, proto.SizeRepeatedBytes(anyFieldNum, [][]byte(nil)))
	require.Zero(t, proto.MarshalRepeatedBytes(nil, anyFieldNum, [][]byte(nil)))
	require.Zero(t, proto.SizeRepeatedBytes(anyFieldNum, [][]byte{}))
	require.Zero(t, proto.MarshalRepeatedBytes(nil, anyFieldNum, [][]byte{}))
	require.Zero(t, proto.SizeRepeatedBytes(anyFieldNum, []string(nil)))
	require.Zero(t, proto.MarshalRepeatedBytes(nil, anyFieldNum, []string(nil)))
	require.Zero(t, proto.SizeRepeatedBytes(anyFieldNum, []string{}))
	require.Zero(t, proto.MarshalRepeatedBytes(nil, anyFieldNum, []string{}))

	v := []string{"Hello", "World", "", "Bob", "Alice"} // unlike single byte array, zero must be explicitly encoded
	sz := proto.SizeRepeatedBytes(anyFieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, proto.MarshalRepeatedBytes(b, anyFieldNum, v))

	var res []string
	for len(b) > 0 {
		num, typ, tagLn := protowire.ConsumeTag(b)
		require.Positive(t, tagLn, protowire.ParseError(tagLn))
		require.EqualValues(t, anyFieldNum, num)
		require.EqualValues(t, protowire.BytesType, typ)

		bs, ln := protowire.ConsumeBytes(b[tagLn:])
		require.Positive(t, tagLn, protowire.ParseError(ln))
		res = append(res, string(bs))
		b = b[tagLn+ln:]
	}
	require.Equal(t, v, res)
}
