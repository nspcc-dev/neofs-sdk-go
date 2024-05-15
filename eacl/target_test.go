package eacl_test

import (
	"bytes"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestTarget_Role(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.Role())

	tgt.SetRole(13)
	require.EqualValues(t, 13, tgt.Role())
	tgt.SetRole(42)
	require.EqualValues(t, 42, tgt.Role())
}

func TestTarget_PublicKeys(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.PublicKeys())

	tgt.SetPublicKeys([][]byte{[]byte("key1"), []byte("key2")})
	require.Equal(t, [][]byte{[]byte("key1"), []byte("key2")}, tgt.PublicKeys())
	tgt.SetPublicKeys([][]byte{[]byte("key3"), []byte("key4")})
	require.Equal(t, [][]byte{[]byte("key3"), []byte("key4")}, tgt.PublicKeys())
}

func TestTarget_CopyTo(t *testing.T) {
	src := eacltest.Target()
	src.SetPublicKeys([][]byte{[]byte("key1"), []byte("key2")})

	var dst eacl.Target
	src.CopyTo(&dst)
	require.Equal(t, src, dst)

	originKey := bytes.Clone(src.PublicKeys()[0])
	src.PublicKeys()[0][0]++
	require.Equal(t, originKey, dst.PublicKeys()[0])
}
