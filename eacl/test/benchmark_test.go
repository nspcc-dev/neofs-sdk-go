package eacltest

import (
	"bytes"
	"math/rand"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func baseBenchmarkTableBinaryComparison(b *testing.B, factor int) {
	t := TableN(factor)
	exp := t.Marshal()

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !bytes.Equal(exp, t.Marshal()) {
			b.Fail()
		}
	}
}

func baseBenchmarkTableEqualsComparison(b *testing.B, factor int) {
	t := TableN(factor)
	t2 := eacl.NewTable()
	err := t2.Unmarshal(t.Marshal())
	require.NoError(b, err)

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !eacl.EqualTables(*t, *t2) {
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

	for i := 0; i < n; i++ {
		keys[i] = make([]byte, 32)
		//nolint:staticcheck
		rand.Read(keys[i])
	}

	x.SetBinaryKeys(keys)

	return &x
}

// Record returns random eacl.Record.
func RecordN(n int) *eacl.Record {
	fs := make([]eacl.Filter, n)
	for i := 0; i < n; i++ {
		fs[i] = eacl.ConstructFilter(eacl.HeaderFromObject, "", eacl.MatchStringEqual, cidtest.ID().EncodeToString())
	}

	x := eacl.ConstructRecord(eacl.ActionAllow, eacl.OperationRangeHash, []eacl.Target{*TargetN(n)}, fs...)
	return &x
}

func TableN(n int) *eacl.Table {
	x := eacl.NewTable()

	x.SetCID(cidtest.ID())

	for i := 0; i < n; i++ {
		x.AddRecord(RecordN(n))
	}

	x.SetVersion(versiontest.Version())

	return x
}
