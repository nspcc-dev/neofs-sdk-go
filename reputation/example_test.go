package reputation_test

import (
	apiGoReputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.reputation package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleGlobalTrust_WriteToV2() {
	// import apiGoReputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"

	var trust reputation.GlobalTrust
	var msg apiGoReputation.GlobalTrust
	trust.WriteToV2(&msg)

	// send trust
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.reputation package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleGlobalTrust_ReadFromV2() {
	// import apiGoReputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"

	// recv msg

	var trust reputation.GlobalTrust
	var msg apiGoReputation.GlobalTrust

	_ = trust.ReadFromV2(msg)

	// process trust
}
