package ec_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object/ec"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestErrParts(t *testing.T) {
	ids := oidtest.IDs(10)
	err := ec.ErrParts(ids)

	t.Run("errors.As", func(t *testing.T) {
		var target ec.ErrParts
		require.ErrorAs(t, err, &target)
		require.EqualValues(t, ids, target)
	})

	require.Implements(t, new(error), err)
	require.EqualError(t, err, "10 EC parts")
}
