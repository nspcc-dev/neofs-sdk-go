package object_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestLinkEncoding(t *testing.T) {
	link := *objecttest.Link()

	t.Run("binary", func(t *testing.T) {
		data := link.Marshal()

		var link2 object.Link
		require.NoError(t, link2.Unmarshal(data))

		require.Equal(t, link, link2)
	})
}

func TestWriteLink(t *testing.T) {
	link := objecttest.Link()
	var link2 object.Link
	var o object.Object

	o.WriteLink(*link)

	require.NoError(t, o.ReadLink(&link2))
	require.Equal(t, *link, link2)

	// corrupt payload
	o.Payload()[0]++

	require.Error(t, o.ReadLink(&link2))
}
