package eacl_test

import (
	"encoding/json"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

func TestRecord_Marshal(t *testing.T) {
	for i := range anyValidRecords {
		require.Equal(t, anyValidBinRecords[i], anyValidRecords[i].Marshal(), i)
	}
}

func TestRecord_Unmarshal(t *testing.T) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := new(eacl.Record).Unmarshal([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var r eacl.Record
	for i := range anyValidBinRecords {
		require.NoError(t, r.Unmarshal(anyValidBinRecords[i]), i)
		require.EqualValues(t, anyValidRecords[i], r, i)
	}
}

func TestRecord_MarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(eacl.Record).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var r1, r2 eacl.Record
	for i := range anyValidRecords {
		b, err := anyValidRecords[i].MarshalJSON()
		require.NoError(t, err, i)
		require.NoError(t, r1.UnmarshalJSON(b), i)
		require.Equal(t, anyValidRecords[i], r1, i)

		b, err = json.Marshal(anyValidRecords[i])
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &r2), i)
		require.Equal(t, anyValidRecords[i], r2, i)
	}
}

func TestRecord_UnmarshalJSON(t *testing.T) {
	var r1, r2 eacl.Record
	for i := range anyValidJSONRecords {
		require.NoError(t, r1.UnmarshalJSON([]byte(anyValidJSONRecords[i])), i)
		require.Equal(t, anyValidRecords[i], r1, i)

		require.NoError(t, json.Unmarshal([]byte(anyValidJSONRecords[i]), &r2), i)
		require.Equal(t, r1, r2, i)
	}
}

func TestRecord_SetAction(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Action())
	r.SetAction(anyValidAction)
	require.Equal(t, anyValidAction, r.Action())
}

func TestRecord_Comment(t *testing.T) {
	t.Run("invalid UTF-8", func(t *testing.T) {
		t.Run("setter", func(t *testing.T) {
			var r eacl.Record
			err := r.SetComment(string([]byte{0xff}))
			require.ErrorContains(t, err, "invalid UTF-8")
			require.Empty(t, r.Comment())
		})
		t.Run("binary", func(t *testing.T) {
			// Field #5 with a one-byte invalid UTF-8 payload.
			err := new(eacl.Record).Unmarshal([]byte{42, 1, 0xff})
			require.ErrorContains(t, err, "invalid UTF-8")
		})
	})

	t.Run("zero byte", func(t *testing.T) {
		t.Run("setter", func(t *testing.T) {
			var r eacl.Record
			err := r.SetComment(string([]byte{'a', 0, 'b'}))
			require.ErrorContains(t, err, "comment contains zero byte")
			require.Empty(t, r.Comment())
		})
		t.Run("binary", func(t *testing.T) {
			err := new(eacl.Record).Unmarshal([]byte{42, 3, 'a', 0, 'b'})
			require.ErrorContains(t, err, "comment contains zero byte")
		})
		t.Run("JSON", func(t *testing.T) {
			err := new(eacl.Record).UnmarshalJSON([]byte(`{"comment":"a\u0000b"}`))
			require.ErrorContains(t, err, "comment contains zero byte")
		})
	})

	const comment = "Application rule"

	var r eacl.Record
	require.Empty(t, r.Comment())
	require.NoError(t, r.SetComment(comment))
	require.Equal(t, comment, r.Comment())

	var cp eacl.Record
	r.CopyTo(&cp)
	require.Equal(t, comment, cp.Comment())

	var fromBinary eacl.Record
	require.NoError(t, fromBinary.Unmarshal(r.Marshal()))
	require.Equal(t, comment, fromBinary.Comment())

	b, err := r.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"action":"ACTION_UNSPECIFIED","operation":"OPERATION_UNSPECIFIED","filters":[],"targets":[],"comment":"Application rule"}`, string(b))

	var fromJSON eacl.Record
	require.NoError(t, fromJSON.UnmarshalJSON(b))
	require.Equal(t, comment, fromJSON.Comment())
}

func TestRecord_SetOperation(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Operation())
	r.SetOperation(anyValidOp)
	require.Equal(t, anyValidOp, r.Operation())
}

func TestRecord_SetTargets(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Targets())
	r.SetTargets(anyValidTargets...)
	require.Equal(t, anyValidTargets, r.Targets())
}

func TestRecord_SetFilters(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Filters())
	r.SetFilters(anyValidFilters)
	require.Equal(t, anyValidFilters, r.Filters())
}

func TestConstructRecord(t *testing.T) {
	r := eacl.ConstructRecord(anyValidAction, anyValidOp, anyValidTargets)
	require.Equal(t, anyValidAction, r.Action())
	require.Equal(t, anyValidOp, r.Operation())
	require.Equal(t, anyValidTargets, r.Targets())
	require.Zero(t, r.Filters())
	r = eacl.ConstructRecord(anyValidAction, anyValidOp, anyValidTargets, anyValidFilters...)
	require.Equal(t, anyValidFilters, r.Filters())
}
