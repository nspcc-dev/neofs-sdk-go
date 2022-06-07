package netmap

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
		} `json:"placement,omitempty"`
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

func TestPlacementPolicy_Interopability(t *testing.T) {
	const testsDir = "./json_tests"

	f, err := os.Open(testsDir)
	require.NoError(t, err)

	ds, err := f.ReadDir(0)
	require.NoError(t, err)

	for i := range ds {
		bs, err := ioutil.ReadFile(filepath.Join(testsDir, ds[i].Name()))
		require.NoError(t, err)

		var tc TestCase
		require.NoError(t, json.Unmarshal(bs, &tc), "cannot unmarshal %s", ds[i].Name())

		srcNodes := make([]NodeInfo, len(tc.Nodes))
		copy(srcNodes, tc.Nodes)

		t.Run(tc.Name, func(t *testing.T) {
			var nm NetMap
			nm.SetNodes(tc.Nodes)

			for name, tt := range tc.Tests {
				t.Run(name, func(t *testing.T) {
					v, err := nm.ContainerNodes(tt.Policy, tt.Pivot)
					if tt.Result == nil {
						require.Error(t, err)
						require.Contains(t, err.Error(), tt.Error)
					} else {
						require.NoError(t, err)
						require.Equal(t, srcNodes, tc.Nodes)

						compareNodes(t, tt.Result, tc.Nodes, v)

						if tt.Placement.Result != nil {
							res, err := nm.PlacementVectors(v, tt.Placement.Pivot)
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
		bs, err := ioutil.ReadFile(filepath.Join(testsDir, ds[i].Name()))
		require.NoError(b, err)

		var tc TestCase
		require.NoError(b, json.Unmarshal(bs, &tc), "cannot unmarshal %s", ds[i].Name())

		b.Run(tc.Name, func(b *testing.B) {
			var nm NetMap
			nm.SetNodes(tc.Nodes)
			require.NoError(b, err)

			for name, tt := range tc.Tests {
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						b.StartTimer()
						v, err := nm.ContainerNodes(tt.Policy, tt.Pivot)
						b.StopTimer()
						if tt.Result == nil {
							require.Error(b, err)
							require.Contains(b, err.Error(), tt.Error)
						} else {
							require.NoError(b, err)

							compareNodes(b, tt.Result, tc.Nodes, v)

							if tt.Placement.Result != nil {
								b.StartTimer()
								res, err := nm.PlacementVectors(v, tt.Placement.Pivot)
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
	bs, err := ioutil.ReadFile(testsFile)
	require.NoError(b, err)

	var tc TestCase
	require.NoError(b, json.Unmarshal(bs, &tc))
	tt, ok := tc.Tests["Select"]
	require.True(b, ok)

	var nm NetMap
	nm.SetNodes(tc.Nodes)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err = nm.ContainerNodes(tt.Policy, tt.Pivot)
		if err != nil {
			b.FailNow()
		}
	}
}
