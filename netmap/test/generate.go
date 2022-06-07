package netmaptest

import (
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	subnetidtest "github.com/nspcc-dev/neofs-sdk-go/subnet/id/test"
)

func filter(withInner bool) (x netmap.Filter) {
	x.SetName("name")
	if withInner {
		x.LogicalOR(filter(false), filter(false))
	} else {
		x.NumericGE("epoch", 13)
	}

	return x
}

// Filter returns random netmap.Filter.
func Filter() netmap.Filter {
	return filter(true)
}

// Replica returns random netmap.ReplicaDescriptor.
func Replica() (x netmap.ReplicaDescriptor) {
	x.SetAmount(666)
	x.SetSelectorName("selector")

	return
}

// Selector returns random netmap.Selector.
func Selector() (x netmap.Selector) {
	x.SetNodeAmount(11)
	x.SetName("name")
	x.SetFilterName("filter")
	x.SelectByBucketAttribute("attribute")
	x.SelectDistinct()

	return
}

// PlacementPolicy returns random netmap.PlacementPolicy.
func PlacementPolicy() (p netmap.PlacementPolicy) {
	p.SetContainerBackupFactor(9)
	p.AddFilters(Filter(), Filter())
	p.AddReplicas(Replica(), Replica())
	p.AddSelectors(Selector(), Selector())
	p.RestrictSubnet(subnetidtest.ID())

	return
}

// NetworkInfo returns random netmap.NetworkInfo.
func NetworkInfo() (x netmap.NetworkInfo) {
	x.SetCurrentEpoch(21)
	x.SetMagicNumber(32)
	x.SetMsPerBlock(43)
	x.SetAuditFee(1)
	x.SetStoragePrice(2)
	x.SetContainerFee(3)
	x.SetEigenTrustAlpha(0.4)
	x.SetEigenTrustIterationAmount(5)
	x.SetEpochDuration(6)
	x.SetIRCandidateFee(7)
	x.SetMaxObjectSize(8)
	x.SetWithdrawalFee(9)

	return
}
