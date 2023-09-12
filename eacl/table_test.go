package eacl_test

import (
	"crypto/rand"
	"testing"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("invalid records", func(t *testing.T) {
		require.Panics(t, func() { eacl.New(nil) })
		require.Panics(t, func() { eacl.New([]eacl.Record{}) })
		var unset eacl.Record
		require.Panics(t, func() { eacl.New([]eacl.Record{unset}) })
	})

	records := []eacl.Record{
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, eacl.NewTargetWithRole(eacl.RoleContainerOwner),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "my_photo.png"),
			eacl.NewFilter(eacl.HeaderFromRequest, "MODE", eacl.MatchStringNotEqual, "GATEWAY"),
		),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, eacl.NewTargetWithKey(test.RandomSigner(t).Public()),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "passport.pdf"),
		),
	}

	eACL := eacl.New(records)

	require.Equal(t, records, eACL.Records())
	_, ok := eACL.Container()
	require.False(t, ok)
}

func TestNewForContainer(t *testing.T) {
	cnr := cidtest.ID()

	t.Run("invalid records", func(t *testing.T) {
		require.Panics(t, func() { eacl.NewForContainer(cnr, nil) })
		require.Panics(t, func() { eacl.NewForContainer(cnr, []eacl.Record{}) })
		var unset eacl.Record
		require.Panics(t, func() { eacl.NewForContainer(cnr, []eacl.Record{unset}) })
	})

	records := []eacl.Record{
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, eacl.NewTargetWithRole(eacl.RoleContainerOwner),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "my_photo.png"),
			eacl.NewFilter(eacl.HeaderFromRequest, "MODE", eacl.MatchStringNotEqual, "GATEWAY"),
		),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, eacl.NewTargetWithKey(test.RandomSigner(t).Public()),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "passport.pdf"),
		),
	}

	eACL := eacl.NewForContainer(cnr, records)

	require.Equal(t, records, eACL.Records())
	cnrGot, ok := eACL.Container()
	require.True(t, ok)
	require.Equal(t, cnr, cnrGot)
}

func TestTable_LimitByContainer(t *testing.T) {
	anyValidRecords := []eacl.Record{
		eacl.NewRecord(eacl.ActionAllow, acl.OpObjectGet, eacl.NewTargetWithRole(eacl.RoleContainerOwner),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "my_photo.png"),
			eacl.NewFilter(eacl.HeaderFromRequest, "MODE", eacl.MatchStringNotEqual, "GATEWAY"),
		),
		eacl.NewRecord(eacl.ActionDeny, acl.OpObjectGet, eacl.NewTargetWithKey(test.RandomSigner(t).Public()),
			eacl.NewFilter(eacl.HeaderFromObject, "FileName", eacl.MatchStringEqual, "passport.pdf"),
		),
	}

	eACL := eacl.New(anyValidRecords)

	cnr := cidtest.ID()

	_, ok := eACL.Container()
	require.False(t, ok)

	eACL.LimitByContainer(cnr)

	cnrGot, ok := eACL.Container()
	require.True(t, ok)
	require.Equal(t, cnr, cnrGot)
}

func TestTableEncoding(t *testing.T) {
	tab := eacltest.Table(t)

	t.Run("binary", func(t *testing.T) {
		data := tab.Marshal()

		var tab2 eacl.Table
		require.NoError(t, tab2.Unmarshal(data))

		require.Equal(t, tab, tab2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := tab.MarshalJSON()
		require.NoError(t, err)

		var tab2 eacl.Table
		require.NoError(t, tab2.UnmarshalJSON(data))

		require.Equal(t, tab, tab2)
	})

	t.Run("NeoFS API protocol", func(t *testing.T) {
		sent := eacltest.Table(t)

		var msg1 v2acl.Table
		sent.WriteToV2(&msg1)

		var received eacl.Table
		require.NoError(t, received.ReadFromV2(msg1))
		require.Equal(t, sent, received)

		var msg2 v2acl.Table
		received.WriteToV2(&msg2)

		require.Equal(t, msg1, msg2)
	})
}

func TestTable_ReadFromV2(t *testing.T) {
	newValidTargetKeys := func() [][]byte {
		return [][]byte{
			neofscrypto.PublicKeyBytes(test.RandomSigner(t).Public()),
			neofscrypto.PublicKeyBytes(test.RandomSigner(t).Public()),
		}
	}

	newFullValidTarget := func() v2acl.Target {
		var target v2acl.Target
		target.SetRole(v2acl.RoleOthers)

		return target
	}

	newFullValidFilter := func() v2acl.HeaderFilter {
		var filter v2acl.HeaderFilter
		filter.SetHeaderType(v2acl.HeaderTypeRequest)
		filter.SetKey("any_key")
		filter.SetMatchType(v2acl.MatchTypeStringEqual)
		filter.SetValue("any_value")

		return filter
	}

	newFullValidRecord := func() v2acl.Record {
		var record v2acl.Record
		record.SetAction(v2acl.ActionAllow)
		record.SetOperation(v2acl.OperationPut)
		record.SetTargets([]v2acl.Target{newFullValidTarget()})
		record.SetFilters([]v2acl.HeaderFilter{newFullValidFilter()})

		return record
	}

	newFullValidMessage := func() v2acl.Table {
		var ver refs.Version
		ver.SetMajor(3)
		ver.SetMinor(4)

		bCnr := make([]byte, 32)
		_, err := rand.Read(bCnr)
		require.NoError(t, err)

		var cnr refs.ContainerID
		cnr.SetValue(bCnr)

		var m v2acl.Table
		m.SetVersion(&ver)
		m.SetContainerID(&cnr)
		m.SetRecords([]v2acl.Record{newFullValidRecord()})

		return m
	}

	for _, tc := range []struct {
		name    string
		corrupt func(table *v2acl.Table)
	}{
		{"missing version", func(table *v2acl.Table) {
			table.SetVersion(nil)
		}},
		{"invalid container reference", func(table *v2acl.Table) {
			var cnr refs.ContainerID
			cnr.SetValue(make([]byte, 31))
			table.SetContainerID(&cnr)
		}},
		{"missing records", func(table *v2acl.Table) {
			table.SetRecords([]v2acl.Record{})
		}},
		{"missing target in one record", func(table *v2acl.Table) {
			record := newFullValidRecord()
			record.SetTargets([]v2acl.Target{})
			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
		{"missing target in one record", func(table *v2acl.Table) {
			record := newFullValidRecord()
			record.SetTargets([]v2acl.Target{})

			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
		{"missing both role and keys in one target", func(table *v2acl.Table) {
			target := newFullValidTarget()
			target.SetRole(0)
			target.SetKeys([][]byte{})

			record := newFullValidRecord()
			record.SetTargets([]v2acl.Target{newFullValidTarget(), target})

			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
		{"both role and keys are set in one target", func(table *v2acl.Table) {
			target := newFullValidTarget()
			target.SetRole(v2acl.RoleUser)
			target.SetKeys(newValidTargetKeys())

			record := newFullValidRecord()
			record.SetTargets([]v2acl.Target{newFullValidTarget(), target})

			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
		{"empty key in one target", func(table *v2acl.Table) {
			keys := newValidTargetKeys()
			keys[len(keys)-1] = nil

			target := newFullValidTarget()
			target.SetRole(0)
			target.SetKeys(keys)

			record := newFullValidRecord()
			record.SetTargets([]v2acl.Target{newFullValidTarget(), target})

			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
		{"missing key in one filter", func(table *v2acl.Table) {
			filter := newFullValidFilter()
			filter.SetKey("")

			record := newFullValidRecord()
			record.SetFilters([]v2acl.HeaderFilter{filter})

			table.SetRecords([]v2acl.Record{newFullValidRecord(), record})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newFullValidMessage()

			var eACL eacl.Table
			require.NoError(t, eACL.ReadFromV2(m))

			tc.corrupt(&m)

			require.Error(t, eACL.ReadFromV2(m))
		})
	}

	t.Run("unset container", func(t *testing.T) {
		m := newFullValidMessage()
		m.SetContainerID(nil)

		var eACL eacl.Table
		require.NoError(t, eACL.ReadFromV2(m))

		_, ok := eACL.Container()
		require.False(t, ok)

		var m2 v2acl.Table
		eACL.WriteToV2(&m2)

		require.Nil(t, m2.GetContainerID())
	})
}
