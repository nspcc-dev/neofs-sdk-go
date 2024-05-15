package apistatus_test

import (
	"errors"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestInternalServerError_Error(t *testing.T) {
	var e apistatus.InternalServerError
	require.EqualError(t, e, "status: code = 1024 (internal server error)")
	e = apistatus.NewInternalServerError(errors.New("some reason"))
	require.EqualError(t, e, "status: code = 1024 (internal server error) message = some reason")
}

func TestInternalServerError_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrServerInternal)
}

func TestInternalServerError_As(t *testing.T) {
	var src, dst apistatus.InternalServerError
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	src = apistatus.NewInternalServerError(errors.New("some reason"))
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "some reason", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestInternalServerError_ErrorToV2(t *testing.T) {
	var e apistatus.InternalServerError
	st := e.ErrorToV2()
	require.EqualValues(t, 1024, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewInternalServerError(errors.New("some reason"))
	st = e.ErrorToV2()
	require.EqualValues(t, 1024, st.Code)
	require.Equal(t, "some reason", st.Message)
	require.Zero(t, st.Details)
}

func TestWrongNetMagic_Error(t *testing.T) {
	var e apistatus.WrongNetMagic
	require.EqualError(t, e, "status: code = 1025 (wrong network magic)")
	e = apistatus.NewWrongNetMagicError(4594136436)
	require.EqualError(t, e, "status: code = 1025 (wrong network magic, expected 4594136436)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 1025, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 1025 (wrong network magic) message = any message")

	e2, err = apistatus.ErrorFromV2(&status.Status{Code: 1025, Message: "any message", Details: []*status.Status_Detail{{
		Id:    0,
		Value: []byte{0, 0, 0, 1, 107, 18, 46, 11},
	}}})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 1025 (wrong network magic, expected 6091320843) message = any message")
}

func TestWrongNetMagic_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrWrongNetMagic)
}

func TestWrongNetMagic_As(t *testing.T) {
	var e, dst apistatus.WrongNetMagic
	require.ErrorAs(t, e, &dst)
	require.EqualValues(t, e.CorrectMagic(), dst.CorrectMagic())
	require.Equal(t, e.Error(), dst.Error())
	require.Equal(t, e.ErrorToV2(), dst.ErrorToV2())

	e = apistatus.NewWrongNetMagicError(3254368)
	require.ErrorAs(t, e, &dst)
	require.EqualValues(t, 3254368, dst.CorrectMagic())
	require.Equal(t, e.Error(), dst.Error())
	require.Equal(t, e.ErrorToV2(), dst.ErrorToV2())
}

func TestWrongNetMagic_ErrorToV2(t *testing.T) {
	var e apistatus.WrongNetMagic
	st := e.ErrorToV2()
	require.EqualValues(t, 1025, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewWrongNetMagicError(6091320843)
	st = e.ErrorToV2()
	require.EqualValues(t, 1025, st.Code)
	require.Zero(t, st.Message)
	require.Equal(t, []*status.Status_Detail{{
		Id:    0,
		Value: []byte{0, 0, 0, 1, 107, 18, 46, 11},
	}}, st.Details)
}

func TestSignatureVerificationFailure_Error(t *testing.T) {
	var e apistatus.SignatureVerificationFailure
	require.EqualError(t, e, "status: code = 1026 (signature verification failed)")
	e = apistatus.NewSignatureVerificationFailure(errors.New("some reason"))
	require.EqualError(t, e, "status: code = 1026 (signature verification failed) message = some reason")
}

func TestSignatureVerificationFailure_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrSignatureVerification)
}

func TestSignatureVerificationFailure_As(t *testing.T) {
	var src, dst apistatus.SignatureVerificationFailure
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
	src = apistatus.NewSignatureVerificationFailure(errors.New("some reason"))
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "some reason", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestSignatureVerificationFailure_ErrorToV2(t *testing.T) {
	var e apistatus.SignatureVerificationFailure
	st := e.ErrorToV2()
	require.EqualValues(t, 1026, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewSignatureVerificationFailure(errors.New("some reason"))
	st = e.ErrorToV2()
	require.EqualValues(t, 1026, st.Code)
	require.Equal(t, "some reason", st.Message)
	require.Zero(t, st.Details)
}

func TestNodeUnderMaintenance_Error(t *testing.T) {
	var e apistatus.NodeUnderMaintenance
	require.EqualError(t, e, "status: code = 1027 (node is under maintenance)")

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 1027, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	require.EqualError(t, e, "status: code = 1027 (node is under maintenance) message = any message")
}

func TestNodeUnderMaintenance_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrNodeUnderMaintenance)
}

func TestNodeUnderMaintenance_As(t *testing.T) {
	var src, dst apistatus.NodeUnderMaintenance
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	e, err := apistatus.ErrorFromV2(&status.Status{Code: 1027, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e, &src)
	require.ErrorAs(t, src, &dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestNodeUnderMaintenance_ErrorToV2(t *testing.T) {
	var e apistatus.NodeUnderMaintenance
	st := e.ErrorToV2()
	require.EqualValues(t, 1027, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e2, err := apistatus.ErrorFromV2(&status.Status{Code: 1027, Message: "any message"})
	require.NoError(t, err)
	require.ErrorAs(t, e2, &e)
	st = e.ErrorToV2()
	require.EqualValues(t, 1027, st.Code)
	require.Equal(t, "any message", st.Message)
	require.Zero(t, st.Details)
}
