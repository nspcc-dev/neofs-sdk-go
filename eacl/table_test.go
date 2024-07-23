package eacl_test

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"

	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

var invalidProtoEACLTestcases = []struct {
	name    string
	err     string
	corrupt func(*protoacl.Table)
}{
	{name: "invalid container/nil value", err: "invalid container ID: invalid length 0",
		corrupt: func(m *protoacl.Table) { m.SetContainerID(new(refs.ContainerID)) }},
	{name: "invalid container/empty value", err: "invalid container ID: invalid length 0", corrupt: func(m *protoacl.Table) {
		var mc refs.ContainerID
		mc.SetValue([]byte{})
		m.SetContainerID(&mc)
	}},
	{name: "invalid container/undersized value", err: "invalid container ID: invalid length 31", corrupt: func(m *protoacl.Table) {
		var mc refs.ContainerID
		mc.SetValue(make([]byte, 31))
		m.SetContainerID(&mc)
	}},
	{name: "invalid container/oversized value", err: "invalid container ID: invalid length 33", corrupt: func(m *protoacl.Table) {
		var mc refs.ContainerID
		mc.SetValue(make([]byte, 33))
		m.SetContainerID(&mc)
	}},
}

func TestTable_AddRecord(t *testing.T) {
	for i := range anyValidRecords {
		var tbl eacl.Table
		tbl.AddRecord(&anyValidRecords[i])
		require.Len(t, tbl.Records(), 1)
		require.Equal(t, anyValidRecords[i], tbl.Records()[0])
	}
}

func TestTable_ToV2(t *testing.T) {
	assert := func(t testing.TB, tbl eacl.Table, m *protoacl.Table) {
		require.EqualValues(t, 2, tbl.Version().Major())
		require.EqualValues(t, 16, tbl.Version().Minor())
		require.Len(t, m.GetRecords(), len(tbl.Records()))
		assertProtoRecordsEqual(t, tbl.Records(), m.GetRecords())
	}

	tbl := eacl.ConstructTable(anyValidRecords)
	assert(t, tbl, tbl.ToV2())

	t.Run("with container", func(t *testing.T) {
		cnr := cidtest.ID()
		tbl = eacl.NewTableForContainer(cnr, anyValidRecords)
		m := tbl.ToV2()
		assert(t, tbl, m)
		require.Equal(t, cnr[:], m.GetContainerID().GetValue())
	})
	t.Run("default values", func(t *testing.T) {
		table := eacl.NewTable()

		// check initial values
		require.Equal(t, version.Current(), table.Version())
		require.Nil(t, table.Records())
		_, set := table.CID()
		require.False(t, set)

		// convert to v2 message
		tableV2 := table.ToV2()

		var verV2 refs.Version
		version.Current().WriteToV2(&verV2)
		require.Equal(t, verV2, *tableV2.GetVersion())
		require.Nil(t, tableV2.GetRecords())
		require.Nil(t, tableV2.GetContainerID())
	})
}

func TestTable_LimitToContainer(t *testing.T) {
	cnr := cidtest.ID()
	var tbl eacl.Table
	require.Zero(t, tbl.GetCID())
	tbl.SetCID(cnr)
	require.Equal(t, cnr, tbl.GetCID())
}

func TestTable_CID(t *testing.T) {
	cnr := cidtest.ID()
	var tbl eacl.Table
	_, ok := tbl.CID()
	require.False(t, ok)
	tbl.SetCID(cnr)
	res, ok := tbl.CID()
	require.True(t, ok)
	require.Equal(t, cnr, res)
}

func TestTable_SetCID(t *testing.T) {
	cnr := cidtest.ID()
	var tbl eacl.Table
	tbl.SetCID(cnr)
	require.Equal(t, cnr, tbl.GetCID())
}

func TestNewTableFromV2(t *testing.T) {
	var ver refs.Version
	ver.SetMajor(rand.Uint32())
	ver.SetMinor(rand.Uint32())

	bCnr := make([]byte, cid.Size)
	//nolint:staticcheck
	rand.Read(bCnr)
	var cnr refs.ContainerID
	cnr.SetValue(bCnr)

	rs := make([]protoacl.Record, 2)
	for i := range rs {
		ts := make([]protoacl.Target, 2)
		for i := range ts {
			ts[i].SetRole(protoacl.Role(rand.Uint32()))
			ts[i].SetKeys(anyValidBinPublicKeys)
		}
		fs := make([]protoacl.HeaderFilter, 2)
		for i := range fs {
			fs[i].SetHeaderType(protoacl.HeaderType(rand.Uint32()))
			fs[i].SetKey("key_" + strconv.Itoa(rand.Int()))
			fs[i].SetMatchType(protoacl.MatchType(rand.Uint32()))
			fs[i].SetValue("val_" + strconv.Itoa(rand.Int()))
		}
		rs[i].SetAction(protoacl.Action(rand.Uint32()))
		rs[i].SetOperation(protoacl.Operation(rand.Uint32()))
		rs[i].SetTargets(ts)
		rs[i].SetFilters(fs)
	}

	var m protoacl.Table
	m.SetVersion(&ver)
	m.SetContainerID(&cnr)
	m.SetRecords(rs)

	tbl := eacl.NewTableFromV2(&m)
	require.EqualValues(t, ver.GetMajor(), tbl.Version().Major())
	require.EqualValues(t, ver.GetMinor(), tbl.Version().Minor())
	resCnr, ok := tbl.CID()
	require.True(t, ok)
	require.Equal(t, bCnr, resCnr[:])
	assertProtoRecordsEqual(t, tbl.Records(), rs)

	t.Run("nil", func(t *testing.T) {
		require.Equal(t, new(eacl.Table), eacl.NewTableFromV2(nil))
	})
}

func TestTable_ReadFromV2(t *testing.T) {
	var ver refs.Version
	ver.SetMajor(rand.Uint32())
	ver.SetMinor(rand.Uint32())

	bCnr := make([]byte, cid.Size)
	//nolint:staticcheck
	rand.Read(bCnr)
	var cnr refs.ContainerID
	cnr.SetValue(bCnr)

	rs := make([]protoacl.Record, 2)
	for i := range rs {
		ts := make([]protoacl.Target, 2)
		for i := range ts {
			ts[i].SetRole(protoacl.Role(rand.Uint32()))
			ts[i].SetKeys(anyValidBinPublicKeys)
		}
		fs := make([]protoacl.HeaderFilter, 2)
		for i := range fs {
			fs[i].SetHeaderType(protoacl.HeaderType(rand.Uint32()))
			fs[i].SetKey("key_" + strconv.Itoa(rand.Int()))
			fs[i].SetMatchType(protoacl.MatchType(rand.Uint32()))
			fs[i].SetValue("val_" + strconv.Itoa(rand.Int()))
		}
		rs[i].SetAction(protoacl.Action(rand.Uint32()))
		rs[i].SetOperation(protoacl.Operation(rand.Uint32()))
		rs[i].SetTargets(ts)
		rs[i].SetFilters(fs)
	}

	var m protoacl.Table
	m.SetVersion(&ver)
	m.SetContainerID(&cnr)
	m.SetRecords(rs)

	var tbl eacl.Table
	require.NoError(t, tbl.ReadFromV2(m))
	require.EqualValues(t, ver.GetMajor(), tbl.Version().Major())
	require.EqualValues(t, ver.GetMinor(), tbl.Version().Minor())
	require.EqualValues(t, bCnr, tbl.GetCID())
	assertProtoRecordsEqual(t, tbl.Records(), rs)

	m.SetContainerID(nil)
	require.NoError(t, tbl.ReadFromV2(m))
	require.Zero(t, tbl.GetCID())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidProtoEACLTestcases {
			t.Run(tc.name, func(t *testing.T) {
				m := anyValidEACL.ToV2()
				require.NotNil(t, m)
				tc.corrupt(m)
				require.EqualError(t, new(eacl.Table).ReadFromV2(*m), tc.err)
			})
		}
	})
}

func TestTable_SignedData(t *testing.T) {
	require.Equal(t, anyValidEACLBytes, anyValidEACL.SignedData())
}

func TestTable_Marshal(t *testing.T) {
	require.Equal(t, anyValidEACLBytes, anyValidEACL.Marshal())
}

func testUnmarshalTableFunc(t *testing.T, f func(*eacl.Table, []byte) error) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := f(new(eacl.Table), []byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var tbl eacl.Table
	require.NoError(t, f(&tbl, anyValidEACLBytes))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, anyValidEACL, tbl)
}

func TestTable_Unmarshal(t *testing.T) {
	testUnmarshalTableFunc(t, (*eacl.Table).Unmarshal)
}

func TestUnmarshal(t *testing.T) {
	testUnmarshalTableFunc(t, func(tbl *eacl.Table, b []byte) error {
		res, err := eacl.Unmarshal(b)
		if err == nil {
			*tbl = res
		}
		return err
	})
}

func TestTable_MarshalJSON(t *testing.T) {
	var tbl1 eacl.Table
	b, err := anyValidEACL.MarshalJSON()
	require.NoError(t, err)
	require.NoError(t, tbl1.UnmarshalJSON(b))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, anyValidEACL, tbl1)

	var tbl2 eacl.Table
	b, err = json.Marshal(anyValidEACL)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &tbl2))
	require.Equal(t, anyValidEACL, tbl2)
}

func testUnmarshalTableJSONFunc(t *testing.T, f func(*eacl.Table, []byte) error) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := f(new(eacl.Table), []byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var tbl eacl.Table
	require.NoError(t, f(&tbl, []byte(anyValidEACLJSON)))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, anyValidEACL, tbl)
}

func TestTable_UnmarshalJSON(t *testing.T) {
	testUnmarshalTableJSONFunc(t, (*eacl.Table).UnmarshalJSON)

	var tbl eacl.Table
	require.NoError(t, json.Unmarshal([]byte(anyValidEACLJSON), &tbl))
	require.Equal(t, anyValidEACL, tbl)

	tbl3, err := eacl.UnmarshalJSON([]byte(anyValidEACLJSON))
	require.NoError(t, err)
	require.Equal(t, anyValidEACL, tbl3)
}

func TestUnmarshalJSON(t *testing.T) {
	testUnmarshalTableJSONFunc(t, func(tbl *eacl.Table, b []byte) error {
		res, err := eacl.UnmarshalJSON(b)
		if err == nil {
			*tbl = res
		}
		return err
	})

	tbl, err := eacl.UnmarshalJSON([]byte(anyValidEACLJSON))
	require.NoError(t, err)
	require.Equal(t, anyValidEACL, tbl)
}

func TestSetRecords(t *testing.T) {
	var tbl eacl.Table
	require.Zero(t, tbl.Records())
	tbl.SetRecords(anyValidRecords)
	require.Equal(t, anyValidRecords, tbl.Records())
}

func TestConstructTable(t *testing.T) {
	tbl := eacl.ConstructTable(anyValidRecords)
	require.Equal(t, anyValidRecords, tbl.Records())
	_, ok := tbl.CID()
	require.False(t, ok)
	require.EqualValues(t, 2, tbl.Version().Major())
	require.EqualValues(t, 16, tbl.Version().Minor())
}

func TestNewTableForContainer(t *testing.T) {
	cnr := cidtest.ID()
	tbl := eacl.NewTableForContainer(cnr, anyValidRecords)
	require.Equal(t, anyValidRecords, tbl.Records())
	cnr2, ok := tbl.CID()
	require.True(t, ok)
	require.Equal(t, cnr, cnr2)
	require.EqualValues(t, 2, tbl.Version().Major())
	require.EqualValues(t, 16, tbl.Version().Minor())
}

func TestCreateTable(t *testing.T) {
	tbl := eacl.CreateTable(anyValidContainerID)
	require.EqualValues(t, 2, tbl.Version().Major())
	require.EqualValues(t, 16, tbl.Version().Minor())
	cnr, ok := tbl.CID()
	require.True(t, ok)
	require.Equal(t, anyValidContainerID, cnr)
	require.Zero(t, tbl.Records())
}
