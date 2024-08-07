package eacl

import (
	"bytes"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func TestTable_CopyTo(t *testing.T) {
	id := cidtest.ID()

	var table Table
	table.SetVersion(version.Current())
	table.SetCID(id)

	var target Target
	target.SetRole(1)
	target.SetBinaryKeys([][]byte{
		{1, 2, 3},
	})

	record := CreateRecord(ActionAllow, OperationPut)
	record.SetTargets(target)
	record.AddObjectAttributeFilter(MatchStringEqual, "key", "value")

	table.AddRecord(record)

	t.Run("copy", func(t *testing.T) {
		var dst Table
		table.CopyTo(&dst)

		require.Equal(t, table, dst)
		require.True(t, bytes.Equal(table.Marshal(), dst.Marshal()))
	})

	t.Run("change version", func(t *testing.T) {
		var dst Table
		table.CopyTo(&dst)

		require.True(t, table.Version().Equal(dst.Version()))

		var newVersion version.Version
		newVersion.SetMajor(10)
		newVersion.SetMinor(100)

		dst.SetVersion(newVersion)

		require.False(t, table.Version().Equal(dst.Version()))
	})

	t.Run("change cid", func(t *testing.T) {
		var dst Table
		table.CopyTo(&dst)
		require.Equal(t, table.GetCID(), dst.GetCID())

		dst.SetCID(cidtest.OtherID(id))
		require.NotEqual(t, table.GetCID(), dst.GetCID())
	})

	t.Run("change record", func(t *testing.T) {
		var dst Table
		table.CopyTo(&dst)

		require.Equal(t, table.records[0].action, dst.records[0].action)
		dst.records[0].SetAction(ActionDeny)
		require.NotEqual(t, table.records[0].action, dst.records[0].action)

		require.Equal(t, table.records[0].operation, dst.records[0].operation)
		dst.records[0].SetOperation(OperationDelete)
		require.NotEqual(t, table.records[0].operation, dst.records[0].operation)

		require.Equal(t, table.records[0].targets[0].role, dst.records[0].targets[0].role)
		table.records[0].targets[0].SetRole(1234)
		require.NotEqual(t, table.records[0].targets[0].role, dst.records[0].targets[0].role)
	})
}
