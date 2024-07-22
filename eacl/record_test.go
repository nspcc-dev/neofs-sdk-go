package eacl_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	record := eacl.NewRecord()
	record.SetOperation(eacl.OperationRange)
	record.SetAction(eacl.ActionAllow)
	record.AddFilter(eacl.HeaderFromRequest, eacl.MatchStringEqual, "A", "B")
	record.AddFilter(eacl.HeaderFromRequest, eacl.MatchStringNotEqual, "C", "D")

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleSystem)
	eacl.AddRecordTarget(record, target)

	v2 := record.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, v2acl.OperationRange, v2.GetOperation())
	require.Equal(t, v2acl.ActionAllow, v2.GetAction())
	require.Len(t, v2.GetFilters(), len(record.Filters()))
	require.Len(t, v2.GetTargets(), len(record.Targets()))

	newRecord := eacl.NewRecordFromV2(v2)
	require.Equal(t, record, newRecord)

	t.Run("create record", func(t *testing.T) {
		record := eacl.CreateRecord(eacl.ActionAllow, eacl.OperationGet)
		require.Equal(t, eacl.ActionAllow, record.Action())
		require.Equal(t, eacl.OperationGet, record.Operation())
	})

	t.Run("new from nil v2 record", func(t *testing.T) {
		require.Equal(t, new(eacl.Record), eacl.NewRecordFromV2(nil))
	})
}

func TestAddFormedTarget(t *testing.T) {
	k1, k2 := randomPublicKey(t), randomPublicKey(t)
	var r eacl.Record

	eacl.AddFormedTarget(&r, eacl.RoleUnspecified, *k1, *k2)
	require.Len(t, r.Targets(), 1)
	require.Zero(t, r.Targets()[0].Role())
	require.Len(t, r.Targets()[0].BinaryKeys(), 2)
	require.Equal(t, elliptic.MarshalCompressed(k1.Curve, k1.X, k1.Y), r.Targets()[0].BinaryKeys()[0])
	require.Equal(t, elliptic.MarshalCompressed(k2.Curve, k2.X, k2.Y), r.Targets()[0].BinaryKeys()[1])

	role := eacl.Role(rand.Uint32())
	eacl.AddFormedTarget(&r, role)
	require.Len(t, r.Targets(), 2)
	require.Equal(t, role, r.Targets()[1].Role())
	require.Zero(t, r.Targets()[1].BinaryKeys())
}

func TestRecord_AddFilter(t *testing.T) {
	filters := []eacl.Filter{
		eacl.NewObjectPropertyFilter("some name", eacl.MatchStringEqual, "ContainerID"),
		eacl.NewObjectPropertyFilter("X-Header-Name", eacl.MatchStringNotEqual, "X-Header-Value"),
	}

	r := eacl.NewRecord()
	for _, filter := range filters {
		r.AddFilter(filter.From(), filter.Matcher(), filter.Key(), filter.Value())
	}

	require.Equal(t, filters, r.Filters())
}

func TestRecordEncoding(t *testing.T) {
	r := eacl.NewRecord()
	r.SetOperation(eacl.OperationHead)
	r.SetAction(eacl.ActionDeny)
	r.AddObjectAttributeFilter(eacl.MatchStringEqual, "key", "value")
	eacl.AddFormedTarget(r, eacl.RoleSystem, *randomPublicKey(t))

	t.Run("binary", func(t *testing.T) {
		r2 := eacl.NewRecord()
		require.NoError(t, r2.Unmarshal(r.Marshal()))

		require.Equal(t, r, r2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := r.MarshalJSON()
		require.NoError(t, err)

		r2 := eacl.NewRecord()
		require.NoError(t, r2.UnmarshalJSON(data))

		require.Equal(t, r, r2)
	})
}

func TestRecord_ToV2(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		record := eacl.NewRecord()

		// check initial values
		require.Zero(t, record.Operation())
		require.Zero(t, record.Action())
		require.Nil(t, record.Targets())
		require.Nil(t, record.Filters())

		// convert to v2 message
		recordV2 := record.ToV2()

		require.Equal(t, v2acl.OperationUnknown, recordV2.GetOperation())
		require.Equal(t, v2acl.ActionUnknown, recordV2.GetAction())
		require.Nil(t, recordV2.GetTargets())
		require.Nil(t, recordV2.GetFilters())
	})
}

func TestReservedRecords(t *testing.T) {
	var (
		v       = versiontest.Version()
		oid     = oidtest.ID()
		cid     = cidtest.ID()
		ownerid = usertest.ID()
		h       = checksumtest.Checksum()
		typ     = new(object.Type)
	)

	testSuit := []struct {
		f     func(r *eacl.Record)
		key   string
		value string
	}{
		{
			f:     func(r *eacl.Record) { r.AddObjectAttributeFilter(eacl.MatchStringEqual, "foo", "bar") },
			key:   "foo",
			value: "bar",
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectVersionFilter(eacl.MatchStringEqual, &v) },
			key:   v2acl.FilterObjectVersion,
			value: v.String(),
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectIDFilter(eacl.MatchStringEqual, oid) },
			key:   v2acl.FilterObjectID,
			value: oid.EncodeToString(),
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectContainerIDFilter(eacl.MatchStringEqual, cid) },
			key:   v2acl.FilterObjectContainerID,
			value: cid.EncodeToString(),
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectOwnerIDFilter(eacl.MatchStringEqual, &ownerid) },
			key:   v2acl.FilterObjectOwnerID,
			value: ownerid.EncodeToString(),
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectCreationEpoch(eacl.MatchStringEqual, 100) },
			key:   v2acl.FilterObjectCreationEpoch,
			value: "100",
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectPayloadLengthFilter(eacl.MatchStringEqual, 5000) },
			key:   v2acl.FilterObjectPayloadLength,
			value: "5000",
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectPayloadHashFilter(eacl.MatchStringEqual, h) },
			key:   v2acl.FilterObjectPayloadHash,
			value: h.String(),
		},
		{
			f:     func(r *eacl.Record) { r.AddObjectHomomorphicHashFilter(eacl.MatchStringEqual, h) },
			key:   v2acl.FilterObjectHomomorphicHash,
			value: h.String(),
		},
		{
			f: func(r *eacl.Record) {
				require.True(t, typ.DecodeString("REGULAR"))
				r.AddObjectTypeFilter(eacl.MatchStringEqual, *typ)
			},
			key:   v2acl.FilterObjectType,
			value: "REGULAR",
		},
		{
			f: func(r *eacl.Record) {
				require.True(t, typ.DecodeString("TOMBSTONE"))
				r.AddObjectTypeFilter(eacl.MatchStringEqual, *typ)
			},
			key:   v2acl.FilterObjectType,
			value: "TOMBSTONE",
		},
		{
			f: func(r *eacl.Record) {
				require.True(t, typ.DecodeString("STORAGE_GROUP"))
				r.AddObjectTypeFilter(eacl.MatchStringEqual, *typ)
			},
			key:   v2acl.FilterObjectType,
			value: "STORAGE_GROUP",
		},
	}

	for n, testCase := range testSuit {
		desc := fmt.Sprintf("case #%d", n)
		record := eacl.NewRecord()
		testCase.f(record)
		require.Len(t, record.Filters(), 1, desc)
		f := record.Filters()[0]
		require.Equal(t, f.Key(), testCase.key, desc)
		require.Equal(t, f.Value(), testCase.value, desc)
	}
}

func randomPublicKey(t *testing.T) *ecdsa.PublicKey {
	p, err := keys.NewPrivateKey()
	require.NoError(t, err)
	return &p.PrivateKey.PublicKey
}

func TestRecord_SetFilters(t *testing.T) {
	fs := []eacl.Filter{
		eacl.ConstructFilter(eacl.FilterHeaderType(rand.Uint32()), "key1", eacl.Match(rand.Uint32()), "val1"),
		eacl.ConstructFilter(eacl.FilterHeaderType(rand.Uint32()), "key2", eacl.Match(rand.Uint32()), "val2"),
	}
	var r eacl.Record
	require.Zero(t, r.Filters())
	r.SetFilters(fs)
	require.Equal(t, fs, r.Filters())
}

func TestConstructRecord(t *testing.T) {
	a := eacl.Action(rand.Uint32())
	op := eacl.Operation(rand.Uint32())
	ts := []eacl.Target{
		eacl.NewTargetByRole(eacl.Role(rand.Uint32())),
		eacl.NewTargetByAccounts(usertest.IDs(5)),
	}
	fs := []eacl.Filter{
		eacl.ConstructFilter(eacl.FilterHeaderType(rand.Uint32()), "key1", eacl.Match(rand.Uint32()), "val1"),
		eacl.ConstructFilter(eacl.FilterHeaderType(rand.Uint32()), "key2", eacl.Match(rand.Uint32()), "val2"),
	}

	r := eacl.ConstructRecord(a, op, ts)
	require.Equal(t, a, r.Action())
	require.Equal(t, op, r.Operation())
	require.Equal(t, ts, r.Targets())
	require.Zero(t, r.Filters())
	r = eacl.ConstructRecord(a, op, ts, fs...)
	require.Equal(t, fs, r.Filters())
}
