package eacl_test

import (
	"bytes"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestRecord_Action(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Action())

	r.SetAction(13)
	require.EqualValues(t, 13, r.Action())
	r.SetAction(42)
	require.EqualValues(t, 42, r.Action())
}

func TestRecord_Operation(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Operation())

	r.SetOperation(13)
	require.EqualValues(t, 13, r.Operation())
	r.SetOperation(42)
	require.EqualValues(t, 42, r.Operation())
}

func TestRecord_Filters(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Filters())

	fs := make([]eacl.Filter, 2)
	fs[0].SetKey("key1")
	fs[1].SetKey("key2")
	fs[0].SetValue("val1")
	fs[1].SetValue("val2")
	fs[0].SetAttributeType(1)
	fs[1].SetAttributeType(2)
	fs[0].SetMatcher(3)
	fs[1].SetMatcher(4)

	r.SetFilters(fs)
	require.Equal(t, fs, r.Filters())

	fs = make([]eacl.Filter, 2)
	fs[0].SetKey("key3")
	fs[1].SetKey("key4")
	fs[0].SetValue("val3")
	fs[1].SetValue("val4")
	fs[0].SetAttributeType(4)
	fs[1].SetAttributeType(6)
	fs[0].SetMatcher(7)
	fs[1].SetMatcher(8)

	r.SetFilters(fs)
	require.Equal(t, fs, r.Filters())
}

func TestRecord_Targets(t *testing.T) {
	var r eacl.Record
	require.Zero(t, r.Targets())

	ts := make([]eacl.Target, 2)
	ts[0].SetRole(1)
	ts[1].SetRole(2)
	ts[0].SetPublicKeys([][]byte{[]byte("key1"), []byte("key2")})
	ts[1].SetPublicKeys([][]byte{[]byte("key3"), []byte("key4")})

	r.SetTargets(ts)
	require.Equal(t, ts, r.Targets())

	ts = make([]eacl.Target, 2)
	ts[0].SetRole(3)
	ts[1].SetRole(4)
	ts[0].SetPublicKeys([][]byte{[]byte("key5"), []byte("key6")})
	ts[1].SetPublicKeys([][]byte{[]byte("key7"), []byte("key8")})

	r.SetTargets(ts)
	require.Equal(t, ts, r.Targets())
}

func TestRecord_CopyTo(t *testing.T) {
	ts := eacltest.NTargets(2)
	ts[0].SetPublicKeys([][]byte{[]byte("key1"), []byte("key2")})
	src := eacltest.Record()
	src.SetTargets(ts)

	var dst eacl.Record
	src.CopyTo(&dst)
	require.Equal(t, src, dst)

	originKey := src.Filters()[0].Key()
	src.Filters()[0].SetKey(originKey + "_extra")
	require.Equal(t, originKey+"_extra", src.Filters()[0].Key())
	require.Equal(t, originKey, dst.Filters()[0].Key())

	originPubKey := bytes.Clone(src.Targets()[0].PublicKeys()[0])
	src.Targets()[0].PublicKeys()[0][0]++
	require.Equal(t, originPubKey, dst.Targets()[0].PublicKeys()[0])
}
