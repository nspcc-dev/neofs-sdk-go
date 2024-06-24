package checksumtest

import (
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
)

type fixedHash []byte

func (x fixedHash) Sum([]byte) []byte {
	return x
}

func (x fixedHash) Write([]byte) (n int, err error) { panic("unexpected call") }
func (x fixedHash) Reset()                          { panic("unexpected call") }
func (x fixedHash) Size() int                       { panic("unexpected call") }
func (x fixedHash) BlockSize() int                  { panic("unexpected call") }

// Checksum returns random checksum.Checksum.
func Checksum() checksum.Checksum {
	return checksum.NewFromHash(checksum.Type(rand.Uint32()%256), fixedHash("Hello, world!"))
}
