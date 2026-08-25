package encoding_test

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	stdproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type enum int32

type timestamp timestamppb.Timestamp

func (x *timestamp) MarshaledSize() int {
	var sz int
	if x != nil {
		if x.Seconds != 0 {
			sz += protowire.SizeTag(1) + protowire.SizeVarint(uint64(x.Seconds))
		}
		if x.Nanos != 0 {
			sz += protowire.SizeTag(2) + protowire.SizeVarint(uint64(x.Nanos))
		}
	}
	return sz
}

func (x *timestamp) MarshalStable(b []byte) {
	if x != nil {
		b2 := b[:0:cap(b)]
		if x.Seconds != 0 {
			b2 = protowire.AppendTag(b2, 1, protowire.VarintType)
			b2 = protowire.AppendVarint(b2, uint64(x.Seconds))
		}
		if x.Nanos != 0 {
			b2 = protowire.AppendTag(b2, 2, protowire.VarintType)
			protowire.AppendVarint(b2, uint64(x.Nanos))
		}
	}
}

func randFieldNum() protowire.Number {
	n := protowire.Number(rand.Uint32() % uint32(protowire.MaxValidNumber))
	if n < protowire.MinValidNumber {
		return protowire.MinValidNumber
	}
	return n
}

func randBytes[T protoencoding.Bytes]() T {
	ln := rand.Uint32() % 64
	if ln == 0 {
		ln = 1
	}
	return T(testutil.RandByteSlice(ln))
}

func randVarint[T protoencoding.Varint]() T { return T(rand.Uint64()) }

func randRepeatedBytes[T protoencoding.Bytes]() []T {
	n := rand.Uint32() % 10
	if n == 0 {
		n = 1
	}
	res := make([]T, n)
	for i := range res {
		res[i] = randBytes[T]()
	}
	// unlike non-repeated field, zero element of repeated field must be presented
	res[rand.Uint32()%n] = T("")
	return res
}

func randRepeatedVarint[T protoencoding.Varint]() []T {
	n := rand.Uint32() % 10
	if n == 0 {
		n = 1
	}
	res := make([]T, n)
	for i := range res {
		res[i] = T(rand.Uint64())
	}
	// unlike non-repeated field, zero element of repeated field must be presented
	res[rand.Uint32()%n] = 0
	return res
}

func consumeVarint[T protoencoding.Varint](b []byte) (T, int) {
	v, ln := protowire.ConsumeVarint(b)
	return T(v), ln
}

func consumeRepeatedVarint[T protoencoding.Varint](b []byte) ([]T, int) {
	bs, code := protowire.ConsumeBytes(b)
	if code < 0 {
		return nil, code
	}
	var res []T
	for len(bs) > 0 {
		i, ln := protowire.ConsumeVarint(bs)
		if ln < 0 {
			return nil, ln
		}
		res = append(res, T(i))
		bs = bs[ln:]
	}
	return res, 0
}

func consumeBytes[T protoencoding.Bytes](b []byte) (T, int) {
	v, ln := protowire.ConsumeBytes(b)
	return T(v), ln
}

func consumePackedRepeated[T protoencoding.Bytes](b []byte, num protowire.Number) ([]T, int) {
	var res []T
	for len(b) > 0 {
		n, t, tagLn := protowire.ConsumeTag(b)
		if tagLn < 0 {
			return nil, tagLn
		} else if n != num {
			return nil, -2
		} else if t != protowire.BytesType {
			return nil, math.MinInt
		}

		bs, ln := protowire.ConsumeBytes(b[tagLn:])
		if ln < 0 {
			return nil, ln
		}
		res = append(res, T(bs))
		b = b[tagLn+ln:]
	}
	return res, 0
}

type anySupportedType interface {
	protoencoding.Varint | protoencoding.Bytes | bool | float32 | float64 | []uint64 | []uint32 | []int64 | []int32 | []enum
}

func testEncoding[T anySupportedType](
	t testing.TB,
	wireType protowire.Type,
	sizeFunc func(protowire.Number, T) int,
	marshalFunc func([]byte, protowire.Number, T) int,
	consumeFunc func([]byte) (T, int),
	randFunc func() T,
) {
	var v T
	fieldNum := randFieldNum()
	require.Zero(t, sizeFunc(fieldNum, v), fieldNum)
	require.Zero(t, marshalFunc(nil, fieldNum, v), fieldNum)

	v = randFunc()
	msg := fmt.Sprintf("num=%d,val=%v", fieldNum, v)

	sz := sizeFunc(fieldNum, v)
	if sz > 0 {
		require.Panics(t, func() { marshalFunc(make([]byte, sz-1), fieldNum, v) }, msg)
	}
	b := make([]byte, sz)
	require.EqualValues(t, sz, marshalFunc(b, fieldNum, v), msg)

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.NoError(t, protowire.ParseError(tagLn), msg)
	require.EqualValues(t, fieldNum, num, msg)
	require.Equal(t, wireType, typ, msg)

	res, code := consumeFunc(b[tagLn:])
	require.NoError(t, protowire.ParseError(code))
	require.EqualValues(t, v, res, msg)
}

func testPackedRepeated[T protoencoding.Bytes](t testing.TB,
	sizeFunc func(protowire.Number, []T) int,
	marshalFunc func([]byte, protowire.Number, []T) int,
	consumeFunc func([]byte, protowire.Number) ([]T, int),
	randFunc func() []T,
) {
	var v []T
	fieldNum := randFieldNum()
	require.Zero(t, sizeFunc(fieldNum, v), fieldNum)
	require.Zero(t, marshalFunc(nil, fieldNum, v), fieldNum)
	require.Zero(t, sizeFunc(fieldNum, []T{}), fieldNum)
	require.Zero(t, marshalFunc(nil, fieldNum, []T{}), fieldNum)

	v = randFunc()
	msg := fmt.Sprintf("num=%d,val=%v", fieldNum, v)

	sz := sizeFunc(fieldNum, v)
	if sz > 0 {
		require.Panics(t, func() { marshalFunc(make([]byte, sz-1), fieldNum, v) }, msg)
	}
	b := make([]byte, sz)
	require.EqualValues(t, sz, marshalFunc(b, fieldNum, v), msg)

	res, code := consumeFunc(b, fieldNum)
	require.NoError(t, protowire.ParseError(code))
	require.EqualValues(t, v, res, msg)
}

func benchmarkType[T anySupportedType](
	b *testing.B,
	v T,
	sizeFunc func(protowire.Number, T) int,
	marshalFunc func([]byte, protowire.Number, T) int,
) {
	const fieldNum = protowire.MaxValidNumber
	buf := make([]byte, sizeFunc(fieldNum, v))
	b.ReportAllocs()
	for b.Loop() {
		marshalFunc(buf, fieldNum, v)
	}
}

func benchmarkRepeatedType[T anySupportedType](
	b *testing.B,
	v []T,
	sizeFunc func(protowire.Number, []T) int,
	marshalFunc func([]byte, protowire.Number, []T) int,
) {
	const fieldNum = protowire.MaxValidNumber
	buf := make([]byte, sizeFunc(fieldNum, v))

	for b.Loop() {
		marshalFunc(buf, fieldNum, v)
	}
}

func TestOptionalVarint(t *testing.T) {
	fieldNum := protowire.Number(1)
	zero := uint64(0)

	t.Run("nil", func(t *testing.T) {
		require.Zero(t, protoencoding.SizeOptionalVarint[uint64](fieldNum, nil))
		require.Zero(t, protoencoding.MarshalToOptionalVarint[uint64](nil, fieldNum, nil))
	})
	t.Run("zero", func(t *testing.T) {
		buf := make([]byte, protoencoding.SizeOptionalVarint(fieldNum, &zero))
		require.Equal(t, len(buf), protoencoding.MarshalToOptionalVarint(buf, fieldNum, &zero))
		require.Equal(t, []byte{8, 0}, buf)
	})
}

func TestVarint(t *testing.T) {
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[uint64], protoencoding.MarshalToVarint[uint64], protowire.ConsumeVarint, randVarint[uint64])
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[uint32], protoencoding.MarshalToVarint[uint32], consumeVarint[uint32], randVarint[uint32])
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[int64], protoencoding.MarshalToVarint[int64], consumeVarint[int64], randVarint[int64])
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[int32], protoencoding.MarshalToVarint[int32], consumeVarint[int32], randVarint[int32])
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[enum], protoencoding.MarshalToVarint[enum], consumeVarint[enum], randVarint[enum])
	testEncoding(t, protowire.VarintType, protoencoding.SizeVarint[int], protoencoding.MarshalToVarint[int], consumeVarint[int], randVarint[int])
}

func benchmarkMarshalVarint[T protoencoding.Varint](b *testing.B, v T) {
	b.Run(fmt.Sprintf("%T", v), func(b *testing.B) {
		benchmarkType(b, v, protoencoding.SizeVarint[T], protoencoding.MarshalToVarint[T])
	})
}

func BenchmarkMarshalVarint(b *testing.B) {
	v := uint64(math.MaxUint64)
	benchmarkMarshalVarint(b, v)
	benchmarkMarshalVarint(b, uint32(v))
	benchmarkMarshalVarint(b, int64(v))
	benchmarkMarshalVarint(b, int32(v))
}

func TestBool(t *testing.T) {
	fieldNum := randFieldNum()
	require.Zero(t, protoencoding.SizeBool(fieldNum, false))
	require.Zero(t, protoencoding.MarshalToBool(nil, fieldNum, false))

	sz := protoencoding.SizeBool(fieldNum, true)
	b := make([]byte, sz)
	require.EqualValues(t, sz, protoencoding.MarshalToBool(b, fieldNum, true))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, fieldNum, num)
	require.EqualValues(t, protowire.VarintType, typ)

	res, code := protowire.ConsumeVarint(b[tagLn:])
	require.NoError(t, protowire.ParseError(code))
	require.EqualValues(t, 1, res)
}

func BenchmarkMarshalBool(b *testing.B) {
	benchmarkType(b, true, protoencoding.SizeBool, protoencoding.MarshalToBool)
}

func TestFixed32(t *testing.T) {
	testEncoding(t, protowire.Fixed32Type, protoencoding.SizeFixed32, protoencoding.MarshalToFixed32, protowire.ConsumeFixed32, rand.Uint32)
}

func BenchmarkMarshalFixed32(b *testing.B) {
	benchmarkType(b, math.MaxUint32, protoencoding.SizeFixed32, protoencoding.MarshalToFixed32)
}

func TestFixed64(t *testing.T) {
	testEncoding(t, protowire.Fixed64Type, protoencoding.SizeFixed64, protoencoding.MarshalToFixed64, protowire.ConsumeFixed64, rand.Uint64)
}

func BenchmarkMarshalFixed64(b *testing.B) {
	benchmarkType(b, math.MaxUint64, protoencoding.SizeFixed64, protoencoding.MarshalToFixed64)
}

func TestFloat(t *testing.T) {
	testEncoding(t, protowire.Fixed32Type, protoencoding.SizeFloat, protoencoding.MarshalToFloat, func(b []byte) (float32, int) {
		v, ln := protowire.ConsumeFixed32(b)
		return math.Float32frombits(v), ln
	}, func() float32 {
		v := -math.MaxFloat32 + rand.Float32()*math.MaxFloat32*2
		require.False(t, math.IsNaN(float64(v)))
		return v
	})
}

func BenchmarkMarshalFloat(b *testing.B) {
	benchmarkType(b, math.Float32frombits(math.MaxUint32), protoencoding.SizeFloat, protoencoding.MarshalToFloat)
}

func TestDouble(t *testing.T) {
	testEncoding(t, protowire.Fixed64Type, protoencoding.SizeDouble, protoencoding.MarshalToDouble, func(b []byte) (float64, int) {
		v, ln := protowire.ConsumeFixed64(b)
		return math.Float64frombits(v), ln
	}, func() float64 {
		v := rand.NormFloat64()
		require.False(t, math.IsNaN(v))
		return v
	})
}

func BenchmarkDouble(b *testing.B) {
	benchmarkType(b, math.Float64frombits(math.MaxUint64), protoencoding.SizeDouble, protoencoding.MarshalToDouble)
}

func TestBytes(t *testing.T) {
	testEncoding(t, protowire.BytesType, protoencoding.SizeBytes[[]byte], protoencoding.MarshalToBytes[[]byte], consumeBytes[[]byte], randBytes[[]byte])
	testEncoding(t, protowire.BytesType, protoencoding.SizeBytes[string], protoencoding.MarshalToBytes[string], consumeBytes[string], randBytes[string])
}

func benchmarkMarshalBytes[T protoencoding.Bytes](b *testing.B, v T) {
	b.Run(fmt.Sprintf("%T", v), func(b *testing.B) {
		benchmarkType(b, v, protoencoding.SizeBytes[T], protoencoding.MarshalToBytes[T])
	})
}

func BenchmarkMarshalBytes(b *testing.B) {
	const bs = "Hello, world!"
	benchmarkMarshalBytes(b, bs)
	benchmarkMarshalBytes(b, []byte(bs))
}

func TestEmbedded(t *testing.T) {
	fieldNum := randFieldNum()
	t.Run("nil", func(t *testing.T) {
		require.Zero(t, protoencoding.SizeEmbedded(fieldNum, nil))
		require.Zero(t, protoencoding.MarshalToEmbedded(nil, fieldNum, nil))
		require.Zero(t, protoencoding.SizeEmbedded(fieldNum, (*timestamp)(nil)))
		require.Zero(t, protoencoding.MarshalToEmbedded(nil, fieldNum, (*timestamp)(nil)))
	})
	t.Run("zero", func(t *testing.T) {
		const fieldNum = 480010005
		const expLen = 6
		tz := new(timestamp)
		require.EqualValues(t, expLen, protoencoding.SizeEmbedded(fieldNum, tz))
		b := make([]byte, expLen)
		protoencoding.MarshalToEmbedded(b, fieldNum, tz)
		require.Equal(t, []byte{170, 241, 139, 167, 14, 0}, b) // first 5 bytes is num, last zero is size
	})

	v := (*timestamp)(timestamppb.Now())
	sz := protoencoding.SizeEmbedded(fieldNum, v)
	b := make([]byte, sz)
	require.EqualValues(t, sz, protoencoding.MarshalToEmbedded(b, fieldNum, v))

	num, typ, tagLn := protowire.ConsumeTag(b)
	require.Positive(t, tagLn, protowire.ParseError(tagLn))
	require.EqualValues(t, fieldNum, num)
	require.EqualValues(t, protowire.BytesType, typ)

	res, code := protowire.ConsumeBytes(b[tagLn:])
	require.NoError(t, protowire.ParseError(code))
	var v2 timestamp
	require.NoError(t, stdproto.Unmarshal(res, (*timestamppb.Timestamp)(&v2)))
	require.Equal(t, v.Seconds, v2.Seconds)
	require.Equal(t, v.Nanos, v2.Nanos)
}

func BenchmarkMarshalEmbedded(b *testing.B) {
	const fieldNum = protowire.MaxValidNumber
	v := (*timestamp)(timestamppb.Now())
	buf := make([]byte, protoencoding.SizeEmbedded(fieldNum, v))

	for b.Loop() {
		protoencoding.MarshalToEmbedded(buf, fieldNum, v)
	}
}

func TestRepeatedVarint(t *testing.T) {
	testEncoding(t, protowire.BytesType, protoencoding.SizeRepeatedVarint[uint64], protoencoding.MarshalToRepeatedVarint[uint64], consumeRepeatedVarint[uint64], randRepeatedVarint[uint64])
	testEncoding(t, protowire.BytesType, protoencoding.SizeRepeatedVarint[uint32], protoencoding.MarshalToRepeatedVarint[uint32], consumeRepeatedVarint[uint32], randRepeatedVarint[uint32])
	testEncoding(t, protowire.BytesType, protoencoding.SizeRepeatedVarint[int64], protoencoding.MarshalToRepeatedVarint[int64], consumeRepeatedVarint[int64], randRepeatedVarint[int64])
	testEncoding(t, protowire.BytesType, protoencoding.SizeRepeatedVarint[int32], protoencoding.MarshalToRepeatedVarint[int32], consumeRepeatedVarint[int32], randRepeatedVarint[int32])
	testEncoding(t, protowire.BytesType, protoencoding.SizeRepeatedVarint[enum], protoencoding.MarshalToRepeatedVarint[enum], consumeRepeatedVarint[enum], randRepeatedVarint[enum])
}

func benchmarkMarshalRepeatedVarint[T protoencoding.Varint](b *testing.B, v T) {
	b.Run(fmt.Sprintf("%T", v), func(b *testing.B) {
		benchmarkRepeatedType(b, []T{v, v, v}, protoencoding.SizeRepeatedVarint[T], protoencoding.MarshalToRepeatedVarint[T])
	})
}

func BenchmarkMarshalRepeatedVarint(b *testing.B) {
	v := uint64(math.MaxUint64)
	benchmarkMarshalRepeatedVarint(b, v)
	benchmarkMarshalRepeatedVarint(b, uint32(v))
	benchmarkMarshalRepeatedVarint(b, int64(v))
	benchmarkMarshalRepeatedVarint(b, int32(v))
	benchmarkMarshalRepeatedVarint(b, enum(v))
}

func TestRepeatedBytes(t *testing.T) {
	testPackedRepeated(t, protoencoding.SizeRepeatedBytes[[]byte], protoencoding.MarshalToRepeatedBytes[[]byte], consumePackedRepeated[[]byte], randRepeatedBytes[[]byte])
	testPackedRepeated(t, protoencoding.SizeRepeatedBytes[string], protoencoding.MarshalToRepeatedBytes[string], consumePackedRepeated[string], randRepeatedBytes[string])
}

func benchmarkMarshalRepeatedBytes[T protoencoding.Bytes](b *testing.B, v T) {
	b.Run(fmt.Sprintf("%T", v), func(b *testing.B) {
		benchmarkRepeatedType(b, []T{v, v, v}, protoencoding.SizeRepeatedBytes[T], protoencoding.MarshalToRepeatedBytes[T])
	})
}

func BenchmarkMarshalRepeatedBytes(b *testing.B) {
	const bs = "Hello, world!"
	benchmarkMarshalRepeatedBytes(b, bs)
	benchmarkMarshalRepeatedBytes(b, []byte(bs))
}

func TestSizeEmbeddedLENField(t *testing.T) {
	t.Run("negative length", func(t *testing.T) {
		require.PanicsWithValue(t, "negative length -1", func() {
			protoencoding.SizeEmbeddedLENField(1, -1)
		})
	})

	t.Run("invalid field number", func(t *testing.T) {
		for _, num := range []protowire.Number{-1, 0, 536870912} {
			t.Run(strconv.Itoa(int(num)), func(t *testing.T) {
				msg := fmt.Sprintf("invalid field number %d", num)
				require.PanicsWithValue(t, msg, func() {
					protoencoding.SizeEmbeddedLENField(num, 0)
				})
			})
		}
	})

	t.Run("zero length", func(t *testing.T) {
		ln := protoencoding.SizeEmbeddedLENField(1, 0)
		require.Zero(t, ln)
	})

	for i, tc := range []struct {
		num protowire.Number
		ln  int
		exp int
	}{
		{num: 1, ln: 1, exp: 1 + 1 + 1},
		{num: 128, ln: 1, exp: 2 + 1 + 1},
		{num: 1, ln: 128, exp: 1 + 2 + 128},
		{num: 128, ln: 128, exp: 2 + 2 + 128},
	} {
		ln := protoencoding.SizeEmbeddedLENField(tc.num, tc.ln)
		require.EqualValues(t, tc.exp, ln, i)
	}
}
