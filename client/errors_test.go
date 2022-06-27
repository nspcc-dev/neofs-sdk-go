package client_test

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		check func(error) bool
		errs  []error
	}{
		{
			check: client.IsErrContainerNotFound,
			errs: []error{
				apistatus.ContainerNotFound{},
				new(apistatus.ContainerNotFound),
			},
		},
		{
			check: client.IsErrObjectNotFound,
			errs: []error{
				apistatus.ObjectNotFound{},
				new(apistatus.ObjectNotFound),
			},
		},
		{
			check: client.IsErrObjectAlreadyRemoved,
			errs: []error{
				apistatus.ObjectAlreadyRemoved{},
				new(apistatus.ObjectAlreadyRemoved),
			},
		},
		{
			check: client.IsErrSessionExpired,
			errs: []error{
				apistatus.SessionTokenExpired{},
				new(apistatus.SessionTokenExpired),
			},
		}, {
			check: client.IsErrSessionNotFound,
			errs: []error{
				apistatus.SessionTokenNotFound{},
				new(apistatus.SessionTokenNotFound),
			},
		},
	} {
		require.NotEmpty(t, tc.errs)

		for i := range tc.errs {
			require.True(t, tc.check(tc.errs[i]), tc.errs[i])
			require.True(t, tc.check(fmt.Errorf("some context: %w", tc.errs[i])), tc.errs[i])
		}
	}
}
