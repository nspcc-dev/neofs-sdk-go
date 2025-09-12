package cidtest

import (
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
)

// ID returns random cid.ID.
func ID() cid.ID {
	return cid.ID(testutil.RandByteSlice(cid.Size))
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
