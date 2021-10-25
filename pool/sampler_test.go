package pool

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/session"
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

func newNetmapMock(name string, needErr bool) clientMock {
	var err error
	if needErr {
		err = fmt.Errorf("not available")
	}
	return clientMock{name: name, err: err}
}

func (n clientMock) EndpointInfo(_ context.Context, _ ...client.CallOption) (*client.EndpointInfo, error) {
	return nil, n.err
}

func (n clientMock) CreateSession(_ context.Context, _ uint64, _ ...client.CallOption) (*session.Token, error) {
	return session.NewToken(), n.err
}

func TestHealthyReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		options = &BuilderOptions{weights: weights}
		buffer  = make([]float64, len(weights))
	)

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	p := &pool{
		sampler: NewSampler(weights, rand.NewSource(0)),
		clientPacks: []*clientPack{
			{client: newNetmapMock(names[0], true), healthy: true, address: "address0"},
			{client: newNetmapMock(names[1], false), healthy: true, address: "address1"}},
		cache: NewCache(),
		key:   &key.PrivateKey,
	}

	// check getting first node connection before rebalance happened
	connection0, _, err := p.Connection()
	require.NoError(t, err)
	mock0 := connection0.(clientMock)
	require.Equal(t, names[0], mock0.name)

	updateNodesHealth(context.TODO(), p, options, buffer)

	connection1, _, err := p.Connection()
	require.NoError(t, err)
	mock1 := connection1.(clientMock)
	require.Equal(t, names[1], mock1.name)

	// enabled first node again
	p.lock.Lock()
	p.clientPacks[0].client = newNetmapMock(names[0], false)
	p.lock.Unlock()

	updateNodesHealth(context.TODO(), p, options, buffer)
	p.sampler = NewSampler(weights, rand.NewSource(0))

	connection0, _, err = p.Connection()
	require.NoError(t, err)
	mock0 = connection0.(clientMock)
	require.Equal(t, names[0], mock0.name)
}

func TestHealthyNoReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		options = &BuilderOptions{weights: weights}
		buffer  = make([]float64, len(weights))
	)

	sampler := NewSampler(weights, rand.NewSource(0))
	p := &pool{
		sampler: sampler,
		clientPacks: []*clientPack{
			{client: newNetmapMock(names[0], false), healthy: true},
			{client: newNetmapMock(names[1], false), healthy: true}},
	}

	updateNodesHealth(context.TODO(), p, options, buffer)

	p.lock.RLock()
	defer p.lock.RUnlock()
	require.Truef(t, sampler == p.sampler, "Sampler must not be changed. Expected: %p, actual: %p", sampler, p.sampler)
}
