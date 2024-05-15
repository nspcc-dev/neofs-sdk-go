package apistatus_test

import (
	"fmt"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func assertErrorIs[T error, PTR interface {
	*T
	error
}](t testing.TB, constErr T) {
	var e T
	var pe PTR = &e
	require.ErrorIs(t, e, e)
	require.ErrorIs(t, e, pe)
	require.ErrorIs(t, pe, e)
	require.ErrorIs(t, pe, pe)
	require.ErrorIs(t, e, apistatus.Error)
	require.ErrorIs(t, e, constErr)
	we := fmt.Errorf("wrapped %w", e)
	require.ErrorIs(t, we, e)
	require.ErrorIs(t, we, pe)
	require.ErrorIs(t, we, apistatus.Error)
	require.ErrorIs(t, we, constErr)
	wwe := fmt.Errorf("again %w", e)
	require.ErrorIs(t, wwe, e)
	require.ErrorIs(t, wwe, pe)
	require.ErrorIs(t, wwe, apistatus.Error)
	require.ErrorIs(t, wwe, constErr)
}
