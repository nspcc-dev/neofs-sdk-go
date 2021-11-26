package netmap

import (
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// EnterSubnet writes to NodeInfo the intention to enter the subnet. Must not be called on nil.
// Zero NodeInfo belongs to zero subnet.
func (i *NodeInfo) EnterSubnet(id subnetid.ID) {
	i.changeSubnet(id, true)
}

// ExitSubnet writes to NodeInfo the intention to exit subnet. Must not be called on nil.
func (i *NodeInfo) ExitSubnet(id subnetid.ID) {
	i.changeSubnet(id, false)
}

func (i *NodeInfo) changeSubnet(id subnetid.ID, isMember bool) {
	var (
		idv2 refs.SubnetID
		info netmap.NodeSubnetInfo
	)

	id.WriteToV2(&idv2)

	info.SetID(&idv2)
	info.SetEntryFlag(isMember)

	netmap.WriteSubnetInfo((*netmap.NodeInfo)(i), info)
}

// ErrRemoveSubnet is returned when a node needs to leave the subnet.
var ErrRemoveSubnet = netmap.ErrRemoveSubnet

// IterateSubnets iterates over all subnets the node belongs to and passes the IDs to f.
// Must not be called on nil. Handler must not be nil.
//
// If f returns ErrRemoveSubnet, then removes subnet entry. Note that this leads to an instant mutation of NodeInfo.
// Breaks on any other non-nil error and returns it.
//
// Returns an error if subnet incorrectly enabled/disabled.
// Returns an error if the node is not included in any subnet by the end of the loop.
func (i *NodeInfo) IterateSubnets(f func(subnetid.ID) error) error {
	var id subnetid.ID

	return netmap.IterateSubnets((*netmap.NodeInfo)(i), func(idv2 refs.SubnetID) error {
		id.FromV2(idv2)

		err := f(id)
		if errors.Is(err, ErrRemoveSubnet) {
			return netmap.ErrRemoveSubnet
		}

		return err
	})
}

var errAbortSubnetIter = errors.New("abort subnet iterator")

// BelongsToSubnet checks if node belongs to subnet by ID.
//
// Function is NPE-safe: nil NodeInfo always belongs to zero subnet only.
func BelongsToSubnet(node *NodeInfo, id subnetid.ID) bool {
	err := node.IterateSubnets(func(id_ subnetid.ID) error {
		if id.Equals(&id_) {
			return errAbortSubnetIter
		}

		return nil
	})

	return errors.Is(err, errAbortSubnetIter)
}
