package netmap_test

import (
	"strings"
	"testing"

	netmapv2 "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
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

func TestPlacementPolicyEncoding(t *testing.T) {
	v := netmaptest.PlacementPolicy()

	t.Run("binary", func(t *testing.T) {
		var v2 netmap.PlacementPolicy
		require.NoError(t, v2.Unmarshal(v.Marshal()))

		require.Equal(t, v, v2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := v.MarshalJSON()
		require.NoError(t, err)

		var v2 netmap.PlacementPolicy
		require.NoError(t, v2.UnmarshalJSON(data))

		require.Equal(t, v, v2)
	})
}

func TestPlacementPolicy_ContainerBackupFactor(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Zero(t, p.ContainerBackupFactor())

	p = netmaptest.PlacementPolicy()
	p.SetContainerBackupFactor(42)
	require.EqualValues(t, 42, p.ContainerBackupFactor())

	var m netmapv2.PlacementPolicy
	p.WriteToV2(&m)
	require.EqualValues(t, 42, m.GetContainerBackupFactor())

	m.SetContainerBackupFactor(13)
	err := p.ReadFromV2(m)
	require.NoError(t, err)
	require.EqualValues(t, 13, m.GetContainerBackupFactor())
}

func TestPlacementPolicy_Replicas(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Replicas())

	var r netmap.ReplicaDescriptor
	var rs []netmap.ReplicaDescriptor

	r.SetSelectorName("selector_1")
	r.SetNumberOfObjects(1)
	rs = append(rs, r)

	r.SetSelectorName("selector_2")
	r.SetNumberOfObjects(2)
	rs = append(rs, r)

	p.AddReplicas(rs...)
	require.Equal(t, rs, p.Replicas())

	var m netmapv2.PlacementPolicy
	p.WriteToV2(&m)
	rsm := m.GetReplicas()
	require.Len(t, rsm, 2)
	rm := rsm[0]
	require.Equal(t, "selector_1", rm.GetSelector())
	require.EqualValues(t, 1, rm.GetCount())
	rm = rsm[1]
	require.Equal(t, "selector_2", rm.GetSelector())
	require.EqualValues(t, 2, rm.GetCount())

	err := p.ReadFromV2(m)
	require.NoError(t, err)

	rs = p.Replicas()
	r = rs[0]
	require.Equal(t, "selector_1", r.SelectorName())
	require.EqualValues(t, 1, r.NumberOfObjects())
	r = rs[1]
	require.Equal(t, "selector_2", r.SelectorName())
	require.EqualValues(t, 2, r.NumberOfObjects())
}

func TestPlacementPolicy_Selectors(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Selectors())

	var s netmap.Selector
	var ss []netmap.Selector

	s.SetName("name_1")
	s.SelectByBucketAttribute("bucket_1")
	s.SetFilterName("filter_1")
	s.SetNumberOfNodes(1)
	s.SelectSame()
	ss = append(ss, s)

	s.SetName("name_2")
	s.SelectByBucketAttribute("bucket_2")
	s.SetFilterName("filter_2")
	s.SetNumberOfNodes(2)
	s.SelectDistinct()
	ss = append(ss, s)

	p.AddSelectors(ss...)
	require.Equal(t, ss, p.Selectors())

	var m netmapv2.PlacementPolicy
	p.WriteToV2(&m)
	ssm := m.GetSelectors()
	require.Len(t, ssm, 2)
	sm := ssm[0]
	require.Equal(t, "name_1", sm.GetName())
	require.Equal(t, "bucket_1", sm.GetAttribute())
	require.Equal(t, "filter_1", sm.GetFilter())
	require.EqualValues(t, 1, sm.GetCount())
	require.Equal(t, netmapv2.Same, sm.GetClause())
	sm = ssm[1]
	require.Equal(t, "name_2", sm.GetName())
	require.Equal(t, "bucket_2", sm.GetAttribute())
	require.Equal(t, "filter_2", sm.GetFilter())
	require.EqualValues(t, 2, sm.GetCount())
	require.Equal(t, netmapv2.Distinct, sm.GetClause())

	m.SetReplicas([]netmapv2.Replica{{}}) // required
	err := p.ReadFromV2(m)
	require.NoError(t, err)

	ss = p.Selectors()
	s = ss[0]
	require.Equal(t, "name_1", s.Name())
	require.Equal(t, "bucket_1", s.BucketAttribute())
	require.Equal(t, "filter_1", s.FilterName())
	require.EqualValues(t, 1, s.NumberOfNodes())
	require.True(t, s.IsSame())
	s = ss[1]
	require.Equal(t, "name_2", s.Name())
	require.Equal(t, "bucket_2", s.BucketAttribute())
	require.Equal(t, "filter_2", s.FilterName())
	require.EqualValues(t, 2, s.NumberOfNodes())
	require.True(t, s.IsDistinct())
}

func TestPlacementPolicy_Filters(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Filters())

	var f netmap.Filter
	var fs []netmap.Filter

	// 'key_1' == 'val_1'
	f.SetName("filter_1")
	f.Equal("key_1", "val_1")
	fs = append(fs, f)

	// 'key_2' != 'val_2'
	f.SetName("filter_2")
	f.NotEqual("key_2", "val_2")
	fs = append(fs, f)

	// 'key_3_1' > 31 || 'key_3_2' >= 32
	var sub1 netmap.Filter
	sub1.SetName("filter_3_1")
	sub1.NumericGT("key_3_1", 3_1)

	var sub2 netmap.Filter
	sub2.SetName("filter_3_2")
	sub2.NumericGE("key_3_2", 3_2)

	f.SetName("filter_3")
	f.LogicalOR(sub1, sub2)
	fs = append(fs, f)

	// 'key_4_1' < 41 || 'key_4_2' <= 42
	sub1.SetName("filter_4_1")
	sub1.NumericLT("key_4_1", 4_1)

	sub2.SetName("filter_4_2")
	sub2.NumericLE("key_4_2", 4_2)

	f = netmap.Filter{}
	f.SetName("filter_4")
	f.LogicalAND(sub1, sub2)
	fs = append(fs, f)

	p.AddFilters(fs...)
	require.Equal(t, fs, p.Filters())

	var m netmapv2.PlacementPolicy
	p.WriteToV2(&m)
	fsm := m.GetFilters()
	require.Len(t, fsm, 4)
	// 1
	fm := fsm[0]
	require.Equal(t, "filter_1", fm.GetName())
	require.Equal(t, "key_1", fm.GetKey())
	require.Equal(t, netmapv2.EQ, fm.GetOp())
	require.Equal(t, "val_1", fm.GetValue())
	require.Zero(t, fm.GetFilters())
	// 2
	fm = fsm[1]
	require.Equal(t, "filter_2", fm.GetName())
	require.Equal(t, "key_2", fm.GetKey())
	require.Equal(t, netmapv2.NE, fm.GetOp())
	require.Equal(t, "val_2", fm.GetValue())
	require.Zero(t, fm.GetFilters())
	// 3
	fm = fsm[2]
	require.Equal(t, "filter_3", fm.GetName())
	require.Zero(t, fm.GetKey())
	require.Equal(t, netmapv2.OR, fm.GetOp())
	require.Zero(t, fm.GetValue())
	// 3.1
	subm := fm.GetFilters()
	require.Len(t, subm, 2)
	fm = subm[0]
	require.Equal(t, "filter_3_1", fm.GetName())
	require.Equal(t, "key_3_1", fm.GetKey())
	require.Equal(t, netmapv2.GT, fm.GetOp())
	require.Equal(t, "31", fm.GetValue())
	require.Zero(t, fm.GetFilters())
	// 3.2
	fm = subm[1]
	require.Equal(t, "filter_3_2", fm.GetName())
	require.Equal(t, "key_3_2", fm.GetKey())
	require.Equal(t, netmapv2.GE, fm.GetOp())
	require.Equal(t, "32", fm.GetValue())
	require.Zero(t, fm.GetFilters())
	// 4
	fm = fsm[3]
	require.Equal(t, "filter_4", fm.GetName())
	require.Zero(t, fm.GetKey())
	require.Equal(t, netmapv2.AND, fm.GetOp())
	require.Zero(t, fm.GetValue())
	// 4.1
	subm = fm.GetFilters()
	require.Len(t, subm, 2)
	fm = subm[0]
	require.Equal(t, "filter_4_1", fm.GetName())
	require.Equal(t, "key_4_1", fm.GetKey())
	require.Equal(t, netmapv2.LT, fm.GetOp())
	require.Equal(t, "41", fm.GetValue())
	require.Zero(t, fm.GetFilters())
	// 4.2
	fm = subm[1]
	require.Equal(t, "filter_4_2", fm.GetName())
	require.Equal(t, "key_4_2", fm.GetKey())
	require.Equal(t, netmapv2.LE, fm.GetOp())
	require.Equal(t, "42", fm.GetValue())
	require.Zero(t, fm.GetFilters())

	m.SetReplicas([]netmapv2.Replica{{}}) // required
	err := p.ReadFromV2(m)
	require.NoError(t, err)

	fs = p.Filters()
	require.Len(t, fs, 4)
	// 1
	f = fs[0]
	require.Equal(t, "filter_1", f.Name())
	require.Equal(t, "key_1", f.Key())
	require.Equal(t, netmap.FilterOpEQ, f.Op())
	require.Equal(t, "val_1", f.Value())
	require.Zero(t, f.SubFilters())
	// 2
	f = fs[1]
	require.Equal(t, "filter_2", f.Name())
	require.Equal(t, "key_2", f.Key())
	require.Equal(t, netmap.FilterOpNE, f.Op())
	require.Equal(t, "val_2", f.Value())
	require.Zero(t, f.SubFilters())
	// 3
	f = fs[2]
	require.Equal(t, "filter_3", f.Name())
	require.Zero(t, f.Key())
	require.Equal(t, netmap.FilterOpOR, f.Op())
	require.Zero(t, f.Value())
	// 3.1
	sub := f.SubFilters()
	require.Len(t, sub, 2)
	f = sub[0]
	require.Equal(t, "filter_3_1", f.Name())
	require.Equal(t, "key_3_1", f.Key())
	require.Equal(t, netmap.FilterOpGT, f.Op())
	require.Equal(t, "31", f.Value())
	require.Zero(t, f.SubFilters())
	// 3.2
	f = sub[1]
	require.Equal(t, "filter_3_2", f.Name())
	require.Equal(t, "key_3_2", f.Key())
	require.Equal(t, netmap.FilterOpGE, f.Op())
	require.Equal(t, "32", f.Value())
	require.Zero(t, f.SubFilters())
	// 4
	f = fs[3]
	require.Equal(t, "filter_4", f.Name())
	require.Zero(t, f.Key())
	require.Equal(t, netmap.FilterOpAND, f.Op())
	require.Zero(t, f.Value())
	// 4.1
	sub = f.SubFilters()
	require.Len(t, sub, 2)
	f = sub[0]
	require.Equal(t, "filter_4_1", f.Name())
	require.Equal(t, "key_4_1", f.Key())
	require.Equal(t, netmap.FilterOpLT, f.Op())
	require.Equal(t, "41", f.Value())
	require.Zero(t, f.SubFilters())
	// 4.2
	f = sub[1]
	require.Equal(t, "filter_4_2", f.Name())
	require.Equal(t, "key_4_2", f.Key())
	require.Equal(t, netmap.FilterOpLE, f.Op())
	require.Equal(t, "42", f.Value())
	require.Zero(t, f.SubFilters())
}
