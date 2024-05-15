package audittest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/audit"
	audittest "github.com/nspcc-dev/neofs-sdk-go/audit/test"
	"github.com/stretchr/testify/require"
)

func TestDecimal(t *testing.T) {
	r := audittest.Result()
	require.NotEqual(t, r, audittest.Result())

	var r2 audit.Result
	require.NoError(t, r2.Unmarshal(r.Marshal()))
	require.EqualValues(t, r, r2)
}
