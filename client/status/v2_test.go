package apistatus_test

import (
	"errors"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestToStatusV2(t *testing.T) {
	type statusConstructor func() apistatus.Status

	for _, testItem := range [...]struct {
		status interface{} // Status or statusConstructor
		codeV2 uint64
	}{
		{
			status: errors.New("some error"),
			codeV2: 1024,
		},
		{
			status: 1,
			codeV2: 0,
		},
		{
			status: "text",
			codeV2: 0,
		},
		{
			status: true,
			codeV2: 0,
		},
		{
			status: true,
			codeV2: 0,
		},
		{
			status: nil,
			codeV2: 0,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				var st apistatus.ServerInternal

				st.SetMessage("internal error message")

				return st
			}),
			codeV2: 1024,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				var st apistatus.WrongMagicNumber

				st.WriteCorrectMagic(322)

				return st
			}),
			codeV2: 1025,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				return new(apistatus.ObjectLocked)
			}),
			codeV2: 2050,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				return new(apistatus.LockNonRegularObject)
			}),
			codeV2: 2051,
		},
	} {
		var st apistatus.Status

		if cons, ok := testItem.status.(statusConstructor); ok {
			st = cons()
		} else {
			st = testItem.status
		}

		stv2 := apistatus.ToStatusV2(st)

		// must generate the same status.Status message
		require.EqualValues(t, testItem.codeV2, stv2.Code())

		_, ok := st.(apistatus.StatusV2)
		if ok {
			// restore and convert again
			restored := apistatus.FromStatusV2(stv2)

			res := apistatus.ToStatusV2(restored)

			// must generate the same status.Status message
			require.Equal(t, stv2, res)
		}
	}
}

func TestFromStatusV2(t *testing.T) {
	type statusConstructor func() apistatus.Status

	for _, testItem := range [...]struct {
		status interface{} // Status or statusConstructor
		codeV2 uint64
	}{
		{
			status: errors.New("some error"),
			codeV2: 1024,
		},
		{
			status: 1,
			codeV2: 0,
		},
		{
			status: "text",
			codeV2: 0,
		},
		{
			status: true,
			codeV2: 0,
		},
		{
			status: true,
			codeV2: 0,
		},
		{
			status: nil,
			codeV2: 0,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				var st apistatus.ServerInternal

				st.SetMessage("internal error message")

				return st
			}),
			codeV2: 1024,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				var st apistatus.WrongMagicNumber

				st.WriteCorrectMagic(322)

				return st
			}),
			codeV2: 1025,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				return new(apistatus.ObjectLocked)
			}),
			codeV2: 2050,
		},
		{
			status: (statusConstructor)(func() apistatus.Status {
				return new(apistatus.LockNonRegularObject)
			}),
			codeV2: 2051,
		},
	} {
		var st apistatus.Status

		if cons, ok := testItem.status.(statusConstructor); ok {
			st = cons()
		} else {
			st = testItem.status
		}

		stv2 := apistatus.ToStatusV2(st)

		// must generate the same status.Status message
		require.EqualValues(t, testItem.codeV2, stv2.Code())

		_, ok := st.(apistatus.StatusV2)
		if ok {
			// restore and convert again
			restored := apistatus.FromStatusV2(stv2)

			res := apistatus.ToStatusV2(restored)

			// must generate the same status.Status message
			require.Equal(t, stv2, res)
		}
	}
}
