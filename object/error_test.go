package object_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestNewSplitInfoError(t *testing.T) {
	var (
		si = generateSplitInfo()

		err         error = object.NewSplitInfoError(si)
		expectedErr *object.SplitInfoError
	)

	require.True(t, errors.As(err, &expectedErr))
	require.Equal(t, si, expectedErr.SplitInfo())
}

func generateSplitInfo() *object.SplitInfo {
	si := object.NewSplitInfo()
	si.SetSplitID(object.NewSplitID())
	si.SetLastPart(oidtest.ID())
	si.SetLink(oidtest.ID())

	return si
}
