package testutil

import (
	"crypto/rand"

	"github.com/nspcc-dev/neo-go/pkg/util"
)

// Integer is just some int type to simplify [RandByteSlice] use.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// RandScriptHash returns random 20-byte array.
func RandScriptHash() util.Uint160 {
	return util.Uint160(RandByteSlice(util.Uint160Size))
}

// RandByteSlice returns randomized byte slice of specified length.
func RandByteSlice[I Integer](ln I) []byte {
	b := make([]byte, ln)
	_, _ = rand.Read(b)
	return b
}
