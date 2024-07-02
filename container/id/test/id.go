package cidtest

import (
	"crypto/sha256"
	"math/rand"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// ID returns random cid.ID.
func ID() cid.ID {
	var res cid.ID
	//nolint:staticcheck
	rand.Read(res[:])
	return res
}

// IDs returns n random cid.ID instances.
func IDs(n int) []cid.ID {
	res := make([]cid.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}

// OtherID returns random cid.ID other than any given one.
func OtherID(vs ...cid.ID) cid.ID {
loop:
	for {
		v := ID()
		for i := range vs {
			if v == vs[i] {
				continue loop
			}
		}
		return v
	}
}

// IDWithChecksum returns cid.ID initialized
// with specified checksum.
// Deprecated: use [ID], [OtherID] or manual creation instead.
func IDWithChecksum(cs [sha256.Size]byte) cid.ID {
	var id cid.ID
	id.SetSHA256(cs)

	return id
}
