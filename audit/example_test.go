package audit_test

import (
	"github.com/nspcc-dev/neofs-sdk-go/audit"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

func ExampleResult() {
	var res audit.Result
	var cnr cid.ID

	res.ForEpoch(32)
	res.ForContainer(cnr)
	// ...
	res.SetCompleted(true)

	// Result instances can be stored in a binary format on client side.
	data := res.Marshal()

	// ...

	// Result instances can be restored from a binary format on server side.
	var auditResult audit.Result
	_ = auditResult.Unmarshal(data)
}
