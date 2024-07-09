package checksumtest

import (
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
)

var allSupportedTypes = []checksum.Type{checksum.SHA256, checksum.TillichZemor}

// Checksum returns random checksum.Checksum.
func Checksum() checksum.Checksum {
	data := make([]byte, 32)
	//nolint:staticcheck
	rand.Read(data)
	i := rand.Int() % len(allSupportedTypes)
	res, err := checksum.NewFromData(allSupportedTypes[i], data)
	if err != nil {
		panic(fmt.Errorf("unexpected NewFromData error: %w", err))
	}
	return res
}
