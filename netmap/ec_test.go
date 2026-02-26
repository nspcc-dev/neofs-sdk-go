package netmap_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

// set by init.
var (
	anyValidECRules = make([]netmap.ECRule, 2)
	validECPolicy   netmap.PlacementPolicy
)

// corresponds to validECPolicy.
var validBinECPolicy = []byte{
	50, 11, 8, 1, 16, 2, 26, 5, 115, 108, 99, 116, 114, 50, 4, 8, 31, 16, 32,
}

var validJSONECPolicy = `
{
 "replicas": [],
 "containerBackupFactor": 0,
 "selectors": [],
 "filters": [],
 "subnetId": null,
 "ecRules": [
  {
   "dataPartNum": 1,
   "parityPartNum": 2,
   "selector": "slctr"
  },
  {
   "dataPartNum": 31,
   "parityPartNum": 32,
   "selector": ""
  }
 ],
 "initial": null
}
`

func init() {
	anyValidECRules[0].SetDataPartNum(1)
	anyValidECRules[0].SetParityPartNum(2)
	anyValidECRules[0].SetSelectorName("slctr")
	anyValidECRules[1].SetDataPartNum(31)
	anyValidECRules[1].SetParityPartNum(32)

	validECPolicy.SetECRules(anyValidECRules)
	validECPolicy.SetReplicas([]netmap.ReplicaDescriptor{})
	validECPolicy.SetFilters([]netmap.Filter{})
	validECPolicy.SetSelectors([]netmap.Selector{})
}

func TestNewECRule(t *testing.T) {
	r := netmap.NewECRule(12, 34)
	require.EqualValues(t, 12, r.DataPartNum())
	require.EqualValues(t, 34, r.ParityPartNum())
	require.Zero(t, r.SelectorName())
}

func TestECRule_SetDataPartNum(t *testing.T) {
	var r netmap.ECRule
	require.Zero(t, r.DataPartNum())

	r.SetDataPartNum(123)
	require.EqualValues(t, 123, r.DataPartNum())
}

func TestECRule_SetParityPartNum(t *testing.T) {
	var r netmap.ECRule
	require.Zero(t, r.ParityPartNum())

	r.SetParityPartNum(123)
	require.EqualValues(t, 123, r.ParityPartNum())
}

func TestECRule_SetSelector(t *testing.T) {
	var r netmap.ECRule
	require.Zero(t, r.SelectorName())

	r.SetSelectorName("slctr")
	require.EqualValues(t, "slctr", r.SelectorName())
}

func TestPlacementPolicy_ECRules(t *testing.T) {
	var p netmap.PlacementPolicy
	require.Zero(t, p.ECRules())

	rs := []netmap.ECRule{
		netmap.NewECRule(10, 11),
		netmap.NewECRule(20, 21),
	}
	p.SetECRules(rs)
	require.Equal(t, rs, p.ECRules())
}

func TestPlacementPolicy_WriteStringTo_EC(t *testing.T) {
	var p netmap.PlacementPolicy

	check := func(t *testing.T, exp string) {
		var b strings.Builder
		require.NoError(t, p.WriteStringTo(&b))
		require.Equal(t, exp, b.String())

		var p2 netmap.PlacementPolicy
		require.NoError(t, p2.DecodeString(exp))
		require.Equal(t, p.Marshal(), p2.Marshal())
	}

	rs := []netmap.ECRule{
		netmap.NewECRule(10, 11),
		netmap.NewECRule(20, 21),
	}

	p.SetECRules(rs)
	check(t, `EC 10/11
EC 20/21`)

	rs[1].SetSelectorName("slctr")
	p.SetECRules(rs)
	check(t, `EC 10/11
EC 20/21 IN slctr`)
}

func testPolicyDecodeStringEC(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{
			"ec", "eC", "Ec",
			"EC", "EC 1", "EC a", "EC 1 2", "EC 1,2",
			"EC /", "EC /2", "EC 1/", "EC a/2", "EC 1/a",
			"EC 1.5/2", "EC 1/2.5",
		} {
			t.Run(s, func(t *testing.T) {
				require.Error(t, new(netmap.PlacementPolicy).DecodeString(s))
				require.Error(t, new(netmap.PlacementPolicy).DecodeString("EC 1/2\n"+s))
			})
		}
	})

	for _, tc := range []struct {
		s   string
		cmp func(t *testing.T, p netmap.PlacementPolicy)
	}{
		{s: "EC 1/2", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			require.Equal(t, []netmap.ECRule{
				netmap.NewECRule(1, 2),
			}, p.ECRules())
			require.Zero(t, p.ContainerBackupFactor())
			require.Empty(t, p.Selectors())
			require.Empty(t, p.Filters())
			require.Zero(t, p.Replicas())
		}},
		{s: "EC 1/2\nEC 3/4", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			require.Equal(t, []netmap.ECRule{
				netmap.NewECRule(1, 2),
				netmap.NewECRule(3, 4),
			}, p.ECRules())
			require.Zero(t, p.ContainerBackupFactor())
			require.Empty(t, p.Selectors())
			require.Empty(t, p.Filters())
			require.Zero(t, p.Replicas())
		}},
		{s: "EC 1/2 CBF 1", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			require.Equal(t, []netmap.ECRule{
				netmap.NewECRule(1, 2),
			}, p.ECRules())
			require.EqualValues(t, 1, p.ContainerBackupFactor())
			require.Empty(t, p.Selectors())
			require.Empty(t, p.Filters())
			require.Zero(t, p.Replicas())
		}},
		{s: "EC 1/2 IN slctr", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			rs := p.ECRules()
			require.Len(t, rs, 1)
			require.EqualValues(t, 1, rs[0].DataPartNum())
			require.EqualValues(t, 2, rs[0].ParityPartNum())
			require.Zero(t, p.ContainerBackupFactor())
			require.Empty(t, p.Selectors())
			require.Empty(t, p.Filters())
			require.Zero(t, p.Replicas())
		}},
		{s: "EC 1/2 IN slctr CBF 1 SELECT 4 FROM * AS slctr", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			rs := p.ECRules()
			require.Len(t, rs, 1)
			require.EqualValues(t, 1, rs[0].DataPartNum())
			require.EqualValues(t, 2, rs[0].ParityPartNum())
			require.Equal(t, "slctr", rs[0].SelectorName())
			require.EqualValues(t, 1, p.ContainerBackupFactor())
			ss := p.Selectors()
			require.Len(t, ss, 1)
			require.Equal(t, "slctr", ss[0].Name())
			require.EqualValues(t, 4, ss[0].NumberOfNodes())
			require.Equal(t, "*", ss[0].FilterName())
			require.Empty(t, p.Filters())
			require.Zero(t, p.Replicas())
		}},
		{s: "EC 1/2 IN slctr EC 3/4 CBF 5 SELECT 4 IN DISTINCT attrs FROM fltr AS slctr FILTER key GT val AS fltr", cmp: func(t *testing.T, p netmap.PlacementPolicy) {
			rs := p.ECRules()
			require.Len(t, rs, 2)
			require.EqualValues(t, 1, rs[0].DataPartNum())
			require.EqualValues(t, 2, rs[0].ParityPartNum())
			require.Equal(t, "slctr", rs[0].SelectorName())
			require.Equal(t, netmap.NewECRule(3, 4), rs[1])
			require.EqualValues(t, 5, p.ContainerBackupFactor())
			ss := p.Selectors()
			require.Len(t, ss, 1)
			require.Equal(t, "slctr", ss[0].Name())
			require.EqualValues(t, 4, ss[0].NumberOfNodes())
			require.Equal(t, "attrs", ss[0].BucketAttribute())
			require.True(t, ss[0].IsDistinct())
			require.Equal(t, "fltr", ss[0].FilterName())
			fs := p.Filters()
			require.Len(t, fs, 1)
			require.Equal(t, "fltr", fs[0].Name())
			require.Equal(t, "key", fs[0].Key())
			require.Equal(t, netmap.FilterOpGT, fs[0].Op())
			require.Equal(t, "val", fs[0].Value())
			require.Zero(t, p.Replicas())
		}},
	} {
		t.Run(tc.s, func(t *testing.T) {
			var p netmap.PlacementPolicy
			require.NoError(t, p.DecodeString(tc.s))
			tc.cmp(t, p)
		})
	}
}

func testPolicyProtoMessageEC(t *testing.T) {
	m := validECPolicy.ProtoMessage()

	rs := m.EcRules
	require.Len(t, rs, 2)

	require.EqualValues(t, 1, rs[0].DataPartNum)
	require.EqualValues(t, 2, rs[0].ParityPartNum)
	require.EqualValues(t, "slctr", rs[0].Selector)

	require.EqualValues(t, 31, rs[1].DataPartNum)
	require.EqualValues(t, 32, rs[1].ParityPartNum)
	require.Zero(t, rs[1].Selector)
}

func testPolicyMarshalEC(t *testing.T) {
	require.Equal(t, validBinECPolicy, validECPolicy.Marshal())
}

func testPolicyUnmarshalEC(t *testing.T) {
	var val netmap.PlacementPolicy
	require.NoError(t, val.Unmarshal(validBinECPolicy))
	require.Equal(t, validECPolicy, val)
}

func testPolicyMarshalJSONEC(t *testing.T) {
	b, err := json.MarshalIndent(validECPolicy, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONECPolicy, string(b))
}

func testPolicyUnmarshalJSONEC(t *testing.T) {
	var val netmap.PlacementPolicy
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONECPolicy)))
	require.Equal(t, validECPolicy, val)
}

func testPolicyVerifyEC(t *testing.T) {
	var policy netmap.PlacementPolicy

	validRs, validSs := make([]netmap.ECRule, 2), make([]netmap.Selector, 2)
	validRs[0] = netmap.NewECRule(20, 10)
	validRs[1] = netmap.NewECRule(10, 20)
	validRs[0].SetSelectorName("SEL1")
	validSs[0].SetName("SEL1")
	validRs[1].SetSelectorName("SEL2")
	validSs[1].SetName("SEL2")

	policy.SetECRules(validRs)
	policy.SetSelectors(validSs)
	require.NoError(t, policy.Verify())

	for _, tc := range []struct {
		name    string
		err     string
		corrupt func(p *netmap.PlacementPolicy)
	}{
		{name: "too many rules", err: "more than 4 EC rules", corrupt: func(p *netmap.PlacementPolicy) {
			rules := make([]netmap.ECRule, 5)
			for i := range rules {
				rules[i].SetDataPartNum(3)
				rules[i].SetParityPartNum(uint32(1 + i))
			}
			p.SetECRules(rules)
		}},
		{name: "zero data parts", err: "invalid EC rule #1: zero data part num", corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetDataPartNum(0)
		}},
		{name: "zero parity parts", err: "invalid EC rule #1: zero parity part num", corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetParityPartNum(0)
		}},
		{name: "too many data parts", err: "invalid EC rule #1: more than 64 total parts", corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetDataPartNum(65)
		}},
		{name: "too many parity parts", err: "invalid EC rule #1: more than 64 total parts", corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetDataPartNum(65)
		}},
		{name: "too many total parts", err: "invalid EC rule #1: more than 64 total parts", corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetDataPartNum(32)
			p.ECRules()[1].SetParityPartNum(33)
		}},
		{name: "missing selector", err: `invalid EC rule #1: missing selector "SEL3"`, corrupt: func(p *netmap.PlacementPolicy) {
			p.ECRules()[1].SetSelectorName("SEL3")
		}},
		{name: "too many nodes for rule", err: `invalid EC rule #1: more than 64 nodes`, corrupt: func(p *netmap.PlacementPolicy) {
			p.Selectors()[1].SetNumberOfNodes(22)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cp netmap.PlacementPolicy
			policy.CopyTo(&cp)
			tc.corrupt(&cp)
			require.EqualError(t, cp.Verify(), tc.err)
		})
	}
}
