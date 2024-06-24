package netmap_test

import (
	"strconv"
	"strings"
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestPlacementPolicy_DecodeString(t *testing.T) {
	testCases := []string{
		`REP 1 IN X
CBF 1
SELECT 2 IN SAME Location FROM * AS X`,

		`REP 1
SELECT 2 IN City FROM Good
FILTER Country EQ RU AS FromRU
FILTER @FromRU AND Rating GT 7 AS Good`,

		`REP 7 IN SPB
SELECT 1 IN City FROM SPBSSD AS SPB
FILTER City EQ SPB AND SSD EQ true OR City EQ SPB AND Rating GE 5 AS SPBSSD`,
	}

	var p netmap.PlacementPolicy

	for _, testCase := range testCases {
		require.NoError(t, p.DecodeString(testCase))

		var b strings.Builder
		require.NoError(t, p.WriteStringTo(&b))
		require.Equal(t, testCase, b.String())
	}

	invalidTestCases := []string{
		`?REP 1`,
		`REP 1 trailing garbage`,
	}

	for i := range invalidTestCases {
		require.Error(t, p.DecodeString(invalidTestCases[i]), "#%d", i)
	}
}

func TestPlacementPolicy_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("replicas", func(t *testing.T) {
			n := netmaptest.PlacementPolicy()
			var m apinetmap.PlacementPolicy

			n.WriteToV2(&m)
			m.Replicas = nil
			require.ErrorContains(t, n.ReadFromV2(&m), "missing replicas")

			n.WriteToV2(&m)
			m.Replicas = []*apinetmap.Replica{}
			require.ErrorContains(t, n.ReadFromV2(&m), "missing replicas")
		})
	})
}

func TestPlacementPolicy_Marshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var n netmap.PlacementPolicy
		msg := []byte("definitely_not_protobuf")
		err := n.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
}

func TestPlacementPolicy_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var n netmap.PlacementPolicy
		msg := []byte("definitely_not_protojson")
		err := n.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
}

func TestPlacementPolicy_ContainerBackupFactor(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Zero(t, p.ContainerBackupFactor())

	const val = 42
	p.SetContainerBackupFactor(val)
	require.EqualValues(t, val, p.ContainerBackupFactor())

	const otherVal = 13
	p.SetContainerBackupFactor(otherVal)
	require.EqualValues(t, otherVal, p.ContainerBackupFactor())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			dst.SetContainerBackupFactor(otherVal)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.ContainerBackupFactor())

			src.SetContainerBackupFactor(val)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.EqualValues(t, val, dst.ContainerBackupFactor())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy
			var msg apinetmap.PlacementPolicy

			// set required data just to satisfy decoder
			src.SetReplicas(netmaptest.NReplicas(2))

			dst.SetContainerBackupFactor(otherVal)

			src.WriteToV2(&msg)
			require.Zero(t, msg.ContainerBackupFactor)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.ContainerBackupFactor())

			src.SetContainerBackupFactor(val)

			src.WriteToV2(&msg)
			require.EqualValues(t, val, msg.ContainerBackupFactor)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, dst.ContainerBackupFactor())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			dst.SetContainerBackupFactor(otherVal)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.ContainerBackupFactor())

			src.SetContainerBackupFactor(val)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.EqualValues(t, val, dst.ContainerBackupFactor())
		})
	})
}

func TestPlacementPolicy_Replicas(t *testing.T) {
	var p netmap.PlacementPolicy

	require.Zero(t, p.Replicas())
	require.Zero(t, p.NumberOfReplicas())

	rs := netmaptest.NReplicas(3)
	p.SetReplicas(rs)
	require.Equal(t, rs, p.Replicas())
	require.EqualValues(t, 3, p.NumberOfReplicas())
	for i := range rs {
		require.Equal(t, rs[i].NumberOfObjects(), p.ReplicaNumberByIndex(i))
	}
	require.Panics(t, func() { p.ReplicaNumberByIndex(len(rs)) })

	rsOther := netmaptest.NReplicas(2)
	p.SetReplicas(rsOther)
	require.Equal(t, rsOther, p.Replicas())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			src.SetReplicas(rs)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, rs, dst.Replicas())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy
			var msg apinetmap.PlacementPolicy

			src.WriteToV2(&msg)
			require.Zero(t, msg.Replicas)
			err := dst.ReadFromV2(&msg)
			require.ErrorContains(t, err, "missing replicas")

			rs := make([]netmap.ReplicaDescriptor, 3)
			for i := range rs {
				rs[i].SetNumberOfObjects(uint32(i + 1))
				rs[i].SetSelectorName("selector_" + strconv.Itoa(i+1))
			}

			src.SetReplicas(rs)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.Replica{
				{Count: 1, Selector: "selector_1"},
				{Count: 2, Selector: "selector_2"},
				{Count: 3, Selector: "selector_3"},
			}, msg.Replicas)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, rs, dst.Replicas())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			src.SetReplicas(rs)
			j, err := src.MarshalJSON()
			require.NoError(t, err)

			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, rs, dst.Replicas())
		})
	})
}

func TestPlacementPolicy_Filters(t *testing.T) {
	var p netmap.PlacementPolicy

	require.Zero(t, p.Filters())

	fs := netmaptest.NFilters(3)
	p.SetFilters(fs)
	require.Equal(t, fs, p.Filters())

	fsOther := netmaptest.NFilters(2)
	p.SetFilters(fsOther)
	require.Equal(t, fsOther, p.Filters())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			dst.SetFilters(fsOther)
			src.SetFilters(fs)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, fs, dst.Filters())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy
			var msg apinetmap.PlacementPolicy

			// set required data just to satisfy decoder
			src.SetReplicas(netmaptest.NReplicas(3))

			dst.SetFilters(fs)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Filters)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Zero(t, dst.Filters())

			fs := make([]netmap.Filter, 8)
			for i := range fs {
				fs[i].SetName("name" + strconv.Itoa(i))
			}
			fs[0].Equal("key0", "val0")
			fs[1].NotEqual("key1", "val1")
			fs[2].NumericGT("key2", 2)
			fs[3].NumericGE("key3", 3)
			fs[4].NumericLT("key4", 4)
			fs[5].NumericLE("key5", 5)
			subs0 := make([]netmap.Filter, 2)
			subs0[0].SetName("sub0_0")
			subs0[0].Equal("key0_0", "val0_0")
			subs0[1].SetName("sub0_1")
			subs0[1].NotEqual("key0_1", "val0_1")
			fs[6].LogicalOR(subs0...)
			subs1 := make([]netmap.Filter, 2)
			subs1[0].SetName("sub1_0")
			subs1[0].NumericGT("key1_0", 6)
			subs1[1].SetName("sub1_1")
			subs1[1].NumericGE("key1_1", 7)
			fs[7].LogicalAND(subs1...)

			src.SetFilters(fs)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.Filter{
				{Name: "name0", Key: "key0", Op: apinetmap.Operation_EQ, Value: "val0"},
				{Name: "name1", Key: "key1", Op: apinetmap.Operation_NE, Value: "val1"},
				{Name: "name2", Key: "key2", Op: apinetmap.Operation_GT, Value: "2"},
				{Name: "name3", Key: "key3", Op: apinetmap.Operation_GE, Value: "3"},
				{Name: "name4", Key: "key4", Op: apinetmap.Operation_LT, Value: "4"},
				{Name: "name5", Key: "key5", Op: apinetmap.Operation_LE, Value: "5"},
				{Name: "name6", Key: "", Op: apinetmap.Operation_OR, Value: "", Filters: []*apinetmap.Filter{
					{Name: "sub0_0", Key: "key0_0", Op: apinetmap.Operation_EQ, Value: "val0_0"},
					{Name: "sub0_1", Key: "key0_1", Op: apinetmap.Operation_NE, Value: "val0_1"},
				}},
				{Name: "name7", Key: "", Op: apinetmap.Operation_AND, Value: "", Filters: []*apinetmap.Filter{
					{Name: "sub1_0", Key: "key1_0", Op: apinetmap.Operation_GT, Value: "6"},
					{Name: "sub1_1", Key: "key1_1", Op: apinetmap.Operation_GE, Value: "7"},
				}},
			}, msg.Filters)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, fs, dst.Filters())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			src.SetFilters(fs)
			j, err := src.MarshalJSON()
			require.NoError(t, err)

			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, fs, dst.Filters())
		})
	})
}
func TestPlacementPolicy_Selectors(t *testing.T) {
	var p netmap.PlacementPolicy

	require.Zero(t, p.Selectors())

	ss := netmaptest.NSelectors(3)
	p.SetSelectors(ss)
	require.Equal(t, ss, p.Selectors())

	ssOther := netmaptest.NSelectors(2)
	p.SetSelectors(ssOther)
	require.Equal(t, ssOther, p.Selectors())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			dst.SetSelectors(ssOther)
			src.SetSelectors(ss)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, ss, dst.Selectors())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy
			var msg apinetmap.PlacementPolicy

			// set required data just to satisfy decoder
			src.SetReplicas(netmaptest.NReplicas(3))

			dst.SetSelectors(ss)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Selectors)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Zero(t, dst.Selectors())

			ss := make([]netmap.Selector, 2)
			for i := range ss {
				si := strconv.Itoa(i + 1)
				ss[i].SetName("name" + si)
				ss[i].SetFilterName("filter" + si)
				ss[i].SelectByBucketAttribute("bucket" + si)
				ss[i].SetNumberOfNodes(uint32(i + 1))
			}
			ss[0].SelectSame()
			ss[1].SelectDistinct()

			src.SetSelectors(ss)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.Selector{
				{Name: "name1", Count: 1, Clause: apinetmap.Clause_SAME, Attribute: "bucket1", Filter: "filter1"},
				{Name: "name2", Count: 2, Clause: apinetmap.Clause_DISTINCT, Attribute: "bucket2", Filter: "filter2"},
			}, msg.Selectors)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, ss, dst.Selectors())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.PlacementPolicy

			src.SetSelectors(ss)
			j, err := src.MarshalJSON()
			require.NoError(t, err)

			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, ss, dst.Selectors())
		})
	})
}
