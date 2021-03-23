package object_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

func TestNewSplitInfoError(t *testing.T) {
	var (
		si = generateSplitInfo()

		err         = object.NewSplitInfoError(si)
		expectedErr *object.SplitInfoError
	)

	require.True(t, errors.As(err, &expectedErr))

	siErr, ok := err.(*object.SplitInfoError)
	require.True(t, ok)
	require.Equal(t, si, siErr.SplitInfo())
}

func generateSplitInfo() *object.SplitInfo {
	var si object.SplitInfo
	si.SetSplitID(object.NewSplitID())
	si.SetLastPart(generateID())
	si.SetLink(generateID())

	return &si
}
