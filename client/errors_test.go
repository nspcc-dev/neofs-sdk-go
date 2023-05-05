package client_test

import (
	"fmt"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		errs        []error
		errVariable error
	}{
		{
			errs: []error{
				apistatus.ContainerNotFound{},
				new(apistatus.ContainerNotFound),
			},
			errVariable: apistatus.ErrContainerNotFound,
		},
		{
			errs: []error{
				apistatus.EACLNotFound{},
				new(apistatus.EACLNotFound),
			},
			errVariable: apistatus.ErrEACLNotFound,
		},
		{
			errs: []error{
				apistatus.ObjectNotFound{},
				new(apistatus.ObjectNotFound),
			},
			errVariable: apistatus.ErrObjectNotFound,
		},
		{
			errs: []error{
				apistatus.ObjectAlreadyRemoved{},
				new(apistatus.ObjectAlreadyRemoved),
			},
			errVariable: apistatus.ErrObjectAlreadyRemoved,
		},
		{
			errs: []error{
				apistatus.SessionTokenExpired{},
				new(apistatus.SessionTokenExpired),
			},
			errVariable: apistatus.ErrSessionTokenExpired,
		}, {
			errs: []error{
				apistatus.SessionTokenNotFound{},
				new(apistatus.SessionTokenNotFound),
			},
			errVariable: apistatus.ErrSessionTokenNotFound,
		},
	} {
		require.NotEmpty(t, tc.errs)
		require.NotNil(t, tc.errVariable)

		for i := range tc.errs {
			require.ErrorIs(t, tc.errs[i], tc.errVariable)

			wrapped := fmt.Errorf("some message %w", tc.errs[i])
			require.ErrorIs(t, wrapped, tc.errVariable)

			wrappedTwice := fmt.Errorf("another message %w", wrapped)
			require.ErrorIs(t, wrappedTwice, tc.errVariable)
		}
	}
}
