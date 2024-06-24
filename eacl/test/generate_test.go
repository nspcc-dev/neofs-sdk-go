package eacltest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	require.NotEqual(t, eacltest.Filter(), eacltest.Filter())
}

func TestNFilters(t *testing.T) {
	res := eacltest.NFilters(3)
	require.Len(t, res, 3)
	require.NotEqual(t, res, eacltest.NFilters(3))
}

func TestTarget(t *testing.T) {
	require.NotEqual(t, eacltest.Target(), eacltest.Target())
}

func TestNTargets(t *testing.T) {
	res := eacltest.NTargets(3)
	require.Len(t, res, 3)
	require.NotEqual(t, res, eacltest.NTargets(3))
}

func TestRecord(t *testing.T) {
	require.NotEqual(t, eacltest.Record(), eacltest.Record())
}

func TestNRecords(t *testing.T) {
	res := eacltest.NRecords(3)
	require.Len(t, res, 3)
	require.NotEqual(t, res, eacltest.NRecords(3))
}

func TestTable(t *testing.T) {
	tbl := eacltest.Table()
	require.NotEqual(t, tbl, eacltest.Table())

	var tbl2 eacl.Table
	require.NoError(t, tbl2.Unmarshal(tbl.Marshal()))
	require.Equal(t, tbl, tbl2)

	var m acl.EACLTable
	tbl.WriteToV2(&m)
	var tbl3 eacl.Table
	require.NoError(t, tbl3.ReadFromV2(&m))
	require.Equal(t, tbl, tbl3)
}
