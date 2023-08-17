package netmap_test

import (
	apiGoNetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.netmap package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleNodeInfo_WriteToV2() {
	// import apiGoNetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"

	var info netmap.NodeInfo
	var msg apiGoNetmap.NodeInfo

	info.WriteToV2(&msg)

	// send msg
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.netmap package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleNodeInfo_ReadFromV2() {
	// import apiGoNetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"

	// recv msg

	var info netmap.NodeInfo
	var msg apiGoNetmap.NodeInfo

	_ = info.ReadFromV2(msg)

	// process info
}
