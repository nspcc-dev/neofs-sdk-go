package cidtest

import (
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

// NIDs returns n random cid.ID instances.
func NIDs(n int) []cid.ID {
	res := make([]cid.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}

// ChangeID returns container ID other than the given one.
func ChangeID(id cid.ID) cid.ID {
	id[0]++
	return id
}
