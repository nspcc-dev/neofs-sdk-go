package client_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/stretchr/testify/require"
)

func Test_SignError(t *testing.T) {
	someErr := errors.New("some error")
	signErr := client.NewSignError(someErr)

	require.ErrorIs(t, signErr, someErr)
	require.ErrorIs(t, signErr, client.ErrSign)
}
