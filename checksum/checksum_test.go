package checksum

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/stretchr/testify/require"
)

func randSHA256(t *testing.T) [sha256.Size]byte {
	cSHA256 := [sha256.Size]byte{}
	_, err := rand.Read(cSHA256[:])
	require.NoError(t, err)

	return cSHA256
}

func TestChecksum(t *testing.T) {
	c := New()

	cSHA256 := [sha256.Size]byte{}
	_, _ = rand.Read(cSHA256[:])

	c.SetSHA256(cSHA256)

	require.Equal(t, SHA256, c.Type())
	require.Equal(t, cSHA256[:], c.Sum())

	cV2 := c.ToV2()

	require.Equal(t, refs.SHA256, cV2.GetType())
	require.Equal(t, cSHA256[:], cV2.GetSum())

	cTZ := [64]byte{}
	_, _ = rand.Read(cSHA256[:])

	c.SetTillichZemor(cTZ)

	require.Equal(t, TZ, c.Type())
	require.Equal(t, cTZ[:], c.Sum())

	cV2 = c.ToV2()

	require.Equal(t, refs.TillichZemor, cV2.GetType())
	require.Equal(t, cTZ[:], cV2.GetSum())
}

func TestEqualChecksums(t *testing.T) {
	require.True(t, Equal(nil, nil))

	csSHA := [sha256.Size]byte{}
	_, _ = rand.Read(csSHA[:])

	cs1 := New()
	cs1.SetSHA256(csSHA)

	cs2 := New()
	cs2.SetSHA256(csSHA)

	require.True(t, Equal(cs1, cs2))

	csSHA[0]++
	cs2.SetSHA256(csSHA)

	require.False(t, Equal(cs1, cs2))
}

func TestChecksumEncoding(t *testing.T) {
	cs := New()
	cs.SetSHA256(randSHA256(t))

	t.Run("binary", func(t *testing.T) {
		data, err := cs.Marshal()
		require.NoError(t, err)

		c2 := New()
		require.NoError(t, c2.Unmarshal(data))

		require.Equal(t, cs, c2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := cs.MarshalJSON()
		require.NoError(t, err)

		cs2 := New()
		require.NoError(t, cs2.UnmarshalJSON(data))

		require.Equal(t, cs, cs2)
	})

	t.Run("string", func(t *testing.T) {
		cs2 := New()

		require.NoError(t, cs2.Parse(cs.String()))

		require.Equal(t, cs, cs2)
	})
}

func TestNewChecksumFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *refs.Checksum

		require.Nil(t, NewFromV2(x))
	})
}

func TestChecksum_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *Checksum

		require.Nil(t, x.ToV2())
	})
}

func TestNewChecksum(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		chs := New()

		// check initial values
		require.Equal(t, Unknown, chs.Type())
		require.Nil(t, chs.Sum())

		// convert to v2 message
		chsV2 := chs.ToV2()

		require.Equal(t, refs.UnknownChecksum, chsV2.GetType())
		require.Nil(t, chsV2.GetSum())
	})
}

type enumIface interface {
	FromString(string) bool
	String() string
}

type enumStringItem struct {
	val enumIface
	str string
}

func testEnumStrings(t *testing.T, e enumIface, items []enumStringItem) {
	for _, item := range items {
		require.Equal(t, item.str, item.val.String())

		s := item.val.String()

		require.True(t, e.FromString(s), s)

		require.EqualValues(t, item.val, e, item.val)
	}

	// incorrect strings
	for _, str := range []string{
		"some string",
		"undefined",
	} {
		require.False(t, e.FromString(str))
	}
}

func TestChecksumType_String(t *testing.T) {
	toPtr := func(v Type) *Type {
		return &v
	}

	testEnumStrings(t, new(Type), []enumStringItem{
		{val: toPtr(TZ), str: "TZ"},
		{val: toPtr(SHA256), str: "SHA256"},
		{val: toPtr(Unknown), str: "CHECKSUM_TYPE_UNSPECIFIED"},
	})
}
