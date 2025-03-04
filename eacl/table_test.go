package eacl_test

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

var invalidProtoEACLTestcases = []struct {
	name    string
	err     string
	corrupt func(*protoacl.EACLTable)
}{
	{name: "container/nil value", err: "invalid container ID: invalid length 0",
		corrupt: func(m *protoacl.EACLTable) { m.ContainerId = new(refs.ContainerID) }},
	{name: "container/value/empty", err: "invalid container ID: invalid length 0", corrupt: func(m *protoacl.EACLTable) {
		m.ContainerId = &refs.ContainerID{Value: []byte{}}
	}},
	{name: "container/undersized value", err: "invalid container ID: invalid length 31", corrupt: func(m *protoacl.EACLTable) {
		m.ContainerId = &refs.ContainerID{Value: make([]byte, 31)}
	}},
	{name: "container/oversized value", err: "invalid container ID: invalid length 33", corrupt: func(m *protoacl.EACLTable) {
		m.ContainerId = &refs.ContainerID{Value: make([]byte, 33)}
	}},
	{name: "container/zero", err: "invalid container ID: zero container ID", corrupt: func(m *protoacl.EACLTable) {
		m.ContainerId = &refs.ContainerID{Value: make([]byte, 32)}
	}},
	{name: "record/zero", err: "invalid container ID: zero container ID", corrupt: func(m *protoacl.EACLTable) {
		m.ContainerId = &refs.ContainerID{Value: make([]byte, 32)}
	}},
	{name: "records/nil element", err: "nil record #1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1] = nil
	}},
	{name: "records/negative action", err: "invalid record #1: negative action -1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Action = -1
	}},
	{name: "records/negative op", err: "invalid record #1: negative op -1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Operation = -1
	}},
	{name: "records/filters/nil element", err: "invalid record #1: nil filter #1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Filters[1] = nil
	}},
	{name: "records/filters/negative header type", err: "invalid record #1: invalid filter #1: negative header type -1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Filters[1].HeaderType = -1
	}},
	{name: "records/filters/negative match type", err: "invalid record #1: invalid filter #1: negative match type -1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Filters[1].MatchType = -1
	}},
	{name: "records/targets/nil element", err: "invalid record #1: nil target #1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Targets[1] = nil
	}},
	{name: "records/targets/negative role", err: "invalid record #1: invalid subject descriptor #1: negative role -1", corrupt: func(m *protoacl.EACLTable) {
		m.Records[1].Targets[1].Role = -1
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

func TestTable_FromProtoMessage(t *testing.T) {
	m := &protoacl.EACLTable{
		Version: &refs.Version{
			Major: rand.Uint32(),
			Minor: rand.Uint32(),
		},
		ContainerId: &refs.ContainerID{Value: testutil.RandByteSlice(cid.Size)},
		Records:     []*protoacl.EACLRecord{{}, {}},
	}

	for _, r := range m.Records {
		r.Targets = make([]*protoacl.EACLRecord_Target, 2)
		for i := range r.Targets {
			r.Targets[i] = &protoacl.EACLRecord_Target{
				Role: protoacl.Role(rand.Int31()),
				Keys: anyValidBinPublicKeys,
			}
		}
		r.Filters = make([]*protoacl.EACLRecord_Filter, 2)
		for i := range r.Filters {
			r.Filters[i] = &protoacl.EACLRecord_Filter{
				HeaderType: protoacl.HeaderType(rand.Int31()),
				MatchType:  protoacl.MatchType(rand.Int31()),
				Key:        "key_" + strconv.Itoa(rand.Int()),
				Value:      "val_" + strconv.Itoa(rand.Int()),
			}
		}
		r.Action = protoacl.Action(rand.Int31())
		r.Operation = protoacl.Operation(rand.Int31())
	}

	var tbl eacl.Table
	require.NoError(t, tbl.FromProtoMessage(m))
	require.EqualValues(t, m.Version.GetMajor(), tbl.Version().Major())
	require.EqualValues(t, m.Version.GetMinor(), tbl.Version().Minor())
	require.EqualValues(t, m.ContainerId.Value, tbl.GetCID())
	assertProtoRecordsEqual(t, tbl.Records(), m.Records)

	m.ContainerId = nil
	require.NoError(t, tbl.FromProtoMessage(m))
	require.Zero(t, tbl.GetCID())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidProtoEACLTestcases {
			t.Run(tc.name, func(t *testing.T) {
				m := anyValidEACL.ProtoMessage()
				require.NotNil(t, m)
				tc.corrupt(m)
				require.EqualError(t, new(eacl.Table).FromProtoMessage(m), tc.err)
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
