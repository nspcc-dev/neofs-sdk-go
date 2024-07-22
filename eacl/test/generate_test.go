package eacltest_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	require.NotEqual(t, eacltest.Target(), eacltest.Target())
}

func TestTargets(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, eacltest.Targets(n), n)
}

func TestFilter(t *testing.T) {
	require.NotEqual(t, eacltest.Filter(), eacltest.Filter())
}

func TestFilters(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, eacltest.Filters(n), n)
}

func TestRecord(t *testing.T) {
	require.NotEqual(t, eacltest.Record(), eacltest.Record())
}

func TestRecords(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, eacltest.Records(n), n)
}

func TestTable(t *testing.T) {
	eACL := eacltest.Table()
	require.NotEqual(t, eACL, eacltest.Table())

	var eACL2 eacl.Table
	require.NoError(t, eACL2.ReadFromV2(*eACL.ToV2()))
	require.Equal(t, eACL, eACL2)

	var eACL3 eacl.Table
	require.NoError(t, eACL3.Unmarshal(eACL.Marshal()))
	require.Equal(t, eACL, eACL3)

	j, err := eACL.MarshalJSON()
	require.NoError(t, err)
	var eACL4 eacl.Table
	require.NoError(t, eACL4.UnmarshalJSON(j))
	require.Equal(t, eACL, eACL4)
}
