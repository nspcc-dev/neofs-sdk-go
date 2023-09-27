package netmap_test

import (
	"fmt"

	apiGoNetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleNodeInfo_marshalling() {
	// import apiGoNetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"

	// On the client side.

	var info netmap.NodeInfo
	var msg apiGoNetmap.NodeInfo
	info.WriteToV2(&msg)
	// *send message*

	// On the server side.

	_ = info.ReadFromV2(msg)
}

// When forming information about storage node to be registered the NeoFS
// network, the node may be optionally associated with some private group of
// storage nodes in the network. The groups are managed by their owners in
// corresponding NeoFS NNS domains.
func ExampleNodeInfo_SetVerifiedNodesDomain() {
	var bNodePublicKey []byte

	var n netmap.NodeInfo
	n.SetPublicKey(bNodePublicKey)
	// other info
	n.SetVerifiedNodesDomain("nodes.some-org.neofs")
	// to be allowed into the network, set public key must be in the access list
	// managed in the specified domain

	// the specified domain is later processed by the system
	fmt.Printf("Verified nodes' domain: %s\n", n.VerifiedNodesDomain())
}
