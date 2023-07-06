package stat

import (
	"encoding/hex"
	"sync"
	"time"
)

// PoolStat is an external statistic for pool connections.
type PoolStat struct {
	errorThreshold uint32

	mu       sync.RWMutex // protects nodeMonitor's map
	monitors map[string]*nodeMonitor
}

// NewPoolStatistic is a constructor for [PoolStat].
func NewPoolStatistic() *PoolStat {
	return &PoolStat{
		mu:       sync.RWMutex{},
		monitors: make(map[string]*nodeMonitor),
	}
}

// OperationCallback implements [stat.OperationCallback].
func (s *PoolStat) OperationCallback(nodeKey []byte, endpoint string, method Method, duration time.Duration, err error) {
	if len(nodeKey) == 0 {
		// situation when we initialize the client connection and make first EndpointInfo call.
		return
	}

	if !IsMethodValid(method) {
		return
	}

	k := hex.EncodeToString(nodeKey)

	s.mu.Lock()
	mon, ok := s.monitors[k]
	if !ok {
		methods := make([]*methodStatus, MethodLast)
		for i := MethodBalanceGet; i < MethodLast; i++ {
			methods[i] = &methodStatus{name: i.String()}
		}

		mon = &nodeMonitor{
			addr:           endpoint,
			mu:             sync.RWMutex{},
			methods:        methods,
			errorThreshold: s.errorThreshold,
		}

		s.monitors[k] = mon
	}
	s.mu.Unlock()

	if duration > 0 {
		mon.methods[method].incRequests(duration)
	}

	if err != nil {
		mon.incErrorRate()
	}
}

// Statistic returns connection statistics.
func (s *PoolStat) Statistic() Statistic {
	stat := Statistic{}

	s.mu.RLock()
	for _, mon := range s.monitors {
		node := NodeStatistic{
			publicKey:     mon.publicKey(),
			address:       mon.address(),
			methods:       mon.methodsStatus(),
			overallErrors: mon.overallErrorRate(),
		}
		stat.nodes = append(stat.nodes, node)
		stat.overallErrors += node.overallErrors
	}
	s.mu.RUnlock()

	return stat
}
