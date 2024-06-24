package checksum

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
)

func ExampleCalculate() {
	payload := []byte{0, 1, 2, 3, 4, 5, 6}

	// checksum contains SHA256 hash of the payload
	cs := Calculate(SHA256, payload)
	fmt.Println(cs.Type(), cs.Value())

	// checksum contains TZ hash of the payload
	cs = Calculate(TZ, payload)
	fmt.Println(cs.Type(), cs.Value())
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleChecksum_marshalling() {
	var h [sha256.Size]byte
	//nolint:staticcheck
	rand.Read(h[:])

	cs := NewSHA256(h)

	// On the client side.
	var msg refs.Checksum
	cs.WriteToV2(&msg)

	fmt.Println(bytes.Equal(cs.Value(), msg.GetSum()))
	// Example output: true

	// *send message*

	// On the server side.

	_ = cs.ReadFromV2(&msg)
}
