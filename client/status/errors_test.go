package apistatus

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		errs        []error
		errVariable error
	}{
		{
			errs:        []error{Incomplete{}, new(Incomplete)},
			errVariable: ErrIncomplete,
		},
		{
			errs:        []error{ServerInternal{}, new(ServerInternal)},
			errVariable: ErrServerInternal,
		},
		{
			errs:        []error{WrongMagicNumber{}, new(WrongMagicNumber)},
			errVariable: ErrWrongMagicNumber,
		},
		{
			errs:        []error{SignatureVerification{}, new(SignatureVerification)},
			errVariable: ErrSignatureVerification,
		},
		{
			errs:        []error{NodeUnderMaintenance{}, new(NodeUnderMaintenance)},
			errVariable: ErrNodeUnderMaintenance,
		},
		{
			errs:        []error{BadRequest{}, new(BadRequest)},
			errVariable: ErrBadRequest,
		},
		{
			errs:        []error{Busy{}, new(Busy)},
			errVariable: ErrBusy,
		},
		{
			errs:        []error{ObjectLocked{}, new(ObjectLocked)},
			errVariable: ErrObjectLocked,
		},
		{
			errs:        []error{LockNonRegularObject{}, new(LockNonRegularObject)},
			errVariable: ErrLockNonRegularObject,
		},
		{
			errs:        []error{ObjectAccessDenied{}, new(ObjectAccessDenied)},
			errVariable: ErrObjectAccessDenied,
		},
		{
			errs:        []error{ObjectNotFound{}, new(ObjectNotFound)},
			errVariable: ErrObjectNotFound,
		},
		{
			errs:        []error{ObjectAlreadyRemoved{}, new(ObjectAlreadyRemoved)},
			errVariable: ErrObjectAlreadyRemoved,
		},
		{
			errs:        []error{ObjectOutOfRange{}, new(ObjectOutOfRange)},
			errVariable: ErrObjectOutOfRange,
		},
		{
			errs:        []error{QuotaExceeded{}, new(QuotaExceeded)},
			errVariable: ErrQuotaExceeded,
		},
		{
			errs:        []error{ContainerNotFound{}, new(ContainerNotFound)},
			errVariable: ErrContainerNotFound,
		},
		{
			errs:        []error{EACLNotFound{}, new(EACLNotFound)},
			errVariable: ErrEACLNotFound,
		},
		{
			errs:        []error{ContainerLocked{}, new(ContainerLocked)},
			errVariable: ErrContainerLocked,
		},
		{
			errs:        []error{ContainerAwaitTimeout{}, new(ContainerAwaitTimeout)},
			errVariable: ErrContainerAwaitTimeout,
		},
		{
			errs:        []error{SessionTokenExpired{}, new(SessionTokenExpired)},
			errVariable: ErrSessionTokenExpired,
		},
		{
			errs:        []error{SessionTokenNotFound{}, new(SessionTokenNotFound)},
			errVariable: ErrSessionTokenNotFound,
		},

		{
			errs:        []error{UnrecognizedStatus{}, new(UnrecognizedStatus)},
			errVariable: ErrUnrecognizedStatus,
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
