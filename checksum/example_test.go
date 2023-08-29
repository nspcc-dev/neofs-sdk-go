package checksum

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

func ExampleCalculate() {
	payload := []byte{0, 1, 2, 3, 4, 5, 6}
	var checksum Checksum

	// checksum contains SHA256 hash of the payload
	Calculate(&checksum, SHA256, payload)

	// checksum contains TZ hash of the payload
	Calculate(&checksum, TZ, payload)
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleChecksum_marshalling() {
	var (
		csRaw [sha256.Size]byte
		csV2  refs.Checksum
		cs    Checksum
	)

	rand.Read(csRaw[:])
	cs.SetSHA256(csRaw)

	// On the client side.

	cs.WriteToV2(&csV2)

	fmt.Println(bytes.Equal(cs.Value(), csV2.GetSum()))
	// Example output: true

	// *send message*

	// On the server side.

	_ = cs.ReadFromV2(csV2)
}
