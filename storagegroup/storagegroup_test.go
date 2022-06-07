package storagegroup_test

import (
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	storagegroupV2 "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
	storagegroupV2test "github.com/nspcc-dev/neofs-api-go/v2/storagegroup/test"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	storagegrouptest "github.com/nspcc-dev/neofs-sdk-go/storagegroup/test"
	"github.com/stretchr/testify/require"
)

func TestStorageGroup(t *testing.T) {
	var sg storagegroup.StorageGroup

	sz := uint64(13)
	sg.SetValidationDataSize(sz)
	require.Equal(t, sz, sg.ValidationDataSize())

	cs := checksumtest.Checksum()
	sg.SetValidationDataHash(cs)
	cs2, set := sg.ValidationDataHash()

	require.True(t, set)
	require.Equal(t, cs, cs2)

	exp := uint64(33)
	sg.SetExpirationEpoch(exp)
	require.Equal(t, exp, sg.ExpirationEpoch())

	members := []oid.ID{oidtest.ID(), oidtest.ID()}
	sg.SetMembers(members)
	require.Equal(t, members, sg.Members())
}

func TestStorageGroup_ReadFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			x  storagegroup.StorageGroup
			v2 storagegroupV2.StorageGroup
		)

		x.ReadFromV2(v2)

		require.Zero(t, x.ExpirationEpoch())
		require.Zero(t, x.ValidationDataSize())
		_, set := x.ValidationDataHash()
		require.False(t, set)
		require.Zero(t, x.Members())
	})

	t.Run("from non-zero", func(t *testing.T) {
		var (
			x  storagegroup.StorageGroup
			v2 = storagegroupV2test.GenerateStorageGroup(false)
		)

		// https://github.com/nspcc-dev/neofs-api-go/issues/394
		v2.SetMembers(generateOIDList())

		size := v2.GetValidationDataSize()
		epoch := v2.GetExpirationEpoch()
		mm := v2.GetMembers()
		hashV2 := v2.GetValidationHash()

		x.ReadFromV2(*v2)

		require.Equal(t, epoch, x.ExpirationEpoch())
		require.Equal(t, size, x.ValidationDataSize())

		var hash checksum.Checksum
		hash.ReadFromV2(*hashV2)
		h, set := x.ValidationDataHash()
		require.True(t, set)
		require.Equal(t, hash, h)

		var oidV2 refs.ObjectID

		for i, m := range mm {
			x.Members()[i].WriteToV2(&oidV2)
			require.Equal(t, m, oidV2)
		}
	})
}

func TestStorageGroupEncoding(t *testing.T) {
	sg := storagegrouptest.StorageGroup()

	t.Run("binary", func(t *testing.T) {
		data, err := sg.Marshal()
		require.NoError(t, err)

		var sg2 storagegroup.StorageGroup
		require.NoError(t, sg2.Unmarshal(data))

		require.Equal(t, sg, sg2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := sg.MarshalJSON()
		require.NoError(t, err)

		var sg2 storagegroup.StorageGroup
		require.NoError(t, sg2.UnmarshalJSON(data))

		require.Equal(t, sg, sg2)
	})
}

func TestStorageGroup_WriteToV2(t *testing.T) {
	t.Run("zero to v2", func(t *testing.T) {
		var (
			x  storagegroup.StorageGroup
			v2 storagegroupV2.StorageGroup
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetValidationHash())
		require.Nil(t, v2.GetMembers())
		require.Zero(t, v2.GetValidationDataSize())
		require.Zero(t, v2.GetExpirationEpoch())
	})

	t.Run("non-zero to v2", func(t *testing.T) {
		var (
			x  = storagegrouptest.StorageGroup()
			v2 storagegroupV2.StorageGroup
		)

		x.WriteToV2(&v2)

		require.Equal(t, x.ExpirationEpoch(), v2.GetExpirationEpoch())
		require.Equal(t, x.ValidationDataSize(), v2.GetValidationDataSize())

		var hash checksum.Checksum
		hash.ReadFromV2(*v2.GetValidationHash())

		h, set := x.ValidationDataHash()
		require.True(t, set)
		require.Equal(t, h, hash)

		var oidV2 refs.ObjectID

		for i, m := range x.Members() {
			m.WriteToV2(&oidV2)
			require.Equal(t, oidV2, v2.GetMembers()[i])
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var sg storagegroup.StorageGroup

		// check initial values
		require.Nil(t, sg.Members())
		_, set := sg.ValidationDataHash()
		require.False(t, set)
		require.Zero(t, sg.ExpirationEpoch())
		require.Zero(t, sg.ValidationDataSize())
	})
}

func generateOIDList() []refs.ObjectID {
	const size = 3

	mmV2 := make([]refs.ObjectID, size)
	for i := 0; i < size; i++ {
		oidV2 := make([]byte, sha256.Size)
		oidV2[i] = byte(i)

		mmV2[i].SetValue(oidV2)
	}

	return mmV2
}

func TestStorageGroup_SetMembers_DoubleSetting(t *testing.T) {
	var sg storagegroup.StorageGroup

	mm := []oid.ID{oidtest.ID(), oidtest.ID(), oidtest.ID()} // cap is 3 at least
	sg.SetMembers(mm)

	// the previous cap is more that a new length;
	// slicing should not lead to `out of range`
	// and apply update correctly
	sg.SetMembers(mm[:1])
}
