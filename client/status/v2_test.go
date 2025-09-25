package apistatus_test

import (
	"errors"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestToError(t *testing.T) {
	type statusConstructor func() error

	for _, testItem := range [...]struct {
		new            statusConstructor
		code           uint64
		message        string
		compatibleErrs []error
		checkAsErr     func(error) bool
	}{
		{
			new: func() error {
				return errors.New("some error")
			},
			code:    1024,
			message: "some error",
		},
		{
			new: func() error {
				return nil
			},
			code: 0,
		},
		{
			new: func() error {
				st := new(apistatus.Incomplete)
				st.SetMessage("2 out of 3 replicas made")

				return st
			},
			code:           1,
			compatibleErrs: []error{apistatus.ErrIncomplete, apistatus.Incomplete{}, &apistatus.Incomplete{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.Incomplete
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				st := new(apistatus.ServerInternal)
				st.SetMessage("internal error message")

				return st
			},
			code:           1024,
			compatibleErrs: []error{apistatus.ErrServerInternal, apistatus.ServerInternal{}, &apistatus.ServerInternal{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ServerInternal
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				st := new(apistatus.WrongMagicNumber)
				st.WriteCorrectMagic(322)

				return st
			},
			code:           1025,
			compatibleErrs: []error{apistatus.ErrWrongMagicNumber, apistatus.WrongMagicNumber{}, &apistatus.WrongMagicNumber{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.WrongMagicNumber
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.ObjectLocked)
			},
			code:           2050,
			compatibleErrs: []error{apistatus.ErrObjectLocked, apistatus.ObjectLocked{}, &apistatus.ObjectLocked{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectLocked
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.LockNonRegularObject)
			},
			code:           2051,
			compatibleErrs: []error{apistatus.ErrLockNonRegularObject, apistatus.LockNonRegularObject{}, &apistatus.LockNonRegularObject{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.LockNonRegularObject
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				st := new(apistatus.ObjectAccessDenied)
				st.WriteReason("any reason")

				return st
			},
			code:           2048,
			compatibleErrs: []error{apistatus.ErrObjectAccessDenied, apistatus.ObjectAccessDenied{}, &apistatus.ObjectAccessDenied{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectAccessDenied
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.ObjectNotFound)
			},
			code:           2049,
			compatibleErrs: []error{apistatus.ErrObjectNotFound, apistatus.ObjectNotFound{}, &apistatus.ObjectNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectNotFound
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.ObjectAlreadyRemoved)
			},
			code:           2052,
			compatibleErrs: []error{apistatus.ErrObjectAlreadyRemoved, apistatus.ObjectAlreadyRemoved{}, &apistatus.ObjectAlreadyRemoved{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectAlreadyRemoved
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.ObjectOutOfRange)
			},
			code:           2053,
			compatibleErrs: []error{apistatus.ErrObjectOutOfRange, apistatus.ObjectOutOfRange{}, &apistatus.ObjectOutOfRange{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectOutOfRange
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.QuotaExceeded)
			},
			code:           2054,
			compatibleErrs: []error{apistatus.ErrQuotaExceeded, apistatus.QuotaExceeded{}, &apistatus.QuotaExceeded{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.QuotaExceeded
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.ContainerNotFound)
			},
			code:           3072,
			compatibleErrs: []error{apistatus.ErrContainerNotFound, apistatus.ContainerNotFound{}, &apistatus.ContainerNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ContainerNotFound
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.EACLNotFound)
			},
			code:           3073,
			compatibleErrs: []error{apistatus.ErrEACLNotFound, apistatus.EACLNotFound{}, &apistatus.EACLNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.EACLNotFound
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.SessionTokenNotFound)
			},
			code:           4096,
			compatibleErrs: []error{apistatus.ErrSessionTokenNotFound, apistatus.SessionTokenNotFound{}, &apistatus.SessionTokenNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.SessionTokenNotFound
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.SessionTokenExpired)
			},
			code:           4097,
			compatibleErrs: []error{apistatus.ErrSessionTokenExpired, apistatus.SessionTokenExpired{}, &apistatus.SessionTokenExpired{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.SessionTokenExpired
				return errors.As(err, &target)
			},
		},
		{
			new: func() error {
				return new(apistatus.NodeUnderMaintenance)
			},
			code:           1027,
			compatibleErrs: []error{apistatus.ErrNodeUnderMaintenance, apistatus.NodeUnderMaintenance{}, &apistatus.NodeUnderMaintenance{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.NodeUnderMaintenance
				return errors.As(err, &target)
			},
		},
	} {
		st := testItem.new()

		m := apistatus.FromError(st)

		// must generate the same status.Status message
		require.EqualValues(t, testItem.code, m.GetCode())
		if len(testItem.message) > 0 {
			require.Equal(t, testItem.message, m.Message)
		}

		// restore and convert again
		restored := apistatus.ToError(m)

		res := apistatus.FromError(restored)

		// must generate the same status.Status message
		require.Equal(t, m, res)

		randomError := errors.New("garbage")
		for _, err := range testItem.compatibleErrs {
			require.ErrorIs(t, st, err)
			require.NotErrorIs(t, randomError, err)
		}

		if testItem.checkAsErr != nil {
			require.True(t, testItem.checkAsErr(st))
		}
	}
}
