package netmap_test

import (
	"encoding/json"
	"strings"
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

const (
	anyValidBackupFactor = 153493707
)

var (
	// set by init.
	anyValidReplicaDescriptors = make([]netmap.ReplicaDescriptor, 2)
	anyValidSelectors          = make([]netmap.Selector, 2)
	anyValidFilters            = make([]netmap.Filter, 2)
)

// set by init.
var validPlacementPolicy netmap.PlacementPolicy

// corresponds to validPlacementPolicy.
var validBinPlacementPolicy = []byte{
	10, 18, 8, 178, 191, 131, 208, 9, 18, 10, 115, 101, 108, 101, 99, 116, 111, 114, 95, 48, 10, 18, 8, 154, 216, 136, 171, 1, 18, 10, 115,
	101, 108, 101, 99, 116, 111, 114, 95, 49, 16, 203, 193, 152, 73, 26, 43, 10, 10, 115, 101, 108, 101, 99, 116, 111, 114, 95, 48, 16,
	148, 185, 173, 225, 6, 24, 1, 34, 11, 97, 116, 116, 114, 105, 98, 117, 116, 101, 95, 48, 42, 8, 102, 105, 108, 116, 101, 114, 95,
	48, 26, 43, 10, 10, 115, 101, 108, 101, 99, 116, 111, 114, 95, 49, 16, 225, 160, 218, 205, 5, 24, 2, 34, 11, 97, 116, 116, 114,
	105, 98, 117, 116, 101, 95, 49, 42, 8, 102, 105, 108, 116, 101, 114, 95, 49, 34, 80, 10, 8, 102, 105, 108, 116, 101, 114, 95, 48,
	24, 8, 42, 32, 10, 10, 102, 105, 108, 116, 101, 114, 95, 48, 95, 48, 18, 7, 107, 101, 121, 95, 48, 95, 48, 24, 1, 34, 7, 118,
	97, 108, 95, 48, 95, 48, 42, 32, 10, 10, 102, 105, 108, 116, 101, 114, 95, 48, 95, 49, 18, 7, 107, 101, 121, 95, 48, 95, 49, 24,
	2, 34, 7, 118, 97, 108, 95, 48, 95, 49, 34, 196, 1, 10, 8, 102, 105, 108, 116, 101, 114, 95, 49, 24, 7, 42, 44, 10, 10, 102, 105,
	108, 116, 101, 114, 95, 49, 95, 48, 18, 7, 107, 101, 121, 95, 49, 95, 48, 24, 3, 34, 19, 49, 56, 56, 57, 52, 48, 55, 55, 48, 56,
	57, 56, 53, 48, 50, 51, 49, 49, 54, 42, 44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 49, 18, 7, 107, 101, 121, 95, 49,
	95, 49, 24, 4, 34, 19, 49, 52, 50, 57, 50, 52, 51, 48, 57, 55, 51, 49, 53, 51, 52, 52, 56, 56, 56, 42, 44, 10, 10, 102, 105,
	108, 116, 101, 114, 95, 49, 95, 50, 18, 7, 107, 101, 121, 95, 49, 95, 50, 24, 5, 34, 19, 51, 55, 50, 50, 54, 53, 54, 48, 54, 48,
	51, 49, 55, 52, 56, 50, 51, 51, 53, 42, 44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 51, 18, 7, 107, 101, 121, 95, 49, 95,
	51, 24, 6, 34, 19, 49, 57, 53, 48, 53, 48, 52, 57, 56, 55, 55, 48, 53, 50, 56, 52, 56, 48, 53,
}

var validJSONPlacementPolicy = `
{
 "replicas": [
  {
   "count": 2583748530,
   "selector": "selector_0"
  },
  {
   "count": 358755354,
   "selector": "selector_1"
  }
 ],
 "containerBackupFactor": 153493707,
 "selectors": [
  {
   "name": "selector_0",
   "count": 1814781076,
   "clause": "SAME",
   "attribute": "attribute_0",
   "filter": "filter_0"
  },
  {
   "name": "selector_1",
   "count": 1505136737,
   "clause": "DISTINCT",
   "attribute": "attribute_1",
   "filter": "filter_1"
  }
 ],
 "filters": [
  {
   "name": "filter_0",
   "key": "",
   "op": "AND",
   "value": "",
   "filters": [
    {
     "name": "filter_0_0",
     "key": "key_0_0",
     "op": "EQ",
     "value": "val_0_0",
     "filters": []
    },
    {
     "name": "filter_0_1",
     "key": "key_0_1",
     "op": "NE",
     "value": "val_0_1",
     "filters": []
    }
   ]
  },
  {
   "name": "filter_1",
   "key": "",
   "op": "OR",
   "value": "",
   "filters": [
    {
     "name": "filter_1_0",
     "key": "key_1_0",
     "op": "GT",
     "value": "1889407708985023116",
     "filters": []
    },
    {
     "name": "filter_1_1",
     "key": "key_1_1",
     "op": "GE",
     "value": "1429243097315344888",
     "filters": []
    },
    {
     "name": "filter_1_2",
     "key": "key_1_2",
     "op": "LT",
     "value": "3722656060317482335",
     "filters": []
    },
    {
     "name": "filter_1_3",
     "key": "key_1_3",
     "op": "LE",
     "value": "1950504987705284805",
     "filters": []
    }
   ]
  }
 ],
 "subnetId": null
}
`

func init() {
	validPlacementPolicy.SetContainerBackupFactor(anyValidBackupFactor)
	// replicas
	anyValidReplicaDescriptors[0].SetSelectorName("selector_0")
	anyValidReplicaDescriptors[0].SetNumberOfObjects(2583748530)
	anyValidReplicaDescriptors[1].SetSelectorName("selector_1")
	anyValidReplicaDescriptors[1].SetNumberOfObjects(358755354)
	validPlacementPolicy.SetReplicas(anyValidReplicaDescriptors)
	// selectors
	anyValidSelectors[0].SetName("selector_0")
	anyValidSelectors[0].SetNumberOfNodes(1814781076)
	anyValidSelectors[0].SelectSame()
	anyValidSelectors[0].SetFilterName("filter_0")
	anyValidSelectors[0].SelectByBucketAttribute("attribute_0")
	anyValidSelectors[1].SetName("selector_1")
	anyValidSelectors[1].SetNumberOfNodes(1505136737)
	anyValidSelectors[1].SelectDistinct()
	anyValidSelectors[1].SetFilterName("filter_1")
	anyValidSelectors[1].SelectByBucketAttribute("attribute_1")
	validPlacementPolicy.SetSelectors(anyValidSelectors)
	// filters
	fs := make([]netmap.Filter, 2)
	subs := make([]netmap.Filter, 2)
	subs[0].SetName("filter_0_0")
	subs[0].Equal("key_0_0", "val_0_0")
	subs[1].SetName("filter_0_1")
	subs[1].NotEqual("key_0_1", "val_0_1")
	fs[0].SetName("filter_0")
	fs[0].LogicalAND(subs...)
	subs = make([]netmap.Filter, 4)
	subs[0].SetName("filter_1_0")
	subs[0].NumericGT("key_1_0", 1889407708985023116)
	subs[1].SetName("filter_1_1")
	subs[1].NumericGE("key_1_1", 1429243097315344888)
	subs[2].SetName("filter_1_2")
	subs[2].NumericLT("key_1_2", 3722656060317482335)
	subs[3].SetName("filter_1_3")
	subs[3].NumericLE("key_1_3", 1950504987705284805)
	fs[1].SetName("filter_1")
	fs[1].LogicalOR(subs...)
	validPlacementPolicy.SetFilters(fs)
}

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

func TestPlacementPolicy_SetContainerBackupFactor(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Zero(t, p.ContainerBackupFactor())

	p.SetContainerBackupFactor(anyValidBackupFactor)
	require.EqualValues(t, anyValidBackupFactor, p.ContainerBackupFactor())

	p.SetContainerBackupFactor(anyValidBackupFactor + 1)
	require.EqualValues(t, anyValidBackupFactor+1, p.ContainerBackupFactor())
}

func TestPlacementPolicy_SetReplicas(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Replicas())
	require.Zero(t, p.NumberOfReplicas())
	require.Panics(t, func() { p.ReplicaNumberByIndex(0) })

	p.SetReplicas(anyValidReplicaDescriptors)
	require.Equal(t, anyValidReplicaDescriptors, p.Replicas())
	require.EqualValues(t, 2, p.NumberOfReplicas())
	require.EqualValues(t, 2583748530, p.ReplicaNumberByIndex(0))
	require.EqualValues(t, 358755354, p.ReplicaNumberByIndex(1))
}

func TestPlacementPolicy_SetSelectors(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Selectors())

	p.SetSelectors(anyValidSelectors)
	require.Equal(t, anyValidSelectors, p.Selectors())
}

func TestPlacementPolicy_SetFilters(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Empty(t, p.Filters())

	p.SetFilters(anyValidFilters)
	require.Equal(t, anyValidFilters, p.Filters())
}

func TestPlacementPolicy_ReadFromV2(t *testing.T) {
	var m apinetmap.PlacementPolicy
	m.SetContainerBackupFactor(anyValidBackupFactor)
	mrs := make([]apinetmap.Replica, 2)
	mrs[0].SetSelector("selector_0")
	mrs[0].SetCount(2583748530)
	mrs[1].SetSelector("selector_1")
	mrs[1].SetCount(358755354)
	m.SetReplicas(mrs)
	mss := make([]apinetmap.Selector, 2)
	mss[0].SetName("selector_0")
	mss[0].SetCount(1814781076)
	mss[0].SetClause(apinetmap.Same)
	mss[0].SetFilter("filter_0")
	mss[0].SetAttribute("attribute_0")
	mss[1].SetName("selector_1")
	mss[1].SetCount(1814781076)
	mss[1].SetClause(apinetmap.Distinct)
	mss[1].SetFilter("filter_1")
	mss[1].SetAttribute("attribute_1")
	m.SetSelectors(mss)
	msubs := make([]apinetmap.Filter, 0, 2)
	addSub := func(name, key string, op apinetmap.Operation, val string) {
		var f apinetmap.Filter
		f.SetName(name)
		f.SetKey(key)
		f.SetOp(op)
		f.SetValue(val)
		msubs = append(msubs, f)
	}
	addSub("filter_0_0", "key_0_0", apinetmap.EQ, "val_0_0")
	addSub("filter_0_1", "key_0_1", apinetmap.NE, "val_0_1")
	mfs := make([]apinetmap.Filter, 2)
	mfs[0].SetName("filter_0")
	mfs[0].SetOp(apinetmap.AND)
	mfs[0].SetFilters(msubs)
	msubs = make([]apinetmap.Filter, 0, 4)
	addSub("filter_1_0", "key_1_0", apinetmap.GT, "1889407708985023116")
	addSub("filter_1_1", "key_1_1", apinetmap.GE, "1429243097315344888")
	addSub("filter_1_2", "key_1_2", apinetmap.LT, "3722656060317482335")
	addSub("filter_1_3", "key_1_3", apinetmap.LE, "1950504987705284805")
	mfs[1].SetName("filter_1")
	mfs[1].SetOp(apinetmap.OR)
	mfs[1].SetFilters(msubs)
	m.SetFilters(mfs)

	var val netmap.PlacementPolicy
	require.NoError(t, val.ReadFromV2(m))
	require.EqualValues(t, anyValidBackupFactor, val.ContainerBackupFactor())
	rs := val.Replicas()
	require.Len(t, rs, 2)
	require.Equal(t, "selector_0", rs[0].SelectorName())
	require.EqualValues(t, 2583748530, rs[0].NumberOfObjects())
	require.Equal(t, "selector_1", rs[1].SelectorName())
	require.EqualValues(t, 358755354, rs[1].NumberOfObjects())
	ss := val.Selectors()
	require.Len(t, ss, 2)
	require.Equal(t, "selector_0", ss[0].Name())
	require.EqualValues(t, 1814781076, ss[0].NumberOfNodes())
	require.True(t, ss[0].IsSame())
	require.Equal(t, "filter_0", ss[0].FilterName())
	require.Equal(t, "attribute_0", ss[0].BucketAttribute())
	require.Equal(t, "selector_1", ss[1].Name())
	require.EqualValues(t, 1814781076, ss[1].NumberOfNodes())
	require.True(t, ss[1].IsDistinct())
	require.Equal(t, "filter_1", ss[1].FilterName())
	require.Equal(t, "attribute_1", ss[1].BucketAttribute())
	fs := val.Filters()
	require.Len(t, fs, 2)
	require.Equal(t, "filter_0", fs[0].Name())
	require.Zero(t, fs[0].Key())
	require.Equal(t, netmap.FilterOpAND, fs[0].Op())
	require.Zero(t, fs[0].Value())
	subs := fs[0].SubFilters()
	require.Equal(t, "filter_0_0", subs[0].Name())
	require.Equal(t, "key_0_0", subs[0].Key())
	require.Equal(t, netmap.FilterOpEQ, subs[0].Op())
	require.Equal(t, "val_0_0", subs[0].Value())
	require.Empty(t, subs[0].SubFilters())
	require.Equal(t, "filter_0_1", subs[1].Name())
	require.Equal(t, "key_0_1", subs[1].Key())
	require.Equal(t, netmap.FilterOpNE, subs[1].Op())
	require.Equal(t, "val_0_1", subs[1].Value())
	require.Empty(t, subs[1].SubFilters())
	require.Equal(t, "filter_1", fs[1].Name())
	require.Zero(t, fs[1].Key())
	require.Equal(t, netmap.FilterOpOR, fs[1].Op())
	require.Zero(t, fs[1].Value())
	subs = fs[1].SubFilters()
	require.Equal(t, "filter_1_0", subs[0].Name())
	require.Equal(t, "key_1_0", subs[0].Key())
	require.Equal(t, netmap.FilterOpGT, subs[0].Op())
	require.Equal(t, "1889407708985023116", subs[0].Value())
	require.Empty(t, subs[0].SubFilters())
	require.Equal(t, "filter_1_1", subs[1].Name())
	require.Equal(t, "key_1_1", subs[1].Key())
	require.Equal(t, netmap.FilterOpGE, subs[1].Op())
	require.Equal(t, "1429243097315344888", subs[1].Value())
	require.Empty(t, subs[1].SubFilters())
	require.Equal(t, "filter_1_2", subs[2].Name())
	require.Equal(t, "key_1_2", subs[2].Key())
	require.Equal(t, netmap.FilterOpLT, subs[2].Op())
	require.Equal(t, "3722656060317482335", subs[2].Value())
	require.Empty(t, subs[2].SubFilters())
	require.Equal(t, "filter_1_3", subs[3].Name())
	require.Equal(t, "key_1_3", subs[3].Key())
	require.Equal(t, netmap.FilterOpLE, subs[3].Op())
	require.Equal(t, "1950504987705284805", subs[3].Value())
	require.Empty(t, subs[3].SubFilters())

	// reset optional fields
	m.SetSelectors(nil)
	m.SetFilters(nil)
	val2 := val
	require.NoError(t, val2.ReadFromV2(m))
	require.Empty(t, val2.Selectors())
	require.Empty(t, val2.Filters())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*apinetmap.PlacementPolicy)
		}{
			{name: "replicas/nil", err: "missing replicas",
				corrupt: func(m *apinetmap.PlacementPolicy) { m.SetReplicas(nil) }},
			{name: "replicas/empty", err: "missing replicas",
				corrupt: func(m *apinetmap.PlacementPolicy) { m.SetReplicas([]apinetmap.Replica{}) }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				var m apinetmap.PlacementPolicy
				st.WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(netmap.PlacementPolicy).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestPlacementPolicy_WriteToV2(t *testing.T) {
	var val netmap.PlacementPolicy
	var m apinetmap.PlacementPolicy

	// zero
	val.WriteToV2(&m)
	require.Zero(t, m.GetContainerBackupFactor())
	require.Zero(t, m.GetReplicas())
	require.Zero(t, m.GetSelectors())
	require.Zero(t, m.GetFilters())
	require.Zero(t, m.GetSubnetID())

	// filled
	validPlacementPolicy.WriteToV2(&m)
	require.EqualValues(t, anyValidBackupFactor, m.GetContainerBackupFactor())

	mrs := m.GetReplicas()
	require.Len(t, mrs, 2)
	require.Equal(t, "selector_0", mrs[0].GetSelector())
	require.EqualValues(t, 2583748530, mrs[0].GetCount())
	require.Equal(t, "selector_1", mrs[1].GetSelector())
	require.EqualValues(t, 358755354, mrs[1].GetCount())

	mss := m.GetSelectors()
	require.Len(t, mss, 2)
	require.Equal(t, "selector_0", mss[0].GetName())
	require.EqualValues(t, 1814781076, mss[0].GetCount())
	require.Equal(t, apinetmap.Same, mss[0].GetClause())
	require.Equal(t, "filter_0", mss[0].GetFilter())
	require.Equal(t, "attribute_0", mss[0].GetAttribute())
	require.Equal(t, "selector_1", mss[1].GetName())
	require.EqualValues(t, 1505136737, mss[1].GetCount())
	require.Equal(t, apinetmap.Distinct, mss[1].GetClause())
	require.Equal(t, "filter_1", mss[1].GetFilter())
	require.Equal(t, "attribute_1", mss[1].GetAttribute())

	mfs := m.GetFilters()
	require.Len(t, mfs, 2)
	// filter#0
	require.Equal(t, "filter_0", mfs[0].GetName())
	require.Zero(t, mfs[0].GetKey())
	require.Equal(t, apinetmap.AND, mfs[0].GetOp())
	require.Zero(t, mfs[0].GetValue())
	msubs := mfs[0].GetFilters()
	require.Len(t, msubs, 2)
	// sub#0
	require.Equal(t, "filter_0_0", msubs[0].GetName())
	require.Equal(t, "key_0_0", msubs[0].GetKey())
	require.Equal(t, apinetmap.EQ, msubs[0].GetOp())
	require.Equal(t, "val_0_0", msubs[0].GetValue())
	require.Zero(t, msubs[0].GetFilters())
	// sub#1
	require.Equal(t, "filter_0_1", msubs[1].GetName())
	require.Equal(t, "key_0_1", msubs[1].GetKey())
	require.Equal(t, apinetmap.NE, msubs[1].GetOp())
	require.Equal(t, "val_0_1", msubs[1].GetValue())
	require.Zero(t, msubs[1].GetFilters())
	// filter#1
	require.Equal(t, "filter_1", mfs[1].GetName())
	require.Zero(t, mfs[1].GetKey())
	require.Equal(t, apinetmap.OR, mfs[1].GetOp())
	require.Zero(t, mfs[1].GetValue())
	msubs = mfs[1].GetFilters()
	require.Len(t, msubs, 4)
	// sub#0
	require.Equal(t, "filter_1_0", msubs[0].GetName())
	require.Equal(t, "key_1_0", msubs[0].GetKey())
	require.Equal(t, apinetmap.GT, msubs[0].GetOp())
	require.Equal(t, "1889407708985023116", msubs[0].GetValue())
	require.Zero(t, msubs[0].GetFilters())
	// sub#1
	require.Equal(t, "filter_1_1", msubs[1].GetName())
	require.Equal(t, "key_1_1", msubs[1].GetKey())
	require.Equal(t, apinetmap.GE, msubs[1].GetOp())
	require.Equal(t, "1429243097315344888", msubs[1].GetValue())
	require.Zero(t, msubs[1].GetFilters())
	// sub#2
	require.Equal(t, "filter_1_2", msubs[2].GetName())
	require.Equal(t, "key_1_2", msubs[2].GetKey())
	require.Equal(t, apinetmap.LT, msubs[2].GetOp())
	require.Equal(t, "3722656060317482335", msubs[2].GetValue())
	require.Zero(t, msubs[2].GetFilters())
	// sub#3
	require.Equal(t, "filter_1_3", msubs[3].GetName())
	require.Equal(t, "key_1_3", msubs[3].GetKey())
	require.Equal(t, apinetmap.LE, msubs[3].GetOp())
	require.Equal(t, "1950504987705284805", msubs[3].GetValue())
	require.Zero(t, msubs[3].GetFilters())
}

func TestPlacementPolicy_Marshal(t *testing.T) {
	require.Equal(t, validBinPlacementPolicy, validPlacementPolicy.Marshal())
}

func TestPlacementPolicy_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(netmap.PlacementPolicy).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
	})

	var val netmap.PlacementPolicy
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Empty(t, val.ContainerBackupFactor())
	require.Empty(t, val.Replicas())
	require.Empty(t, val.Selectors())
	require.Empty(t, val.Filters())

	// filled
	require.NoError(t, val.Unmarshal(validBinPlacementPolicy))
	require.Equal(t, validPlacementPolicy, val)
}

func TestPlacementPolicy_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validPlacementPolicy, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONPlacementPolicy, string(b))
}

func TestPlacementPolicy_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(netmap.PlacementPolicy).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
	})

	var val netmap.PlacementPolicy
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.Zero(t, val.ContainerBackupFactor())
	require.Empty(t, val.Replicas())
	require.Empty(t, val.Selectors())
	require.Empty(t, val.Replicas())

	// filled
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONPlacementPolicy)))
	require.Equal(t, validPlacementPolicy, val)
}
