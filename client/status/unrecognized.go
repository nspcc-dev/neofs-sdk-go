package apistatus

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
)

// unrecognizedStatus is used for unknown statuses which cannot be interpreted
// by this package.
type unrecognizedStatusV2 struct {
	code    uint32
	msg     string
	details []*status.Status_Detail
}

func (x unrecognizedStatusV2) Error() string {
	const desc = "unknown"
	if x.msg != "" {
		return fmt.Sprintf(errFmt, x.code, fmt.Sprintf("%s, details count = %d", desc, len(x.details)), x.msg)
	}
	return fmt.Sprintf(errFmtNoMessage, x.code, fmt.Sprintf("%s, details count = %d", desc, len(x.details)))
}

func (x unrecognizedStatusV2) ErrorToV2() *status.Status {
	return &status.Status{
		Code:    x.code,
		Message: x.msg,
		Details: x.details,
	}
}

// Is checks whether target is of type unrecognizedStatusV2,
// *unrecognizedStatusV2 or [Error]. Is implements interface consumed by
// [errors.Is].
func (x unrecognizedStatusV2) Is(target error) bool { return errorIs(x, target) }
