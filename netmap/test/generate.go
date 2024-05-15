package netmaptest

import (
	"math/rand"
	"strconv"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

func filter(allowComplex bool) (x netmap.Filter) {
	x.SetName("filter_" + strconv.Itoa(rand.Int()))
	rnd := rand.Int() % 8
	switch rnd {
	case 0:
		x.Equal("key_"+strconv.Itoa(rand.Int()), "value_"+strconv.Itoa(rand.Int()))
	case 1:
		x.NotEqual("key_"+strconv.Itoa(rand.Int()), "value_"+strconv.Itoa(rand.Int()))
	case 2:
		x.NumericGT("key_"+strconv.Itoa(rand.Int()), rand.Int63())
	case 3:
		x.NumericGE("key_"+strconv.Itoa(rand.Int()), rand.Int63())
	case 4:
		x.NumericLT("key_"+strconv.Itoa(rand.Int()), rand.Int63())
	case 5:
		x.NumericLE("key_"+strconv.Itoa(rand.Int()), rand.Int63())
	}
	if allowComplex {
		fs := make([]netmap.Filter, 1+rand.Int()%3)
		for i := range fs {
			fs[i] = filter(false)
		}
		if rnd == 6 {
			x.LogicalAND(fs...)
		} else {
			x.LogicalOR(fs...)
		}
	}

	return x
}

// Filter returns random netmap.Filter.
func Filter() netmap.Filter {
	return filter(true)
}

// NFilters returns n random netmap.Filter instances.
func NFilters(n int) []netmap.Filter {
	res := make([]netmap.Filter, n)
	for i := range res {
		res[i] = Filter()
	}
	return res
}

// Replica returns random netmap.ReplicaDescriptor.
func Replica() netmap.ReplicaDescriptor {
	var x netmap.ReplicaDescriptor
	x.SetNumberOfObjects(rand.Uint32())
	x.SetSelectorName("selector_" + strconv.Itoa(rand.Int()))

	return x
}

// NReplicas returns n random netmap.ReplicaDescriptor instances.
func NReplicas(n int) []netmap.ReplicaDescriptor {
	res := make([]netmap.ReplicaDescriptor, n)
	for i := range res {
		res[i] = Replica()
	}
	return res
}

// Selector returns random netmap.Selector.
func Selector() netmap.Selector {
	var x netmap.Selector
	x.SetNumberOfNodes(uint32(rand.Int()))
	x.SetName("selector" + strconv.Itoa(rand.Int()))
	x.SetFilterName("filter_" + strconv.Itoa(rand.Int()))
	x.SelectByBucketAttribute("attribute_" + strconv.Itoa(rand.Int()))
	switch rand.Int() % 3 {
	case 1:
		x.SelectSame()
	case 2:
		x.SelectDistinct()
	}

	return x
}

// NSelectors returns n random netmap.Selector instances.
func NSelectors(n int) []netmap.Selector {
	res := make([]netmap.Selector, n)
	for i := range res {
		res[i] = Selector()
	}
	return res
}

// PlacementPolicy returns random netmap.PlacementPolicy.
func PlacementPolicy() netmap.PlacementPolicy {
	var p netmap.PlacementPolicy
	p.SetContainerBackupFactor(uint32(rand.Int()))
	p.SetReplicas(NReplicas(1 + rand.Int()%3))
	if n := rand.Int() % 4; n > 0 {
		p.SetFilters(NFilters(n))
	}
	if n := rand.Int() % 4; n > 0 {
		p.SetSelectors(NSelectors(n))
	}

	return p
}

// NetworkInfo returns random netmap.NetworkInfo.
func NetworkInfo() netmap.NetworkInfo {
	var x netmap.NetworkInfo
	x.SetCurrentEpoch(rand.Uint64())
	x.SetMagicNumber(rand.Uint64())
	x.SetMsPerBlock(rand.Int63())
	x.SetAuditFee(rand.Uint64())
	x.SetStoragePrice(rand.Uint64())
	x.SetContainerFee(rand.Uint64())
	x.SetNamedContainerFee(rand.Uint64())
	x.SetEigenTrustAlpha(rand.Float64())
	x.SetNumberOfEigenTrustIterations(rand.Uint64())
	x.SetEpochDuration(rand.Uint64())
	x.SetIRCandidateFee(rand.Uint64())
	x.SetMaxObjectSize(rand.Uint64())
	x.SetWithdrawalFee(rand.Uint64())
	x.SetHomomorphicHashingDisabled(rand.Int()%2 == 0)
	x.SetMaintenanceModeAllowed(rand.Int()%2 == 0)
	for i := 0; i < rand.Int()%4; i++ {
		val := make([]byte, rand.Int()%64+1)
		rand.Read(val)
		x.SetRawNetworkParameter("prm_"+strconv.Itoa(rand.Int()), val)
	}

	return x
}

// NodeInfo returns random netmap.NodeInfo.
func NodeInfo() netmap.NodeInfo {
	var x netmap.NodeInfo
	key := make([]byte, 33)
	//nolint:staticcheck
	rand.Read(key)

	x.SetPublicKey(key)

	endpoints := make([]string, 1+rand.Int()%3)
	for i := range endpoints {
		endpoints[i] = "endpoint_" + strconv.Itoa(rand.Int())
	}
	x.SetNetworkEndpoints(endpoints)

	return x
}

// NNodes returns n random netmap.NodeInfo instances.
func NNodes(n int) []netmap.NodeInfo {
	res := make([]netmap.NodeInfo, n)
	for i := range res {
		res[i] = NodeInfo()
	}
	return res
}

// Netmap returns random netmap.NetMap.
func Netmap() netmap.NetMap {
	var x netmap.NetMap
	x.SetEpoch(rand.Uint64())
	if n := rand.Int() % 4; n > 0 {
		x.SetNodes(NNodes(n))
	}
	return x
}
