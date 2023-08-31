package stat

import (
	"errors"
	"time"
)

// Statistic is metrics of the pool.
type Statistic struct {
	overallErrors uint64
	nodes         []NodeStatistic
}

// OverallErrors returns sum of errors on all connections. It doesn't decrease.
func (s Statistic) OverallErrors() uint64 {
	return s.overallErrors
}

// Nodes returns list of nodes statistic.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (s Statistic) Nodes() []NodeStatistic {
	return s.nodes
}

// ErrUnknownNode indicate that node with current address is not found in list.
var ErrUnknownNode = errors.New("unknown node")

// Node returns NodeStatistic by node address.
// If such node doesn't exist ErrUnknownNode error is returned.
func (s Statistic) Node(address string) (*NodeStatistic, error) {
	for i := range s.nodes {
		if s.nodes[i].address == address {
			return &s.nodes[i], nil
		}
	}

	return nil, ErrUnknownNode
}

// NodeStatistic is metrics of certain connections.
type NodeStatistic struct {
	publicKey     []byte
	address       string
	methods       []Snapshot
	overallErrors uint64
}

// Snapshot returns snapshot statistic for method.
func (n NodeStatistic) Snapshot(method Method) (Snapshot, error) {
	if !IsMethodValid(method) {
		return Snapshot{}, errors.New("invalid method")
	}

	return n.methods[method], nil
}

// OverallErrors returns all errors on current node.
// This value never decreases.
func (n NodeStatistic) OverallErrors() uint64 {
	return n.overallErrors
}

// Requests returns number of requests.
func (n NodeStatistic) Requests() (requests uint64) {
	for _, val := range n.methods {
		requests += val.allRequests
	}
	return requests
}

// Address returns node endpoint address.
func (n NodeStatistic) Address() string {
	return n.address
}

// AverageGetBalance returns average time to perform BalanceGet request.
func (n NodeStatistic) AverageGetBalance() time.Duration {
	return n.averageTime(MethodBalanceGet)
}

// AveragePutContainer returns average time to perform ContainerPut request.
func (n NodeStatistic) AveragePutContainer() time.Duration {
	return n.averageTime(MethodContainerPut)
}

// AverageGetContainer returns average time to perform ContainerGet request.
func (n NodeStatistic) AverageGetContainer() time.Duration {
	return n.averageTime(MethodContainerGet)
}

// AverageListContainer returns average time to perform ContainerList request.
func (n NodeStatistic) AverageListContainer() time.Duration {
	return n.averageTime(MethodContainerList)
}

// AverageDeleteContainer returns average time to perform ContainerDelete request.
func (n NodeStatistic) AverageDeleteContainer() time.Duration {
	return n.averageTime(MethodContainerDelete)
}

// AverageGetContainerEACL returns average time to perform ContainerEACL request.
func (n NodeStatistic) AverageGetContainerEACL() time.Duration {
	return n.averageTime(MethodContainerEACL)
}

// AverageSetContainerEACL returns average time to perform ContainerSetEACL request.
func (n NodeStatistic) AverageSetContainerEACL() time.Duration {
	return n.averageTime(MethodContainerSetEACL)
}

// AverageEndpointInfo returns average time to perform EndpointInfo request.
func (n NodeStatistic) AverageEndpointInfo() time.Duration {
	return n.averageTime(MethodEndpointInfo)
}

// AverageNetworkInfo returns average time to perform NetworkInfo request.
func (n NodeStatistic) AverageNetworkInfo() time.Duration {
	return n.averageTime(MethodNetworkInfo)
}

// AveragePutObject returns average time to perform ObjectPut request.
func (n NodeStatistic) AveragePutObject() time.Duration {
	return n.averageTime(MethodObjectPut)
}

// AverageDeleteObject returns average time to perform ObjectDelete request.
func (n NodeStatistic) AverageDeleteObject() time.Duration {
	return n.averageTime(MethodObjectDelete)
}

// AverageGetObject returns average time to perform ObjectGet request.
func (n NodeStatistic) AverageGetObject() time.Duration {
	return n.averageTime(MethodObjectGet)
}

// AverageHeadObject returns average time to perform ObjectHead request.
func (n NodeStatistic) AverageHeadObject() time.Duration {
	return n.averageTime(MethodObjectHead)
}

// AverageRangeObject returns average time to perform ObjectRange request.
func (n NodeStatistic) AverageRangeObject() time.Duration {
	return n.averageTime(MethodObjectRange)
}

// AverageCreateSession returns average time to perform SessionCreate request.
func (n NodeStatistic) AverageCreateSession() time.Duration {
	return n.averageTime(MethodSessionCreate)
}

func (n NodeStatistic) averageTime(method Method) time.Duration {
	stat := n.methods[method]
	if stat.allRequests == 0 {
		return 0
	}
	return time.Duration(stat.allTime / stat.allRequests)
}
