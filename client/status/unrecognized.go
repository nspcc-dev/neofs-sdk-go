package apistatus

import (
	"github.com/nspcc-dev/neofs-api-go/v2/status"
)

type unrecognizedStatusV2 struct {
	v2 status.Status
}

func (x unrecognizedStatusV2) Error() string {
	return errMessageStatusV2("unrecognized", x.v2.Message())
}

// implements method of the FromStatusV2 local interface.
func (x *unrecognizedStatusV2) fromStatusV2(st *status.Status) {
	x.v2 = *st
}
