package storagegroup_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apistoragegroup "github.com/nspcc-dev/neofs-sdk-go/api/storagegroup"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStorageGroupDecoding(t *testing.T) {
	id := oidtest.ID()
	fmt.Println(id[:])
	fmt.Println(id)
	t.Run("invalid binary", func(t *testing.T) {
		var sg storagegroup.StorageGroup
		msg := []byte("definitely_not_protobuf")
		err := sg.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(*apistoragegroup.StorageGroup)
		}{
			{name: "checksum/value/nil", err: "invalid hash: missing value", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.ValidationHash = new(refs.Checksum)
			}},
			{name: "checksum/value/empty", err: "invalid hash: missing value", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.ValidationHash = &refs.Checksum{Sum: []byte{}}
			}},
			{name: "members/nil", err: "invalid member #1: missing value field", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.Members = []*refs.ObjectID{
					{Value: make([]byte, 32)},
					nil,
				}
			}},
			{name: "members/value/nil", err: "invalid member #1: missing value field", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.Members = []*refs.ObjectID{
					{Value: make([]byte, 32)},
					{Value: nil},
				}
			}},
			{name: "members/value/empty", err: "invalid member #1: missing value field", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.Members = []*refs.ObjectID{
					{Value: make([]byte, 32)},
					{Value: []byte{}},
				}
			}},
			{name: "members/value/wrong length", err: "invalid member #1: invalid value length 31", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.Members = []*refs.ObjectID{
					{Value: make([]byte, 32)},
					{Value: make([]byte, 31)},
				}
			}},
			{name: "members/duplicated", err: "duplicated member EMwQfxfrUrLwnYxZDRBCe8NwyThefNPni7s1QLQXWay7", corrupt: func(sg *apistoragegroup.StorageGroup) {
				sg.Members = []*refs.ObjectID{
					{Value: []byte{198, 133, 16, 209, 121, 137, 128, 158, 158, 74, 248, 227, 131, 233, 166, 249, 7, 111, 24, 55, 189, 32, 76, 140, 146, 7, 123, 228, 49, 198, 58, 98}},
					{Value: make([]byte, 32)},
					{Value: []byte{198, 133, 16, 209, 121, 137, 128, 158, 158, 74, 248, 227, 131, 233, 166, 249, 7, 111, 24, 55, 189, 32, 76, 140, 146, 7, 123, 228, 49, 198, 58, 98}},
				}
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst storagegroup.StorageGroup
				var m apistoragegroup.StorageGroup

				require.NoError(t, proto.Unmarshal(src.Marshal(), &m))
				testCase.corrupt(&m)
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.Unmarshal(b), testCase.err)
			})
		}
	})
}

func TestStorageGroup_SetValidationDataSize(t *testing.T) {
	var sg storagegroup.StorageGroup

	require.Zero(t, sg.ValidationDataSize())

	val := rand.Uint64()
	sg.SetValidationDataSize(val)
	require.EqualValues(t, val, sg.ValidationDataSize())

	otherVal := val + 1
	sg.SetValidationDataSize(otherVal)
	require.EqualValues(t, otherVal, sg.ValidationDataSize())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst storagegroup.StorageGroup

			dst.SetValidationDataSize(val)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Zero(t, dst.ValidationDataSize())

			src.SetValidationDataSize(val)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.EqualValues(t, val, dst.ValidationDataSize())
		})
	})
}

func TestStorageGroup_SetValidationDataHash(t *testing.T) {
	var sg storagegroup.StorageGroup

	require.Zero(t, sg.ValidationDataHash())

	cs := checksumtest.Checksum()
	sg.SetValidationDataHash(cs)
	require.Equal(t, cs, sg.ValidationDataHash())

	csOther := checksumtest.Checksum()
	sg.SetValidationDataHash(csOther)
	require.Equal(t, csOther, sg.ValidationDataHash())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst storagegroup.StorageGroup

			dst.SetValidationDataHash(cs)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Zero(t, dst.ValidationDataHash())

			src.SetValidationDataHash(cs)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Equal(t, cs, dst.ValidationDataHash())
		})
	})
}

func TestStorageGroup_SetMembers(t *testing.T) {
	var sg storagegroup.StorageGroup

	require.Zero(t, sg.Members())

	members := oidtest.NIDs(3)
	sg.SetMembers(members)
	require.Equal(t, members, sg.Members())

	otherMembers := oidtest.NIDs(2)
	sg.SetMembers(otherMembers)
	require.Equal(t, otherMembers, sg.Members())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst storagegroup.StorageGroup

			dst.SetMembers(members)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Zero(t, dst.Members())

			src.SetMembers(members)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Equal(t, members, dst.Members())
		})
	})
}

// TODO:
// func TestStorageGroupFromObject(t *testing.T) {
// 	sg := storagegrouptest.StorageGroup()
//
// 	var o objectSDK.Object
//
// 	var expAttr objectSDK.Attribute
// 	expAttr.SetKey(objectV2.SysAttributeExpEpoch)
// 	expAttr.SetValue(strconv.FormatUint(sg.ExpirationEpoch(), 10))
//
// 	sgRaw, err := sg.Marshal()
// 	require.NoError(t, err)
//
// 	o.SetPayload(sgRaw)
// 	o.SetType(objectSDK.TypeStorageGroup)
//
// 	t.Run("correct object", func(t *testing.T) {
// 		o.SetAttributes(objectSDK.Attribute{}, expAttr, objectSDK.Attribute{})
//
// 		var sg2 storagegroup.StorageGroup
// 		require.NoError(t, storagegroup.ReadFromObject(&sg2, o))
// 		require.Equal(t, sg, sg2)
// 	})
//
// 	t.Run("incorrect exp attr", func(t *testing.T) {
// 		var sg2 storagegroup.StorageGroup
//
// 		expAttr.SetValue(strconv.FormatUint(sg.ExpirationEpoch()+1, 10))
// 		o.SetAttributes(expAttr)
//
// 		require.Error(t, storagegroup.ReadFromObject(&sg2, o))
// 	})
//
// 	t.Run("incorrect object type", func(t *testing.T) {
// 		var sg2 storagegroup.StorageGroup
//
// 		o.SetType(objectSDK.TypeTombstone)
// 		require.Error(t, storagegroup.ReadFromObject(&sg2, o))
// 	})
// }
//
// func TestStorageGroupToObject(t *testing.T) {
// 	sg := storagegrouptest.StorageGroup()
//
// 	sgRaw, err := sg.Marshal()
// 	require.NoError(t, err)
//
// 	t.Run("empty object", func(t *testing.T) {
// 		var o objectSDK.Object
// 		storagegroup.WriteToObject(sg, &o)
//
// 		exp, found := expFromObj(t, o)
// 		require.True(t, found)
//
// 		require.Equal(t, sgRaw, o.Payload())
// 		require.Equal(t, sg.ExpirationEpoch(), exp)
// 		require.Equal(t, objectSDK.TypeStorageGroup, o.Type())
// 	})
//
// 	t.Run("obj already has exp attr", func(t *testing.T) {
// 		var o objectSDK.Object
//
// 		var attr objectSDK.Attribute
// 		attr.SetKey(objectV2.SysAttributeExpEpoch)
// 		attr.SetValue(strconv.FormatUint(sg.ExpirationEpoch()+1, 10))
//
// 		o.SetAttributes(objectSDK.Attribute{}, attr, objectSDK.Attribute{})
//
// 		storagegroup.WriteToObject(sg, &o)
//
// 		exp, found := expFromObj(t, o)
// 		require.True(t, found)
//
// 		require.Equal(t, sgRaw, o.Payload())
// 		require.Equal(t, sg.ExpirationEpoch(), exp)
// 		require.Equal(t, objectSDK.TypeStorageGroup, o.Type())
// 	})
// }
//
// func expFromObj(t *testing.T, o objectSDK.Object) (uint64, bool) {
// 	for _, attr := range o.Attributes() {
// 		if attr.Key() == objectV2.SysAttributeExpEpoch {
// 			exp, err := strconv.ParseUint(attr.Value(), 10, 64)
// 			require.NoError(t, err)
//
// 			return exp, true
// 		}
// 	}
//
// 	return 0, false
// }
