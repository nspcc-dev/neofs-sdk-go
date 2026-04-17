package stat

import (
	"sync/atomic"
	"time"
)

// methodStatus provide statistic for specific method.
type methodStatus struct {
	name        string
	allTime     atomic.Uint64
	allRequests atomic.Uint64
}

func (m *methodStatus) snapshot() Snapshot {
	return Snapshot{
		allTime:     m.allTime.Load(),
		allRequests: m.allRequests.Load(), // Technically racy wrt allTime, practically should be good enough.
	}
}

func (m *methodStatus) incRequests(elapsed time.Duration) {
	m.allRequests.Add(1)
	m.allTime.Add(uint64(elapsed))
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

	overallErrorCount atomic.Uint64
}

func (c *nodeMonitor) incErrorRate() {
	c.overallErrorCount.Add(1)
}

func (c *nodeMonitor) overallErrorRate() uint64 {
	return c.overallErrorCount.Load()
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
