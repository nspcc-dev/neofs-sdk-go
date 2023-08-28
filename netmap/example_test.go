package netmap_test

import (
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
