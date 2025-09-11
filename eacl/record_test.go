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
