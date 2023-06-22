/*
Package provides functionality for collecting/maintaining client statistics.
*/

package stat

import (
	"time"
)

// Method is an enumerator to describe [client.Client] methods.
type Method int

// Various client methods.
const (
	MethodBalanceGet Method = iota
	MethodContainerPut
	MethodContainerGet
	MethodContainerList
	MethodContainerDelete
	MethodContainerEACL
	MethodContainerSetEACL
	MethodEndpointInfo
	MethodNetworkInfo
	MethodObjectPut
	MethodObjectDelete
	MethodObjectGet
	MethodObjectHead
	MethodObjectRange
	MethodSessionCreate
	MethodNetMapSnapshot
	MethodObjectHash
	MethodObjectSearch
	MethodContainerAnnounceUsedSpace
	MethodAnnounceIntermediateTrust
	MethodAnnounceLocalTrust
	MethodObjectGetStream
	MethodObjectRangeStream
	MethodObjectSearchStream
	MethodObjectPutStream
	// MethodLast is no a valid method name, it's a system anchor for tests, etc.
	MethodLast
)

// String implements fmt.Stringer.
func (m Method) String() string {
	switch m {
	case MethodBalanceGet:
		return "balanceGet"
	case MethodContainerPut:
		return "containerPut"
	case MethodContainerGet:
		return "containerGet"
	case MethodContainerList:
		return "containerList"
	case MethodContainerDelete:
		return "containerDelete"
	case MethodContainerEACL:
		return "containerEACL"
	case MethodContainerSetEACL:
		return "containerSetEACL"
	case MethodEndpointInfo:
		return "endpointInfo"
	case MethodNetworkInfo:
		return "networkInfo"
	case MethodObjectPut:
		return "objectPut"
	case MethodObjectDelete:
		return "objectDelete"
	case MethodObjectGet:
		return "objectGet"
	case MethodObjectHead:
		return "objectHead"
	case MethodObjectRange:
		return "objectRange"
	case MethodSessionCreate:
		return "sessionCreate"
	case MethodNetMapSnapshot:
		return "netMapSnapshot"
	case MethodObjectHash:
		return "objectHash"
	case MethodObjectSearch:
		return "objectSearch"
	case MethodContainerAnnounceUsedSpace:
		return "containerAnnounceUsedSpace"
	case MethodAnnounceIntermediateTrust:
		return "announceIntermediateTrust"
	case MethodAnnounceLocalTrust:
		return "announceLocalTrust"
	case MethodObjectGetStream:
		return "objectGetStream"
	case MethodObjectRangeStream:
		return "objectRangeStream"
	case MethodObjectSearchStream:
		return "objectSearchStream"
	case MethodObjectPutStream:
		return "objectPutStream"
	case MethodLast:
		return "it's a system name rather than a method"
	default:
		return "unknown"
	}
}

type (
	// OperationCallback describes common interface to external statistic collection.
	//
	// Passing zero duration means only error counting.
	OperationCallback = func(nodeKey []byte, endpoint string, method Method, duration time.Duration, err error)
)
