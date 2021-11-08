package eacl_test

import (
	"crypto/sha256"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	var v version.Version

	sha := sha256.Sum256([]byte("container id"))
	id := cidtest.GenerateIDWithChecksum(sha)

	v.SetMajor(3)
	v.SetMinor(2)

	table := eacl.NewTable()
	table.SetVersion(v)
	table.SetCID(id)
	table.AddRecord(eacl.CreateRecord(eacl.ActionAllow, eacl.OperationPut))

	v2 := table.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, uint32(3), v2.GetVersion().GetMajor())
	require.Equal(t, uint32(2), v2.GetVersion().GetMinor())
	require.Equal(t, sha[:], v2.GetContainerID().GetValue())
	require.Len(t, v2.GetRecords(), 1)

	newTable := eacl.NewTableFromV2(v2)
	require.Equal(t, table, newTable)

	t.Run("new from nil v2 table", func(t *testing.T) {
		require.Equal(t, new(eacl.Table), eacl.NewTableFromV2(nil))
	})

	t.Run("create table", func(t *testing.T) {
		id := cidtest.GenerateID()

		table := eacl.CreateTable(*id)
		require.Equal(t, id, table.CID())
		require.Equal(t, *version.Current(), table.Version())
	})
}

func TestTable_AddRecord(t *testing.T) {
	records := []*eacl.Record{
		eacl.CreateRecord(eacl.ActionDeny, eacl.OperationDelete),
		eacl.CreateRecord(eacl.ActionAllow, eacl.OperationPut),
	}

	table := eacl.NewTable()
	for _, record := range records {
		table.AddRecord(record)
	}

	require.Equal(t, records, table.Records())
}

func TestTableEncoding(t *testing.T) {
	tab := eacltest.Table()

	t.Run("binary", func(t *testing.T) {
		data, err := tab.Marshal()
		require.NoError(t, err)

		tab2 := eacl.NewTable()
		require.NoError(t, tab2.Unmarshal(data))

		// FIXME: we compare v2 messages because
		//  Filter contains fmt.Stringer interface
		require.Equal(t, tab.ToV2(), tab2.ToV2())
	})

	t.Run("json", func(t *testing.T) {
		data, err := tab.MarshalJSON()
		require.NoError(t, err)

		tab2 := eacl.NewTable()
		require.NoError(t, tab2.UnmarshalJSON(data))

		require.Equal(t, tab.ToV2(), tab2.ToV2())
	})
}

func TestTable_SessionToken(t *testing.T) {
	tok := sessiontest.Generate()

	table := eacl.NewTable()
	table.SetSessionToken(tok)

	require.Equal(t, tok, table.SessionToken())
}

func TestTable_Signature(t *testing.T) {
	sig := signature.New()
	sig.SetKey([]byte{1, 2, 3})
	sig.SetSign([]byte{4, 5, 6})

	table := eacl.NewTable()
	table.SetSignature(sig)

	require.Equal(t, sig, table.Signature())
}

func TestTable_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *eacl.Table

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		table := eacl.NewTable()

		// check initial values
		require.Equal(t, *version.Current(), table.Version())
		require.Nil(t, table.Records())
		require.Nil(t, table.CID())
		require.Nil(t, table.SessionToken())
		require.Nil(t, table.Signature())

		// convert to v2 message
		tableV2 := table.ToV2()

		require.Equal(t, version.Current().ToV2(), tableV2.GetVersion())
		require.Nil(t, tableV2.GetRecords())
		require.Nil(t, tableV2.GetContainerID())
	})
}
