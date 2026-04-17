package checksumtest

import (
	"crypto/sha256"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
)

// Checksum returns random checksum.Checksum.
func Checksum() checksum.Checksum {
	data := testutil.RandByteSlice(sha256.Size)
	res, err := checksum.NewFromData(checksum.SHA256, data)
	if err != nil {
		panic(fmt.Errorf("unexpected NewFromData error: %w", err))
	}
	return res
}
