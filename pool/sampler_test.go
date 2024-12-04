package pool

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"

	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
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
		sampler := newSampler(tc.probabilities, rand.NewSource(0))
		res := make([]int, len(tc.probabilities))
		for range COUNT {
			res[sampler.next()]++
		}

		require.Equal(t, tc.expected, res, "probabilities: %v", tc.probabilities)
	}
}

func TestHealthyReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		buffer  = make([]float64, len(weights))
	)

	cache, err := newCache(defaultSessionCacheSize)
	require.NoError(t, err)

	client1 := newMockClient(names[0], neofscryptotest.Signer())
	client1.errOnDial()

	client2 := newMockClient(names[1], neofscryptotest.Signer())

	inner := &innerPool{
		sampler: newSampler(weights, rand.NewSource(0)),
		clients: []internalClient{client1, client2},
	}
	p := &Pool{
		innerPools:      []*innerPool{inner},
		cache:           cache,
		signer:          usertest.User(),
		rebalanceParams: rebalanceParameters{nodesParams: []*nodesParam{{weights: weights}}},
	}

	// check getting first node connection before rebalance happened
	connection0, err := p.connection()
	require.NoError(t, err)
	mock0 := connection0.(*mockClient)
	require.Equal(t, names[0], mock0.address())

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)

	connection1, err := p.connection()
	require.NoError(t, err)
	mock1 := connection1.(*mockClient)
	require.Equal(t, names[1], mock1.address())

	// enabled first node again
	inner.lock.Lock()
	inner.clients[0] = newMockClient(names[0], neofscryptotest.Signer())
	inner.lock.Unlock()

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)
	inner.sampler = newSampler(weights, rand.NewSource(0))

	connection0, err = p.connection()
	require.NoError(t, err)
	mock0 = connection0.(*mockClient)
	require.Equal(t, names[0], mock0.address())
}

func TestHealthyNoReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		buffer  = make([]float64, len(weights))
	)

	sampl := newSampler(weights, rand.NewSource(0))
	inner := &innerPool{
		sampler: sampl,
		clients: []internalClient{
			newMockClient(names[0], neofscryptotest.Signer()),
			newMockClient(names[1], neofscryptotest.Signer()),
		},
	}
	p := &Pool{
		innerPools:      []*innerPool{inner},
		rebalanceParams: rebalanceParameters{nodesParams: []*nodesParam{{weights: weights}}},
	}

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)

	inner.lock.RLock()
	defer inner.lock.RUnlock()
	require.Equal(t, inner.sampler, sampl)
}

func TestSamplerSafety(t *testing.T) {
	// https://github.com/nspcc-dev/neofs-sdk-go/issues/631
	// Note that this test is not 100% consistent so it may PASS, but it FAILs more
	// often when bugs.
	type panicInfo = struct {
		cause any
		stack []byte
	}
	s := newSampler([]float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}, rand.NewSource(0))
	var pr atomic.Value // panicInfo
	var wg sync.WaitGroup
	for range 1000 {
		const rn = 1000
		wg.Add(rn)
		for range rn {
			go func() {
				defer func() {
					wg.Done()
					// [require.NotPanics] should be called in a test func routine, so we simulate it
					if r := recover(); r != nil {
						// in theory, various causes may happen. With this, only the "last" one is
						// caught. In practice, we are chasing the exact one.
						pr.Store(panicInfo{r, debug.Stack()})
					}
				}()
				s.next()
			}()
		}
		wg.Wait()
		if v := pr.Load(); v != nil {
			p := v.(panicInfo)
			require.Fail(t, fmt.Sprintf("should not panic\n\tPanic value:\t%v\n\tPanic stack:\t%s", p.cause, p.stack))
		}
	}
}
