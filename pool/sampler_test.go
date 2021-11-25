package pool

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/stretchr/testify/require"
)

func TestSamplerStability(t *testing.T) {
	const COUNT = 100000

	cases := []struct {
		probabilities []float64
		expected      []int
	}{
		{
			probabilities: []float64{1, 0},
			expected:      []int{COUNT, 0},
		},
		{
			probabilities: []float64{0.1, 0.2, 0.7},
			expected:      []int{10138, 19813, 70049},
		},
		{
			probabilities: []float64{0.2, 0.2, 0.4, 0.1, 0.1, 0},
			expected:      []int{19824, 20169, 39900, 10243, 9864, 0},
		},
	}

	for _, tc := range cases {
		sampler := NewSampler(tc.probabilities, rand.NewSource(0))
		res := make([]int, len(tc.probabilities))
		for i := 0; i < COUNT; i++ {
			res[sampler.Next()]++
		}

		require.Equal(t, tc.expected, res, "probabilities: %v", tc.probabilities)
	}
}

type clientMock struct {
	client.Client
	name string
	err  error
}

func (c *clientMock) EndpointInfo(context.Context, ...client.CallOption) (*client.EndpointInfoRes, error) {
	return nil, nil
}

func (c *clientMock) NetworkInfo(context.Context, ...client.CallOption) (*client.NetworkInfoRes, error) {
	return nil, nil
}

func newNetmapMock(name string, needErr bool) *clientMock {
	var err error
	if needErr {
		err = fmt.Errorf("not available")
	}
	return &clientMock{name: name, err: err}
}

func TestHealthyReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		options = &BuilderOptions{nodesParams: []*NodesParam{{weights: weights}}}
		buffer  = make([]float64, len(weights))
	)

	cache, err := NewCache()
	require.NoError(t, err)

	inner := &innerPool{
		sampler: NewSampler(weights, rand.NewSource(0)),
		clientPacks: []*clientPack{
			{client: newNetmapMock(names[0], true), healthy: true, address: "address0"},
			{client: newNetmapMock(names[1], false), healthy: true, address: "address1"}},
	}
	p := &pool{
		innerPools: []*innerPool{inner},
		cache:      cache,
		key:        newPrivateKey(t),
	}

	// check getting first node connection before rebalance happened
	connection0, _, err := p.Connection()
	require.NoError(t, err)
	mock0 := connection0.(*clientMock)
	require.Equal(t, names[0], mock0.name)

	updateInnerNodesHealth(context.TODO(), p, 0, options, buffer)

	connection1, _, err := p.Connection()
	require.NoError(t, err)
	mock1 := connection1.(*clientMock)
	require.Equal(t, names[1], mock1.name)

	// enabled first node again
	inner.lock.Lock()
	inner.clientPacks[0].client = newNetmapMock(names[0], false)
	inner.lock.Unlock()

	updateInnerNodesHealth(context.TODO(), p, 0, options, buffer)
	inner.sampler = NewSampler(weights, rand.NewSource(0))

	connection0, _, err = p.Connection()
	require.NoError(t, err)
	mock0 = connection0.(*clientMock)
	require.Equal(t, names[0], mock0.name)
}

func TestHealthyNoReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		options = &BuilderOptions{nodesParams: []*NodesParam{{weights: weights}}}
		buffer  = make([]float64, len(weights))
	)

	sampler := NewSampler(weights, rand.NewSource(0))
	inner := &innerPool{
		sampler: sampler,
		clientPacks: []*clientPack{
			{client: newNetmapMock(names[0], false), healthy: true},
			{client: newNetmapMock(names[1], false), healthy: true}},
	}
	p := &pool{
		innerPools: []*innerPool{inner},
	}

	updateInnerNodesHealth(context.TODO(), p, 0, options, buffer)

	inner.lock.RLock()
	defer inner.lock.RUnlock()
	require.Equal(t, inner.sampler, sampler)
}
