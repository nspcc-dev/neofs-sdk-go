package eacl_test

import (
	"encoding/json"
	"math/rand/v2"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

func TestAddFormedTarget(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Targets())

	eacl.AddFormedTarget(&r, eacl.RoleUnspecified, anyECDSAPublicKeys...)
	require.Len(t, r.Targets(), 1)
	require.Zero(t, r.Targets()[0].Role())
	require.Equal(t, anyValidECDSABinPublicKeys, r.Targets()[0].BinaryKeys())

	role := eacl.Role(rand.Int32())
	eacl.AddFormedTarget(&r, role)
	require.Len(t, r.Targets(), 2)
	require.Equal(t, role, r.Targets()[1].Role())
	require.Zero(t, r.Targets()[1].BinaryKeys())
}

func TestRecord_AddFilter(t *testing.T) {
	r := eacl.NewRecord()
	for _, filter := range anyValidFilters {
		r.AddFilter(filter.From(), filter.Matcher(), filter.Key(), filter.Value())
	}

	require.Equal(t, anyValidFilters, r.Filters())
}

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

func assertSingleObjectFilter(t testing.TB, r eacl.Record, k string, m eacl.Match, v string) {
	require.Len(t, r.Filters(), 1)
	require.EqualValues(t, 2, r.Filters()[0].From())
	require.Equal(t, k, r.Filters()[0].Key())
	require.EqualValues(t, m, r.Filters()[0].Matcher())
	require.Equal(t, v, r.Filters()[0].Value())
}

func TestRecord_AddObjectAttributeFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectAttributeFilter(anyValidMatcher, "foo", "bar")
	assertSingleObjectFilter(t, r, "foo", anyValidMatcher, "bar")
}

func TestRecord_AddObjectIDFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectIDFilter(anyValidMatcher, anyValidObjectID)
	assertSingleObjectFilter(t, r, "$Object:objectID", anyValidMatcher, anyValidObjectIDString)
}

func TestRecord_AddObjectVersionFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectVersionFilter(anyValidMatcher, &anyValidProtoVersion)
	assertSingleObjectFilter(t, r, "$Object:version", anyValidMatcher, anyValidProtoVersionString)
}

func TestRecord_AddObjectContainerIDFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectContainerIDFilter(anyValidMatcher, anyValidContainerID)
	assertSingleObjectFilter(t, r, "$Object:containerID", anyValidMatcher, anyValidContainerIDString)
}

func TestRecord_AddObjectOwnerIDFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectOwnerIDFilter(anyValidMatcher, &anyValidUserID)
	assertSingleObjectFilter(t, r, "$Object:ownerID", anyValidMatcher, anyValidUserIDString)
}

func TestRecord_AddObjectCreationEpoch(t *testing.T) {
	var r eacl.Record
	r.AddObjectCreationEpoch(anyValidMatcher, 673984)
	assertSingleObjectFilter(t, r, "$Object:creationEpoch", anyValidMatcher, "673984")
}

func TestRecord_AddObjectPayloadLengthFilter(t *testing.T) {
	var r eacl.Record
	r.AddObjectPayloadLengthFilter(anyValidMatcher, 74928)
	assertSingleObjectFilter(t, r, "$Object:payloadLength", anyValidMatcher, "74928")
}

func TestRecord_AddObjectPayloadHashFilter(t *testing.T) {
	for i := range anyValidChecksums {
		var r eacl.Record
		r.AddObjectPayloadHashFilter(anyValidMatcher, anyValidChecksums[i])
		assertSingleObjectFilter(t, r, "$Object:payloadHash", anyValidMatcher, anyValidStringChecksums[i])
	}
}

func TestRecord_AddObjectHomomorphicHashFilter(t *testing.T) {
	for i := range anyValidChecksums {
		var r eacl.Record
		r.AddObjectHomomorphicHashFilter(anyValidMatcher, anyValidChecksums[i])
		assertSingleObjectFilter(t, r, "$Object:homomorphicHash", anyValidMatcher, anyValidStringChecksums[i])
	}
}

func TestRecord_AddObjectTypeFilter(t *testing.T) {
	for i := range anyValidObjectTypes {
		var r eacl.Record
		r.AddObjectTypeFilter(anyValidMatcher, anyValidObjectTypes[i])
		assertSingleObjectFilter(t, r, "$Object:objectType", anyValidMatcher, anyValidStringObjectTypes[i])
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

func TestCreateRecord(t *testing.T) {
	r := eacl.CreateRecord(anyValidAction, anyValidOp)
	require.Equal(t, anyValidAction, r.Action())
	require.Equal(t, anyValidOp, r.Operation())
	require.Empty(t, r.Targets())
	require.Empty(t, r.Filters())
}
