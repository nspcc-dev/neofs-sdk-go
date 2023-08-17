package audit_test

import (
	"github.com/nspcc-dev/neofs-sdk-go/audit"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// Result type groups values which can be gathered during data audit process.
func Example() {
	var res audit.Result
	var cnr cid.ID

	res.ForEpoch(32)
	res.ForContainer(cnr)
	// ...
	res.Complete()
}

// Result instances can be stored in a binary format on reporter side.
func ExampleResult_Marshal() {
	var res audit.Result
	data := res.Marshal()
	_ = data

	// send data
}

// Result instances can be restored from a binary format on receiver side.
func ExampleResult_Unmarshal() {
	var data []byte

	var res audit.Result
	_ = res.Unmarshal(data)

	// process result
}
