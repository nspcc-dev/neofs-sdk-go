package eacl

import (
	"bytes"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestRecord_CopyTo(t *testing.T) {
	var record = ConstructRecord(ActionAllow, OperationPut,
		[]Target{NewTargetByRole(1), NewTargetByScriptHashes([]util.Uint160{{1, 2, 3}})},
		NewObjectPropertyFilter("key", MatchStringEqual, "value"))

	t.Run("copy", func(t *testing.T) {
		var dst Record
		record.CopyTo(&dst)

		require.Equal(t, record, dst)
		require.True(t, bytes.Equal(record.Marshal(), dst.Marshal()))
	})

	t.Run("change filters", func(t *testing.T) {
		var dst Record
		record.CopyTo(&dst)

		require.Equal(t, record.filters[0].key, dst.filters[0].key)
		require.Equal(t, record.filters[0].matcher, dst.filters[0].matcher)
		require.Equal(t, record.filters[0].value, dst.filters[0].value)
		require.Equal(t, record.filters[0].from, dst.filters[0].from)

		dst.filters[0].key = "key2"
		dst.filters[0].matcher = MatchStringNotEqual
		dst.filters[0].value = staticStringer("staticStringer")
		dst.filters[0].from = 12345

		require.NotEqual(t, record.filters[0].key, dst.filters[0].key)
		require.NotEqual(t, record.filters[0].matcher, dst.filters[0].matcher)
		require.NotEqual(t, record.filters[0].value, dst.filters[0].value)
		require.NotEqual(t, record.filters[0].from, dst.filters[0].from)
	})

	t.Run("change target", func(t *testing.T) {
		var dst Record
		record.CopyTo(&dst)

		require.Equal(t, record.targets[0].role, dst.targets[0].role)
		dst.targets[0].role = 12345
		require.NotEqual(t, record.targets[0].role, dst.targets[0].role)

		for i, key := range dst.targets[0].subjs {
			require.True(t, bytes.Equal(key, record.targets[0].subjs[i]))
			key[0] = 10
			require.False(t, bytes.Equal(key, record.targets[0].subjs[i]))
		}
	})
}
