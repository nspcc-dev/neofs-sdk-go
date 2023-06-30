package stat

import (
	"sync"
	"time"
)

// methodStatus provide statistic for specific method.
type methodStatus struct {
	name string
	mu   sync.RWMutex // protect counters
	Snapshot
}

func (m *methodStatus) snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Snapshot
}

func (m *methodStatus) incRequests(elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allTime += uint64(elapsed)
	m.allRequests++
}

// Snapshot represents statistic for specific method.
type Snapshot struct {
	allTime     uint64
	allRequests uint64
}

// AllTime returns sum of time, spent to specific request. Use with [time.Duration] to get human-readable value.
func (s Snapshot) AllTime() uint64 {
	return s.allTime
}

// AllRequests returns amount of requests to node.
func (s Snapshot) AllRequests() uint64 {
	return s.allRequests
}

type nodeMonitor struct {
	pubKey         []byte
	addr           string
	errorThreshold uint32
	methods        []*methodStatus

	mu                sync.RWMutex // protect counters
	overallErrorCount uint64
}

func (c *nodeMonitor) incErrorRate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.overallErrorCount++
}

func (c *nodeMonitor) overallErrorRate() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.overallErrorCount
}

func (c *nodeMonitor) methodsStatus() []Snapshot {
	result := make([]Snapshot, len(c.methods))
	for i, val := range c.methods {
		result[i] = val.snapshot()
	}

	return result
}

func (c *nodeMonitor) address() string {
	return c.addr
}

func (c *nodeMonitor) publicKey() []byte {
	return c.pubKey
}
