package apistatus_test

import (
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestServerInternal_Message(t *testing.T) {
	const msg = "some message"

	var st apistatus.ServerInternal

	require.Empty(t, st.Message())
	require.Empty(t, apistatus.FromError(st).Message)

	st.SetMessage(msg)

	require.Equal(t, msg, st.Message())
	require.Equal(t, msg, apistatus.FromError(st).Message)
}

func TestWrongMagicNumber_CorrectMagic(t *testing.T) {
	const magic = 1337

	var st apistatus.WrongMagicNumber

	res, ok := st.CorrectMagic()
	require.Zero(t, res)
	require.Zero(t, ok)

	st.WriteCorrectMagic(magic)

	res, ok = st.CorrectMagic()
	require.EqualValues(t, magic, res)
	require.EqualValues(t, 1, ok)

	// corrupt the value
	m := apistatus.FromError(st)
	require.Len(t, m.Details, 1)
	m.Details[0].Value = []byte{1, 2, 3} // any slice with len != 8

	_, ok = st.CorrectMagic()
	require.EqualValues(t, -1, ok)
}

func TestSignatureVerification(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		var st apistatus.SignatureVerification

		require.Empty(t, st.Message())
	})

	t.Run("custom message", func(t *testing.T) {
		var st apistatus.SignatureVerification
		msg := "some message"

		st.SetMessage(msg)

		m := apistatus.FromError(st)

		require.Equal(t, msg, st.Message())
		require.Equal(t, msg, m.Message)
	})

	t.Run("proto", func(t *testing.T) {
		var st apistatus.SignatureVerification

		m := apistatus.FromError(st)

		require.Equal(t, "signature verification failed", m.Message)

		msg := "some other msg"

		st.SetMessage(msg)

		m = apistatus.FromError(st)

		require.Equal(t, msg, m.Message)
	})
}

func TestNodeUnderMaintenance(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		var st apistatus.NodeUnderMaintenance

		require.Empty(t, st.Message())
	})

	t.Run("custom message", func(t *testing.T) {
		var st apistatus.NodeUnderMaintenance
		msg := "some message"

		st.SetMessage(msg)

		m := apistatus.FromError(st)

		require.Equal(t, msg, st.Message())
		require.Equal(t, msg, m.Message)
	})

	t.Run("proto", func(t *testing.T) {
		var st apistatus.NodeUnderMaintenance

		m := apistatus.FromError(st)

		require.Equal(t, "node is under maintenance", m.Message)

		msg := "some other msg"

		st.SetMessage(msg)

		m = apistatus.FromError(st)

		require.Equal(t, msg, m.Message)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		var st apistatus.BadRequest

		require.Empty(t, st.Message())
	})

	t.Run("custom message", func(t *testing.T) {
		var st apistatus.BadRequest
		msg := "some message"

		st.SetMessage(msg)

		m := apistatus.FromError(st)

		require.Equal(t, msg, st.Message())
		require.Equal(t, msg, m.Message)
	})

	t.Run("proto", func(t *testing.T) {
		var st apistatus.BadRequest

		m := apistatus.FromError(st)

		require.Equal(t, "bad request", m.Message)

		msg := "some other msg"

		st.SetMessage(msg)

		m = apistatus.FromError(st)

		require.Equal(t, msg, m.Message)
	})
}

func TestBusy(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		var st apistatus.Busy

		require.Empty(t, st.Message())
	})

	t.Run("custom message", func(t *testing.T) {
		var st apistatus.Busy
		msg := "some message"

		st.SetMessage(msg)

		m := apistatus.FromError(st)

		require.Equal(t, msg, st.Message())
		require.Equal(t, msg, m.Message)
	})

	t.Run("proto", func(t *testing.T) {
		var st apistatus.Busy

		m := apistatus.FromError(st)

		require.Equal(t, "busy, retry later", m.Message)

		msg := "some other msg"

		st.SetMessage(msg)

		m = apistatus.FromError(st)

		require.Equal(t, msg, m.Message)
	})
}
