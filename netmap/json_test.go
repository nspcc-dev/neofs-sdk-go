package netmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

// TestCase represents collection of placement policy tests for a single node set.
type TestCase struct {
	Name  string     `json:"name"`
	Nodes []NodeInfo `json:"nodes"`
	Tests map[string]struct {
		Policy    PlacementPolicy `json:"policy"`
		Pivot     []byte          `json:"pivot,omitempty"`
		Result    [][]int         `json:"result,omitempty"`
		Error     string          `json:"error,omitempty"`
		Placement struct {
			Pivot  []byte
			Result [][]int
		} `json:"placement"`
	}
}

var _, _ json.Unmarshaler = new(NodeInfo), new(PlacementPolicy)

func compareNodes(t testing.TB, expected [][]int, nodes nodes, actual [][]NodeInfo) {
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		require.Equal(t, len(expected[i]), len(actual[i]))
		for j, index := range expected[i] {
			require.Equal(t, nodes[index], actual[i][j])
		}
	}
}

func compareNodesIgnoreOrder(t testing.TB, expected [][]int, nodes nodes, actual [][]NodeInfo) {
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		require.Equal(t, len(expected[i]), len(actual[i]))

		var expectedNodes []NodeInfo
		for _, index := range expected[i] {
			expectedNodes = append(expectedNodes, nodes[index])
		}

		require.ElementsMatch(t, expectedNodes, actual[i])
	}
}

func TestPlacementPolicy_Interopability(t *testing.T) {
	const testsDir = "./json_tests"

	f, err := os.Open(testsDir)
	require.NoError(t, err)

	ds, err := f.ReadDir(0)
	require.NoError(t, err)

	for i := range ds {
		bs, err := os.ReadFile(filepath.Join(testsDir, ds[i].Name()))
		require.NoError(t, err)

		var tc TestCase
		require.NoError(t, json.Unmarshal(bs, &tc), "cannot unmarshal %s", ds[i].Name())

		srcNodes := slices.Clone(tc.Nodes)

		t.Run(tc.Name, func(t *testing.T) {
			var nm NetMap
			nm.SetNodes(tc.Nodes)

			for name, tt := range tc.Tests {
				t.Run(name, func(t *testing.T) {
					var pivot cid.ID
					copy(pivot[:], tt.Pivot)

					v, err := nm.ContainerNodes(tt.Policy, pivot)
					if tt.Result == nil {
						require.Error(t, err)
						require.Contains(t, err.Error(), tt.Error)
					} else {
						require.NoError(t, err)
						require.Equal(t, srcNodes, tc.Nodes)

						compareNodesIgnoreOrder(t, tt.Result, tc.Nodes, v)

						if tt.Placement.Result != nil {
							var placementPivot oid.ID
							copy(placementPivot[:], tt.Placement.Pivot)

							res, err := nm.PlacementVectors(v, placementPivot)
							require.NoError(t, err)
							compareNodes(t, tt.Placement.Result, tc.Nodes, res)
							require.Equal(t, srcNodes, tc.Nodes)
						}
					}
				})
			}
		})
	}
}

func BenchmarkPlacementPolicyInteropability(b *testing.B) {
	const testsDir = "./json_tests"

	f, err := os.Open(testsDir)
	require.NoError(b, err)

	ds, err := f.ReadDir(0)
	require.NoError(b, err)

	for i := range ds {
		bs, err := os.ReadFile(filepath.Join(testsDir, ds[i].Name()))
		require.NoError(b, err)

		var tc TestCase
		require.NoError(b, json.Unmarshal(bs, &tc), "cannot unmarshal %s", ds[i].Name())

		b.Run(tc.Name, func(b *testing.B) {
			var nm NetMap
			nm.SetNodes(tc.Nodes)
			require.NoError(b, err)

			for name, tt := range tc.Tests {
				b.Run(name, func(b *testing.B) {
					var pivot cid.ID
					copy(pivot[:], tt.Pivot)

					b.ReportAllocs()
					for b.Loop() {
						b.StartTimer()
						v, err := nm.ContainerNodes(tt.Policy, pivot)
						b.StopTimer()
						if tt.Result == nil {
							require.Error(b, err)
							require.Contains(b, err.Error(), tt.Error)
						} else {
							require.NoError(b, err)

							compareNodesIgnoreOrder(b, tt.Result, tc.Nodes, v)

							if tt.Placement.Result != nil {
								var placementPivot oid.ID
								copy(placementPivot[:], tt.Placement.Pivot)

								b.StartTimer()
								res, err := nm.PlacementVectors(v, placementPivot)
								b.StopTimer()
								require.NoError(b, err)
								compareNodes(b, tt.Placement.Result, tc.Nodes, res)
							}
						}
					}
				})
			}
		})
	}
}

func BenchmarkManySelects(b *testing.B) {
	testsFile := filepath.Join("json_tests", "many_selects.json")
	bs, err := os.ReadFile(testsFile)
	require.NoError(b, err)

	var tc TestCase
	require.NoError(b, json.Unmarshal(bs, &tc))
	tt, ok := tc.Tests["Select"]
	require.True(b, ok)

	var nm NetMap
	nm.SetNodes(tc.Nodes)

	var pivot cid.ID
	copy(pivot[:], tt.Pivot)

	b.ReportAllocs()

	for b.Loop() {
		_, err = nm.ContainerNodes(tt.Policy, pivot)
		if err != nil {
			b.FailNow()
		}
	}
}
