package apistatus_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestObjectLocked_Error(t *testing.T) {
	var e apistatus.ObjectLocked
	require.EqualError(t, e, "status: code = 2050 (object is locked)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2050, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2050 (object is locked) message = any message")
}

func TestObjectLocked_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrObjectLocked)
}

func TestObjectLocked_As(t *testing.T) {
	var src, dst apistatus.ObjectLocked
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 2050, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestObjectLocked_ErrorToV2(t *testing.T) {
	var e apistatus.ObjectLocked
	st := e.ErrorToV2()
	require.EqualValues(t, 2050, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2050, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 2050, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}

func TestLockIrregularObject_Error(t *testing.T) {
	var e apistatus.LockIrregularObject
	require.EqualError(t, e, "status: code = 2051 (locking irregular object is forbidden)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2051, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2051 (locking irregular object is forbidden) message = any message")
}

func TestLockIrregularObject_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrLockIrregularObject)
}

func TestLockIrregularObject_As(t *testing.T) {
	var src, dst apistatus.LockIrregularObject
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 2051, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestLockIrregularObject_ErrorToV2(t *testing.T) {
	var e apistatus.LockIrregularObject
	st := e.ErrorToV2()
	require.EqualValues(t, 2051, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2051, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 2051, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}

func TestObjectAccessDenied_Error(t *testing.T) {
	var e apistatus.ObjectAccessDenied
	require.EqualError(t, e, "status: code = 2048 (object access denied)")
	e = apistatus.NewObjectAccessDeniedError("some reason")
	require.EqualError(t, e, "status: code = 2048 (object access denied, reason: some reason)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2048, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2048 (object access denied) message = any message")

	e2, err = apistatus.ErrorFromV2(&status.Status{Code: 2048, Message: "any message", Details: []*status.Status_Detail{{
		Id:    0,
		Value: []byte("some reason"),
	}}})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2048 (object access denied, reason: some reason) message = any message")
}

func TestObjectAccessDenied_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrObjectAccessDenied)
}

func TestObjectAccessDenied_As(t *testing.T) {
	var e, dst apistatus.ObjectAccessDenied
	require.ErrorAs(t, e, &dst)
	require.EqualValues(t, e.Reason(), dst.Reason())
	require.Equal(t, e.Error(), dst.Error())
	require.Equal(t, e.ErrorToV2(), dst.ErrorToV2())

	e = apistatus.NewObjectAccessDeniedError("some reason")
	require.ErrorAs(t, e, &dst)
	require.EqualValues(t, "some reason", dst.Reason())
	require.Equal(t, e.Error(), dst.Error())
	require.Equal(t, e.ErrorToV2(), dst.ErrorToV2())
}

func TestObjectAccessDenied_ErrorToV2(t *testing.T) {
	var e apistatus.ObjectAccessDenied
	st := e.ErrorToV2()
	require.EqualValues(t, 2048, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewObjectAccessDeniedError("some reason")
	st = e.ErrorToV2()
	require.EqualValues(t, 2048, st.Code)
	require.Zero(t, st.Message)
	require.Equal(t, []*status.Status_Detail{{
		Id:    0,
		Value: []byte("some reason"),
	}}, st.Details)
}

func TestObjectNotFound_Error(t *testing.T) {
	var e apistatus.ObjectNotFound
	require.EqualError(t, e, "status: code = 2049 (object not found)")
	e = apistatus.NewObjectNotFoundError(errors.New("some reason"))
	require.EqualError(t, e, "status: code = 2049 (object not found) message = some reason")
}

func TestObjectNotFound_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrObjectNotFound)
}

func TestObjectNotFound_As(t *testing.T) {
	var src, dst apistatus.ObjectNotFound
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	src = apistatus.NewObjectNotFoundError(errors.New("some reason"))
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "some reason", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestObjectNotFound_ErrorToV2(t *testing.T) {
	var e apistatus.ObjectNotFound
	st := e.ErrorToV2()
	require.EqualValues(t, 2049, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewObjectNotFoundError(errors.New("some reason"))
	st = e.ErrorToV2()
	require.EqualValues(t, 2049, st.Code)
	require.Equal(t, "some reason", st.Message)
	require.Zero(t, st.Details)
}

func TestObjectAlreadyRemoved_Error(t *testing.T) {
	var e apistatus.ObjectAlreadyRemoved
	require.EqualError(t, e, "status: code = 2052 (object already removed)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2052, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2052 (object already removed) message = any message")
}

func TestObjectAlreadyRemoved_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrObjectAlreadyRemoved)
}

func TestObjectAlreadyRemoved_As(t *testing.T) {
	var src, dst apistatus.ObjectAlreadyRemoved
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 2052, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestObjectAlreadyRemoved_ErrorToV2(t *testing.T) {
	var e apistatus.ObjectAlreadyRemoved
	st := e.ErrorToV2()
	require.EqualValues(t, 2052, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2052, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 2052, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}

func TestObjectOutOfRange_Error(t *testing.T) {
	var e apistatus.ObjectOutOfRange
	require.EqualError(t, e, "status: code = 2053 (out of range)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2053, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 2053 (out of range) message = any message")
}

func TestObjectOutOfRange_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrObjectOutOfRange)
}

func TestObjectOutOfRange_As(t *testing.T) {
	var src, dst apistatus.ObjectOutOfRange
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 2053, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestObjectOutOfRange_ErrorToV2(t *testing.T) {
	var e apistatus.ObjectOutOfRange
	st := e.ErrorToV2()
	require.EqualValues(t, 2053, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 2053, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 2053, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}
