package pool

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
	address       string
	methods       map[string]methodStatus
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
	return n.averageTime(methodBalanceGet)
}

// AveragePutContainer returns average time to perform ContainerPut request.
func (n NodeStatistic) AveragePutContainer() time.Duration {
	return n.averageTime(methodContainerPut)
}

// AverageGetContainer returns average time to perform ContainerGet request.
func (n NodeStatistic) AverageGetContainer() time.Duration {
	return n.averageTime(methodContainerGet)
}

// AverageListContainer returns average time to perform ContainerList request.
func (n NodeStatistic) AverageListContainer() time.Duration {
	return n.averageTime(methodContainerList)
}

// AverageDeleteContainer returns average time to perform ContainerDelete request.
func (n NodeStatistic) AverageDeleteContainer() time.Duration {
	return n.averageTime(methodContainerDelete)
}

// AverageGetContainerEACL returns average time to perform ContainerEACL request.
func (n NodeStatistic) AverageGetContainerEACL() time.Duration {
	return n.averageTime(methodContainerEACL)
}

// AverageSetContainerEACL returns average time to perform ContainerSetEACL request.
func (n NodeStatistic) AverageSetContainerEACL() time.Duration {
	return n.averageTime(methodContainerSetEACL)
}

// AverageEndpointInfo returns average time to perform EndpointInfo request.
func (n NodeStatistic) AverageEndpointInfo() time.Duration {
	return n.averageTime(methodEndpointInfo)
}

// AverageNetworkInfo returns average time to perform NetworkInfo request.
func (n NodeStatistic) AverageNetworkInfo() time.Duration {
	return n.averageTime(methodNetworkInfo)
}

// AveragePutObject returns average time to perform ObjectPut request.
func (n NodeStatistic) AveragePutObject() time.Duration {
	return n.averageTime(methodObjectPut)
}

// AverageDeleteObject returns average time to perform ObjectDelete request.
func (n NodeStatistic) AverageDeleteObject() time.Duration {
	return n.averageTime(methodObjectDelete)
}

// AverageGetObject returns average time to perform ObjectGet request.
func (n NodeStatistic) AverageGetObject() time.Duration {
	return n.averageTime(methodObjectGet)
}

// AverageHeadObject returns average time to perform ObjectHead request.
func (n NodeStatistic) AverageHeadObject() time.Duration {
	return n.averageTime(methodObjectHead)
}

// AverageRangeObject returns average time to perform ObjectRange request.
func (n NodeStatistic) AverageRangeObject() time.Duration {
	return n.averageTime(methodObjectRange)
}

// AverageCreateSession returns average time to perform SessionCreate request.
func (n NodeStatistic) AverageCreateSession() time.Duration {
	return n.averageTime(methodSessionCreate)
}

func (n NodeStatistic) averageTime(method string) time.Duration {
	stat := n.methods[method]
	if stat.allRequests == 0 {
		return 0
	}
	return time.Duration(stat.allTime / stat.allRequests)
}
