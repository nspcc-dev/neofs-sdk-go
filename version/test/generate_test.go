package versiontest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	v := versiontest.Version()
	require.NotEqual(t, v, versiontest.Version())

	var m refs.Version
	v.WriteToV2(&m)
	var v2 version.Version
	require.NoError(t, v2.ReadFromV2(&m))
}
