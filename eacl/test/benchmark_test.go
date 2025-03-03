package eacltest

import (
	"bytes"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/stretchr/testify/require"
)

func baseBenchmarkTableBinaryComparison(b *testing.B, factor int) {
	t := TableN(factor)
	exp := t.Marshal()

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for range b.N {
		if !bytes.Equal(exp, t.Marshal()) {
			b.Fail()
		}
	}
}

func baseBenchmarkTableEqualsComparison(b *testing.B, factor int) {
	t := TableN(factor)
	var t2 eacl.Table
	err := t2.Unmarshal(t.Marshal())
	require.NoError(b, err)

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for range b.N {
		if !eacl.EqualTables(*t, t2) {
			b.Fail()
		}
	}
}

func BenchmarkTableBinaryComparison(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 1)
}

func BenchmarkTableEqualsComparison(b *testing.B) {
	baseBenchmarkTableEqualsComparison(b, 1)
}

func BenchmarkTableBinaryComparison10(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 10)
}

func BenchmarkTableEqualsComparison10(b *testing.B) {
	baseBenchmarkTableEqualsComparison(b, 10)
}

func BenchmarkTableBinaryComparison100(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 100)
}

func BenchmarkTableEqualsComparison100(b *testing.B) {
	baseBenchmarkTableEqualsComparison(b, 100)
}

// Target returns random eacl.Target.
func TargetN(n int) *eacl.Target {
	var x eacl.Target

	x.SetRole(eacl.RoleSystem)
	keys := make([][]byte, n)

	for i := range n {
		keys[i] = testutil.RandByteSlice(32)
	}

	x.SetBinaryKeys(keys)

	return &x
}

// Record returns random eacl.Record.
func RecordN(n int) *eacl.Record {
	fs := make([]eacl.Filter, n)
	for i := range n {
		fs[i] = eacl.ConstructFilter(eacl.HeaderFromObject, "", eacl.MatchStringEqual, cidtest.ID().EncodeToString())
	}

	x := eacl.ConstructRecord(eacl.ActionAllow, eacl.OperationRangeHash, []eacl.Target{*TargetN(n)}, fs...)
	return &x
}

func TableN(n int) *eacl.Table {
	rs := make([]eacl.Record, n)
	for i := range n {
		rs[i] = *RecordN(n)
	}
	x := eacl.NewTableForContainer(cidtest.ID(), rs)
	return &x
}
