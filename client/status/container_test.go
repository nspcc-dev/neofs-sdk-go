package apistatus_test

import (
	"errors"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestContainerNotFound_Error(t *testing.T) {
	var e apistatus.ContainerNotFound
	require.EqualError(t, e, "status: code = 3072 (container not found)")
	e = apistatus.NewContainerNotFoundError(errors.New("some reason"))
	require.EqualError(t, e, "status: code = 3072 (container not found) message = some reason")
}

func TestContainerNotFound_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrContainerNotFound)
}

func TestContainerNotFound_As(t *testing.T) {
	var src, dst apistatus.ContainerNotFound
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	src = apistatus.NewContainerNotFoundError(errors.New("some reason"))
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "some reason", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestContainerNotFound_ErrorToV2(t *testing.T) {
	var e apistatus.ContainerNotFound
	st := e.ErrorToV2()
	require.EqualValues(t, 3072, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewContainerNotFoundError(errors.New("some reason"))
	st = e.ErrorToV2()
	require.EqualValues(t, 3072, st.Code)
	require.Equal(t, "some reason", st.Message)
	require.Zero(t, st.Details)
}

func TestEACLNotFound_Error(t *testing.T) {
	var e apistatus.EACLNotFound
	require.EqualError(t, e, "status: code = 3073 (eACL not found)")
	e = apistatus.NewEACLNotFoundError(errors.New("some reason"))
	require.EqualError(t, e, "status: code = 3073 (eACL not found) message = some reason")
}

func TestEACLNotFound_Is(t *testing.T) {
	assertErrorIs(t, apistatus.ErrEACLNotFound)
}

func TestEACLNotFound_As(t *testing.T) {
	var src, dst apistatus.EACLNotFound
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())

	src = apistatus.NewEACLNotFoundError(errors.New("some reason"))
	require.ErrorAs(t, src, &dst)
	require.EqualValues(t, "some reason", dst)
	require.Equal(t, src.Error(), dst.Error())
	require.Equal(t, src.ErrorToV2(), dst.ErrorToV2())
}

func TestEACLNotFound_ErrorToV2(t *testing.T) {
	var e apistatus.EACLNotFound
	st := e.ErrorToV2()
	require.EqualValues(t, 3073, st.Code)
	require.Zero(t, st.Message)
	require.Zero(t, st.Details)

	e = apistatus.NewEACLNotFoundError(errors.New("some reason"))
	st = e.ErrorToV2()
	require.EqualValues(t, 3073, st.Code)
	require.Equal(t, "some reason", st.Message)
	require.Zero(t, st.Details)
}
