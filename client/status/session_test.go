package apistatus_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestSessionTokenNotFound_Error(t *testing.T) {
	var e apistatus.SessionTokenNotFound
	require.EqualError(t, e, "status: code = 4096 (session token not found)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 4096, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 4096 (session token not found) message = any message")
}

func TestSessionTokenNotFound_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrSessionTokenNotFound)
}

func TestSessionTokenNotFound_As(t *testing.T) {
	var src, dst apistatus.SessionTokenNotFound
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 4096, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestSessionTokenNotFound_ErrorToV2(t *testing.T) {
	var e apistatus.SessionTokenNotFound
	st := e.ErrorToV2()
	require.EqualValues(t, 4096, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 4096, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 4096, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}

func TestSessionTokenExpired_Error(t *testing.T) {
	var e apistatus.SessionTokenExpired
	require.EqualError(t, e, "status: code = 4097 (session token has expired)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 4097, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 4097 (session token has expired) message = any message")
}

func TestSessionTokenExpired_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrSessionTokenExpired)
}

func TestSessionTokenExpired_As(t *testing.T) {
	var src, dst apistatus.SessionTokenExpired
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 4097, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestSessionTokenExpired_ErrorToV2(t *testing.T) {
	var e apistatus.SessionTokenExpired
	st := e.ErrorToV2()
	require.EqualValues(t, 4097, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 4097, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 4097, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}
