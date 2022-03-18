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
	exp, err := t.Marshal()
	require.NoError(b, err)

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		got, _ := t.Marshal()
		if !bytes.Equal(exp, got) {
			b.Fail()
		}
	}
}

func baseBenchmarkTableEqualToComparison(b *testing.B, factor int) {
	t := TableN(factor)
	data, err := t.Marshal()
	require.NoError(b, err)
	t2 := eacl.NewTable()
	err = t2.Unmarshal(data)
	require.NoError(b, err)

	b.StopTimer()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !t.EqualTo(t2) {
			b.Fail()
		}
	}
}

func BenchmarkTableBinaryComparison(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 1)
}

func BenchmarkTableEqualToComparison(b *testing.B) {
	baseBenchmarkTableEqualToComparison(b, 1)
}

func BenchmarkTableBinaryComparison10(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 10)
}

func BenchmarkTableEqualToComparison10(b *testing.B) {
	baseBenchmarkTableEqualToComparison(b, 10)
}

func BenchmarkTableBinaryComparison100(b *testing.B) {
	baseBenchmarkTableBinaryComparison(b, 100)
}

func BenchmarkTableEqualToComparison100(b *testing.B) {
	baseBenchmarkTableEqualToComparison(b, 100)
}

// Target returns random eacl.Target.
func TargetN(n int) *eacl.Target {
	x := eacl.NewTarget()

	x.SetRole(eacl.RoleSystem)
	keys := make([][]byte, n)

	for i := 0; i < n; i++ {
		keys[i] = make([]byte, 32)
		rand.Read(keys[i])
	}

	x.SetBinaryKeys(keys)

	return x
}

// Record returns random eacl.Record.
func RecordN(n int) *eacl.Record {
	x := eacl.NewRecord()

	x.SetAction(eacl.ActionAllow)
	x.SetOperation(eacl.OperationRangeHash)
	x.SetTargets(*TargetN(n))

	for i := 0; i < n; i++ {
		x.AddFilter(eacl.HeaderFromObject, eacl.MatchStringEqual, "", cidtest.ID().String())
	}

	return x
}

func TableN(n int) *eacl.Table {
	x := eacl.NewTable()

	x.SetCID(cidtest.ID())

	for i := 0; i < n; i++ {
		x.AddRecord(RecordN(n))
	}

	x.SetVersion(*versiontest.Version())

	return x
}
