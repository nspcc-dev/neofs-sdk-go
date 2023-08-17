package reputation_test

import (
	apiGoReputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleGlobalTrust_marshalling() {
	// import apiGoReputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"

	// On the client side.
	var trust reputation.GlobalTrust
	var msg apiGoReputation.GlobalTrust
	trust.WriteToV2(&msg)
	// *send message*

	// On the server side.
	_ = trust.ReadFromV2(msg)
}
