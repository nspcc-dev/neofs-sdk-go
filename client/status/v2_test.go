package apistatus_test

import (
	"errors"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestFromStatusV2(t *testing.T) {
	type statusConstructor func() error

	for _, testItem := range [...]struct {
		status         any // Status or statusConstructor
		codeV2         uint64
		messageV2      string
		compatibleErrs []error
		checkAsErr     func(error) bool
	}{
		{
			status: (statusConstructor)(func() error {
				return errors.New("some error")
			}),
			codeV2:    1024,
			messageV2: "some error",
		},
		{
			status: (statusConstructor)(func() error {
				return nil
			}),
			codeV2: 0,
		},
		{
			status: (statusConstructor)(func() error {
				st := new(apistatus.ServerInternal)
				st.SetMessage("internal error message")

				return st
			}),
			codeV2:         1024,
			compatibleErrs: []error{apistatus.ErrServerInternal, apistatus.ServerInternal{}, &apistatus.ServerInternal{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ServerInternal
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				st := new(apistatus.WrongMagicNumber)
				st.WriteCorrectMagic(322)

				return st
			}),
			codeV2:         1025,
			compatibleErrs: []error{apistatus.ErrWrongMagicNumber, apistatus.WrongMagicNumber{}, &apistatus.WrongMagicNumber{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.WrongMagicNumber
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.ObjectLocked)
			}),
			codeV2:         2050,
			compatibleErrs: []error{apistatus.ErrObjectLocked, apistatus.ObjectLocked{}, &apistatus.ObjectLocked{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectLocked
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.LockNonRegularObject)
			}),
			codeV2:         2051,
			compatibleErrs: []error{apistatus.ErrLockNonRegularObject, apistatus.LockNonRegularObject{}, &apistatus.LockNonRegularObject{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.LockNonRegularObject
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				st := new(apistatus.ObjectAccessDenied)
				st.WriteReason("any reason")

				return st
			}),
			codeV2:         2048,
			compatibleErrs: []error{apistatus.ErrObjectAccessDenied, apistatus.ObjectAccessDenied{}, &apistatus.ObjectAccessDenied{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectAccessDenied
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.ObjectNotFound)
			}),
			codeV2:         2049,
			compatibleErrs: []error{apistatus.ErrObjectNotFound, apistatus.ObjectNotFound{}, &apistatus.ObjectNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectNotFound
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.ObjectAlreadyRemoved)
			}),
			codeV2:         2052,
			compatibleErrs: []error{apistatus.ErrObjectAlreadyRemoved, apistatus.ObjectAlreadyRemoved{}, &apistatus.ObjectAlreadyRemoved{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectAlreadyRemoved
				return errors.As(err, &target)
			},
		},
		{
			status: statusConstructor(func() error {
				return new(apistatus.ObjectOutOfRange)
			}),
			codeV2:         2053,
			compatibleErrs: []error{apistatus.ErrObjectOutOfRange, apistatus.ObjectOutOfRange{}, &apistatus.ObjectOutOfRange{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ObjectOutOfRange
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.ContainerNotFound)
			}),
			codeV2:         3072,
			compatibleErrs: []error{apistatus.ErrContainerNotFound, apistatus.ContainerNotFound{}, &apistatus.ContainerNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.ContainerNotFound
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.EACLNotFound)
			}),
			codeV2:         3073,
			compatibleErrs: []error{apistatus.ErrEACLNotFound, apistatus.EACLNotFound{}, &apistatus.EACLNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.EACLNotFound
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.SessionTokenNotFound)
			}),
			codeV2:         4096,
			compatibleErrs: []error{apistatus.ErrSessionTokenNotFound, apistatus.SessionTokenNotFound{}, &apistatus.SessionTokenNotFound{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.SessionTokenNotFound
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.SessionTokenExpired)
			}),
			codeV2:         4097,
			compatibleErrs: []error{apistatus.ErrSessionTokenExpired, apistatus.SessionTokenExpired{}, &apistatus.SessionTokenExpired{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.SessionTokenExpired
				return errors.As(err, &target)
			},
		},
		{
			status: (statusConstructor)(func() error {
				return new(apistatus.NodeUnderMaintenance)
			}),
			codeV2:         1027,
			compatibleErrs: []error{apistatus.ErrNodeUnderMaintenance, apistatus.NodeUnderMaintenance{}, &apistatus.NodeUnderMaintenance{}, apistatus.Error},
			checkAsErr: func(err error) bool {
				var target *apistatus.NodeUnderMaintenance
				return errors.As(err, &target)
			},
		},
	} {
		var st error
		cons, ok := testItem.status.(statusConstructor)
		require.True(t, ok)

		st = cons()

		stv2 := apistatus.ToStatusV2(st)

		// must generate the same status.Status message
		require.EqualValues(t, testItem.codeV2, stv2.Code())
		if len(testItem.messageV2) > 0 {
			require.Equal(t, testItem.messageV2, stv2.Message())
		}

		_, ok = st.(apistatus.StatusV2)
		if ok {
			// restore and convert again
			restored := apistatus.FromStatusV2(stv2)

			res := apistatus.ToStatusV2(restored)

			// must generate the same status.Status message
			require.Equal(t, stv2, res)
		}

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
