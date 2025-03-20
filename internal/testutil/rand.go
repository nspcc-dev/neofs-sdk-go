package testutil

import (
	"encoding/binary"
	"math/rand/v2"
	"time"

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
	RandRead(b)
	return b
}

// RandRead randomizes b bytes.
func RandRead(b []byte) {
	var seed [32]byte
	binary.LittleEndian.PutUint64(seed[:], uint64(time.Now().UnixNano()))
	_, _ = rand.NewChaCha8(seed).Read(b) // docs say never returns an error
}
