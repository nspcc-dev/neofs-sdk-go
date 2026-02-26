package prototest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	stdproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestMarshalStable tests that all given proto.Message instances encode into
// Protocol Buffers V3 correctly. TestMarshalStable also checks that
// [proto.Message.MarshalStable] panics if buffer length is less than
// [proto.Message.MarshaledSize]. Nil and zeroed PTR cases are also tested so no
// need to add them to xs. The xs may be left empty if message has no fields.
func TestMarshalStable[T any, PTR interface {
	*T
	proto.Message
	stdproto.Message
}](t testing.TB, xs []PTR) {
	xs = append(xs, nil, new(T))

	repeatedFields := collectRepeatedFields(PTR(new(T)))

	for i, x := range xs {
		sz := x.MarshaledSize()
		if x == nil {
			require.Zero(t, sz, i)
			require.NotPanics(t, func() { x.MarshalStable(nil) }, i)
			continue
		}
		if sz > 0 {
			require.Panics(t, func() { x.MarshalStable(make([]byte, sz-1)) }, i)
		}

		bx := make([]byte, sz)
		x.MarshalStable(bx)

		// assert ascending field order
		var off int
		prevNum := protowire.Number(-1)
		for len(bx[off:]) > 0 {
			num, _, ln := protowire.ConsumeField(bx[off:])
			require.NoError(t, protowire.ParseError(ln), i)
			require.True(t, num >= 0, i)
			if prevNum >= 0 {
				if num == prevNum {
					require.Contains(t, repeatedFields, num, i)
				} else {
					require.Greater(t, num, prevNum, i)
				}
			}
			prevNum = num
			off += ln
		}

		var y PTR = new(T)
		err := stdproto.Unmarshal(bx, y)
		require.NoError(t, err, i)
		require.Empty(t, y.ProtoReflect().GetUnknown(), i)

		by := make([]byte, y.MarshaledSize())
		y.MarshalStable(by)
		require.Equal(t, bx, by)
		require.True(t, equalProtoMessages(x, y), cmp.Diff(x, y, protocmp.Transform()))
	}
}

// replaces [stdproto.Equal] which considers nil and zero embedded messages as
// different. In NeoFS, such messages are the same.
func equalProtoMessages(x, y stdproto.Message) bool {
	if stdproto.Equal(x, y) || cmp.Equal(x, y, protocmp.Transform(), protocmp.IgnoreEmptyMessages()) {
		return true
	}

	rx, ry := x.ProtoReflect(), y.ProtoReflect()
	if rx.Descriptor() != ry.Descriptor() {
		return false
	} else if len(rx.GetUnknown()) > 0 || len(ry.GetUnknown()) > 0 {
		return false
	}
	nx := 0
	equal := true
	rx.Range(func(fd protoreflect.FieldDescriptor, vx protoreflect.Value) bool {
		if fd.Kind() == protoreflect.MessageKind {
			vxm := vx.Message()
			if ry.Has(fd) {
				nx++
				equal = equalProtoMessages(vxm.Interface(), ry.Get(fd).Message().Interface())
			} else {
				equal = equalProtoMessages(vxm.Interface(), vxm.New().Interface())
			}
		} else {
			nx++
			equal = ry.Has(fd) && vx.Equal(ry.Get(fd))
		}
		return equal
	})
	if !equal {
		return false
	}
	ny := 0
	equal = true
	ry.Range(func(fd protoreflect.FieldDescriptor, vy protoreflect.Value) bool {
		if fd.Kind() == protoreflect.MessageKind && !rx.Has(fd) {
			vym := vy.Message()
			equal = equalProtoMessages(vym.Interface(), vym.New().Interface())
			return equal
		}
		ny++
		return true
	})
	return equal && nx == ny
}

func collectRepeatedFields(m stdproto.Message) []protowire.Number {
	var res []protowire.Number

	flds := m.ProtoReflect().Descriptor().Fields()
	for i := range flds.Len() {
		if fld := flds.Get(i); fld.Cardinality() == protoreflect.Repeated {
			res = append(res, fld.Number())
		}
	}

	return res
}
