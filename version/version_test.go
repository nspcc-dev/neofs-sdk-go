package version_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func TestVersionComparable(t *testing.T) {
	v1 := versiontest.Version()
	require.True(t, v1 == v1)
	v2 := versiontest.Version()
	require.NotEqual(t, v1, v2)
	require.False(t, v1 == v2)
}

func TestCurrent(t *testing.T) {
	require.EqualValues(t, 2, version.Current.Major())
	require.EqualValues(t, 16, version.Current.Minor())
}

func testVersionField(t *testing.T, get func(version.Version) uint32, set func(*version.Version, uint32), getAPI func(*refs.Version) uint32) {
	var v version.Version

	require.Zero(t, get(v))

	val := rand.Uint32()
	set(&v, val)
	require.EqualValues(t, val, get(v))

	otherVal := val + 1
	set(&v, otherVal)
	require.EqualValues(t, otherVal, get(v))

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst version.Version
			var msg refs.Version

			set(&dst, val)

			src.WriteToV2(&msg)
			require.Zero(t, getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)
			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(&msg))
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
	})
}

func TestVersion_SetMajor(t *testing.T) {
	testVersionField(t, version.Version.Major, (*version.Version).SetMajor, (*refs.Version).GetMajor)
}

func TestVersion_SetMinor(t *testing.T) {
	testVersionField(t, version.Version.Minor, (*version.Version).SetMinor, (*refs.Version).GetMinor)
}

func TestEncodeToString(t *testing.T) {
	require.Equal(t, "v2.16", version.EncodeToString(version.Current))
	var v version.Version
	v.SetMajor(578393)
	v.SetMinor(405609340)
	require.Equal(t, "v578393.405609340", version.EncodeToString(v))
}
