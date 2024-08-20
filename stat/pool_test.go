package stat

import (
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStatisticConcurrency(t *testing.T) {
	ps := NewPoolStatistic()
	wg := sync.WaitGroup{}

	type config struct {
		method  Method
		modeKey []byte
		addr    string
		errors  *atomic.Int64
	}

	addrNode1 := "node1"
	addrNode2 := "node2"
	addrNode3 := "node3"

	counterNode1 := &atomic.Int64{}
	counterNode2 := &atomic.Int64{}
	counterNode3 := &atomic.Int64{}

	configs := []*config{
		{method: MethodContainerDelete, modeKey: []byte{1}, addr: addrNode1, errors: counterNode1},
		{method: MethodContainerDelete, modeKey: []byte{2}, addr: addrNode2, errors: counterNode2},
		{method: MethodContainerList, modeKey: []byte{1}, addr: addrNode1, errors: counterNode1},
		{method: MethodContainerList, modeKey: []byte{2}, addr: addrNode2, errors: counterNode2},
		{method: MethodObjectRange, modeKey: []byte{3}, addr: addrNode3, errors: counterNode3},
		{method: MethodLast - 1, modeKey: []byte{3}, addr: addrNode3, errors: counterNode3},
	}

	wg.Add(len(configs))
	n := 30000

	for _, s := range configs {
		go func(c *config) {
			defer wg.Done()

			for i := 0; i < n; i++ {
				var err error
				if rand.N(2) > 0 {
					err = errors.New("some err")
					c.errors.Add(1)
				}

				duration := time.Duration(rand.N(200)+1) * time.Millisecond

				ps.OperationCallback(c.modeKey, c.addr, c.method, duration, err)
			}
		}(s)
	}

	wg.Wait()

	for _, s := range configs {
		node, err := ps.Statistic().Node(s.addr)
		require.NoError(t, err)

		require.Equal(t, uint64(s.errors.Load()), node.OverallErrors())
		require.Equal(t, uint64(n*2), node.Requests(), s.addr)
	}
}
