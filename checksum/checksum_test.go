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
	var c Checksum

	cSHA256 := [sha256.Size]byte{}
	_, _ = rand.Read(cSHA256[:])

	c.SetSHA256(cSHA256)

	require.Equal(t, SHA256, c.Type())
	require.Equal(t, cSHA256[:], c.Sum())

	var cV2 refs.Checksum
	c.WriteToV2(&cV2)

	require.Equal(t, refs.SHA256, cV2.GetType())
	require.Equal(t, cSHA256[:], cV2.GetSum())

	cTZ := [64]byte{}
	_, _ = rand.Read(cSHA256[:])

	c.SetTillichZemor(cTZ)

	require.Equal(t, TZ, c.Type())
	require.Equal(t, cTZ[:], c.Sum())

	c.WriteToV2(&cV2)

	require.Equal(t, refs.TillichZemor, cV2.GetType())
	require.Equal(t, cTZ[:], cV2.GetSum())
}

func TestEqualChecksums(t *testing.T) {
	require.True(t, Equal(&Checksum{}, &Checksum{}))

	csSHA := [sha256.Size]byte{}
	_, _ = rand.Read(csSHA[:])

	var cs1 Checksum
	cs1.SetSHA256(csSHA)

	var cs2 Checksum
	cs2.SetSHA256(csSHA)

	require.True(t, Equal(&cs1, &cs2))

	csSHA[0]++
	cs2.SetSHA256(csSHA)

	require.False(t, Equal(&cs1, &cs2))
}

func TestChecksumEncoding(t *testing.T) {
	var cs Checksum
	cs.SetSHA256(randSHA256(t))

	t.Run("string", func(t *testing.T) {
		var cs2 Checksum

		require.NoError(t, cs2.Parse(cs.String()))

		require.Equal(t, cs, cs2)
	})
}

func TestNewChecksum(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var chs Checksum

		// check initial values
		require.Equal(t, Unknown, chs.Type())
		require.Nil(t, chs.Sum())

		// convert to v2 message
		var chsV2 refs.Checksum
		chs.WriteToV2(&chsV2)

		require.Equal(t, refs.UnknownChecksum, chsV2.GetType())
		require.Nil(t, chsV2.GetSum())
	})
}

type enumIface interface {
	Parse(string) bool
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

		require.True(t, e.Parse(s), s)

		require.EqualValues(t, item.val, e, item.val)
	}

	// incorrect strings
	for _, str := range []string{
		"some string",
		"undefined",
	} {
		require.False(t, e.Parse(str))
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
