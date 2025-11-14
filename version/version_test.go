package version_test

import (
	"math/rand/v2"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

var validVersion = version.New(96984596, 1910541418)

// corresponds to validVersion.
const validJSON = `{"major":96984596,"minor":1910541418}`

func testVersionField(
	t *testing.T,
	get func(version.Version) uint32,
	set func(*version.Version, uint32),
	getAPI func(*refs.Version) uint32,
) {
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

			set(&dst, val)
			msg := src.ProtoMessage()
			require.Zero(t, getAPI(msg))
			require.NoError(t, dst.FromProtoMessage(msg))
			require.Zero(t, get(dst))

			set(&src, val)
			msg = src.ProtoMessage()
			require.EqualValues(t, val, getAPI(msg))
			err := dst.FromProtoMessage(msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst version.Version

			set(&dst, val)
			b, err := src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(b))
			require.Zero(t, get(dst))

			set(&src, val)
			b, err = src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(b))
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

func TestVersionComparable(t *testing.T) {
	x := versiontest.Version()
	y := x
	require.True(t, x == y)
	require.False(t, x != y)
	y = versiontest.Version()
	require.False(t, x == y)
	require.True(t, x != y)
}

func TestVersion_String(t *testing.T) {
	v := versiontest.Version()
	require.NotEmpty(t, v.String())
	require.Equal(t, v.String(), v.String())
	require.NotEqual(t, v.String(), versiontest.Version().String())

	var v2 version.Version
	require.NoError(t, v2.DecodeString(v.String()))
	require.Equal(t, v, v2)
}

func TestVersion_MarshalJSON(t *testing.T) {
	b, err := validVersion.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, validJSON, string(b))
}

func TestVersion_UnmarshalJSON(t *testing.T) {
	var v1 version.Version
	require.NoError(t, v1.UnmarshalJSON([]byte(validJSON)))
	require.Equal(t, validVersion, v1)
	v2, err := version.UnmarshalJSON([]byte(validJSON))
	require.NoError(t, err)
	require.Equal(t, validVersion, v2)

	t.Run("invalid JSON", func(t *testing.T) {
		b := []byte("Hello, world!")
		err := new(version.Version).UnmarshalJSON(b)
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
		_, err = version.UnmarshalJSON(b)
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})
}

func TestEncodeToString(t *testing.T) {
	require.Equal(t, "v2.20", version.EncodeToString(version.Current()))
	require.Equal(t, "v96984596.1910541418", version.EncodeToString(validVersion))
}

func TestNew(t *testing.T) {
	mjr, mnr := rand.Uint32(), rand.Uint32()
	v := version.New(mjr, mnr)
	require.Equal(t, mjr, v.Major())
	require.Equal(t, mnr, v.Minor())
}

func TestCurrent(t *testing.T) {
	v := version.Current()

	require.EqualValues(t, 2, v.Major())
	require.EqualValues(t, 20, v.Minor())
}

func TestDecodeString(t *testing.T) {
	var v version.Version

	require.Error(t, v.DecodeString(""))
	require.Error(t, v.DecodeString("v"))
	require.Error(t, v.DecodeString("v."))
	require.Error(t, v.DecodeString("v0."))
	require.Error(t, v.DecodeString("v0"))
	require.Error(t, v.DecodeString("v100"))
	require.Error(t, v.DecodeString("v10023"))
	require.Error(t, v.DecodeString("v1.1.1"))
	require.Error(t, v.DecodeString("v1..0"))
	require.Error(t, v.DecodeString("v1.01"))
	require.Error(t, v.DecodeString("v01.1"))
	require.Error(t, v.DecodeString("va.1"))
	require.Error(t, v.DecodeString("v1.a"))
	require.Error(t, v.DecodeString("v-1.0"))
	require.Error(t, v.DecodeString("v1.-1"))

	require.NoError(t, v.DecodeString("v2.17"))
	require.Equal(t, uint32(2), v.Major())
	require.Equal(t, uint32(17), v.Minor())

	require.NoError(t, v.DecodeString("v0.0"))
	require.Equal(t, uint32(0), v.Major())
	require.Equal(t, uint32(0), v.Minor())
}
