package checksumtest

import (
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
)

var allSupportedTypes = []checksum.Type{checksum.SHA256, checksum.TillichZemor}

// Checksum returns random checksum.Checksum.
func Checksum() checksum.Checksum {
	data := testutil.RandByteSlice(32)
	i := rand.Int() % len(allSupportedTypes)
	res, err := checksum.NewFromData(allSupportedTypes[i], data)
	if err != nil {
		panic(fmt.Errorf("unexpected NewFromData error: %w", err))
	}
	return res
}
