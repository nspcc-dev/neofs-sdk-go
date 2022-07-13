package pool

import (
	"errors"
	"time"
)

// Statistic is metrics of the pool.
type Statistic struct {
	overallErrors uint64
	nodes         []*NodeStatistic
}

// OverallErrors returns sum of errors on all connections. It doesn't decrease.
func (s Statistic) OverallErrors() uint64 {
	return s.overallErrors
}

// Nodes returns list of nodes statistic.
func (s Statistic) Nodes() []*NodeStatistic {
	return s.nodes
}

// ErrUnknownNode indicate that node with current address is not found in list.
var ErrUnknownNode = errors.New("unknown node")

// Node returns NodeStatistic by node address.
// If such node doesn't exist ErrUnknownNode error is returned.
func (s Statistic) Node(address string) (*NodeStatistic, error) {
	for i := range s.nodes {
		if s.nodes[i].address == address {
			return s.nodes[i], nil
		}
	}

	return nil, ErrUnknownNode
}

// NodeStatistic is metrics of certain connections.
type NodeStatistic struct {
	address       string
	latency       time.Duration
	requests      uint64
	overallErrors uint64
	currentErrors uint32
}

// OverallErrors returns all errors on current node.
// This value never decreases.
func (n NodeStatistic) OverallErrors() uint64 {
	return n.overallErrors
}

// CurrentErrors returns errors on current node.
// This value is always less than 'errorThreshold' from InitParameters.
func (n NodeStatistic) CurrentErrors() uint32 {
	return n.currentErrors
}

// Latency returns average latency for node request.
func (n NodeStatistic) Latency() time.Duration {
	return n.latency
}

// Requests returns number of requests.
func (n NodeStatistic) Requests() uint64 {
	return n.requests
}

// Address returns node endpoint address.
func (n NodeStatistic) Address() string {
	return n.address
}
