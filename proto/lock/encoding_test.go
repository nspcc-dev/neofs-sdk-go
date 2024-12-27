package lock_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/lock"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

func TestLock_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &lock.Lock{
			Members: []*refs.ObjectID{nil, {}},
		}

		var dst lock.Lock
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ms := dst.GetMembers()
		require.Len(t, ms, 2)
		require.Equal(t, ms[0], new(refs.ObjectID))
		require.Equal(t, ms[1], new(refs.ObjectID))
	})

	prototest.TestMarshalStable(t, []*lock.Lock{
		{Members: prototest.RandObjectIDs()},
	})
}
