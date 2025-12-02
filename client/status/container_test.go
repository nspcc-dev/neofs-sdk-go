package apistatus_test

import (
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestNewContainerLocked(t *testing.T) {
	var e apistatus.ContainerLocked
	require.EqualError(t, e, "status: code = 3074 message = container is locked")

	e = apistatus.NewContainerLocked("some message")
	require.EqualError(t, e, "status: code = 3074 message = some message")
}
