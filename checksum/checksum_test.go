package checksum

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

func TestChecksum(t *testing.T) {
	data := make([]byte, 32)
	//nolint:staticcheck
	rand.Read(data)

	cSHA256 := sha256.Sum256(data)
	c := NewSHA256(cSHA256)

	require.Equal(t, SHA256, c.Type())
	require.Equal(t, cSHA256[:], c.Value())

	var cV2 refs.Checksum
	c.WriteToV2(&cV2)

	require.Equal(t, refs.SHA256, cV2.GetType())
	require.Equal(t, cSHA256[:], cV2.GetSum())

	cTZ := tz.Sum(data)
	c = NewTillichZemor(cTZ)

	require.Equal(t, TZ, c.Type())
	require.Equal(t, cTZ[:], c.Value())

	c.WriteToV2(&cV2)

	require.Equal(t, refs.TillichZemor, cV2.GetType())
	require.Equal(t, cTZ[:], cV2.GetSum())
}

func TestNewChecksum(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var chs Checksum

		// check initial values
		require.Zero(t, chs.Type())
		require.Nil(t, chs.Value())

		// convert to v2 message
		var chsV2 refs.Checksum
		chs.WriteToV2(&chsV2)

		require.Equal(t, refs.UnknownChecksum, chsV2.GetType())
		require.Nil(t, chsV2.GetSum())
	})
}

func TestNewFromData(t *testing.T) {
	_, err := NewFromData(0, nil)
	require.EqualError(t, err, "unsupported checksum type 0")
	var cs Checksum
	Calculate(&cs, 0, nil)
	require.Zero(t, cs)
	_, err = NewFromData(TZ+1, nil)
	require.EqualError(t, err, "unsupported checksum type 3")
	Calculate(&cs, TZ+1, nil)
	require.Zero(t, cs)

	payload := []byte{0, 1, 2, 3, 4, 5}

	t.Run("SHA256", func(t *testing.T) {
		orig := sha256.Sum256(payload)

		c, err := NewFromData(SHA256, payload)
		require.NoError(t, err)
		require.Equal(t, orig[:], c.Value())
		require.Equal(t, SHA256, c.Type())

		c = Checksum{}
		Calculate(&c, SHA256, payload)
		require.Equal(t, orig[:], c.Value())
		require.Equal(t, SHA256, c.Type())
	})

	t.Run("TZ", func(t *testing.T) {
		orig := tz.Sum(payload)

		c, err := NewFromData(TZ, payload)
		require.NoError(t, err)
		require.Equal(t, orig[:], c.Value())
		require.Equal(t, TZ, c.Type())

		c = Checksum{}
		Calculate(&c, TZ, payload)
		require.Equal(t, orig[:], c.Value())
		require.Equal(t, TZ, c.Type())
	})
}
