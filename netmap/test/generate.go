package netmaptest

import "github.com/nspcc-dev/neofs-sdk-go/netmap"

func filter(withInner bool) *netmap.Filter {
	x := netmap.NewFilter()

	x.SetName("name")
	x.SetKey("key")
	x.SetValue("value")
	x.SetOperation(netmap.OpAND)

	if withInner {
		x.SetInnerFilters(filter(false), filter(false))
	}

	return x
}

// Filter returns random netmap.Filter.
func Filter() *netmap.Filter {
	return filter(true)
}

// Replica returns random netmap.Replica.
func Replica() *netmap.Replica {
	x := netmap.NewReplica()

	x.SetCount(666)
	x.SetSelector("selector")

	return x
}

// Selector returns random netmap.Selector.
func Selector() *netmap.Selector {
	x := netmap.NewSelector()

	x.SetCount(11)
	x.SetName("name")
	x.SetFilter("filter")
	x.SetAttribute("attribute")
	x.SetClause(netmap.ClauseDistinct)

	return x
}

// PlacementPolicy returns random netmap.PlacementPolicy.
func PlacementPolicy() *netmap.PlacementPolicy {
	x := netmap.NewPlacementPolicy()

	x.SetContainerBackupFactor(9)
	x.SetFilters(Filter(), Filter())
	x.SetReplicas(Replica(), Replica())
	x.SetSelectors(Selector(), Selector())

	return x
}

// NetworkParameter returns random netmap.NetworkParameter.
func NetworkParameter() *netmap.NetworkParameter {
	x := netmap.NewNetworkParameter()

	x.SetKey([]byte("key"))
	x.SetValue([]byte("value"))

	return x
}

// NetworkConfig returns random netmap.NetworkConfig.
func NetworkConfig() *netmap.NetworkConfig {
	x := netmap.NewNetworkConfig()

	x.SetParameters(
		NetworkParameter(),
		NetworkParameter(),
	)

	return x
}

// NetworkInfo returns random netmap.NetworkInfo.
func NetworkInfo() *netmap.NetworkInfo {
	x := netmap.NewNetworkInfo()

	x.SetCurrentEpoch(21)
	x.SetMagicNumber(32)
	x.SetMsPerBlock(43)
	x.SetNetworkConfig(NetworkConfig())

	return x
}
