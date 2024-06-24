package eacl_test

import (
	"testing"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestTable_Version(t *testing.T) {
	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst eacl.Table
			var msg apiacl.EACLTable

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			err = proto.Unmarshal(dst.Marshal(), &msg)
			require.Equal(t, &refs.Version{Major: 2, Minor: 13}, msg.Version)

			msg.Version.Major, msg.Version.Minor = 3, 14

			b, err := proto.Marshal(&msg)
			require.NoError(t, err)
			err = src.Unmarshal(b)
			require.NoError(t, err)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			msg.Version = nil
			err = proto.Unmarshal(dst.Marshal(), &msg)
			require.Equal(t, &refs.Version{Major: 3, Minor: 14}, msg.Version)
		})
		t.Run("api", func(t *testing.T) {
			var src, dst eacl.Table
			var msg apiacl.EACLTable

			src.SetRecords(eacltest.NRecords(2)) // just to satisfy decoder

			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			err = proto.Unmarshal(dst.Marshal(), &msg)
			require.Equal(t, &refs.Version{Major: 2, Minor: 13}, msg.Version)

			msg.Version.Major, msg.Version.Minor = 3, 14

			b, err := proto.Marshal(&msg)
			require.NoError(t, err)
			err = src.Unmarshal(b)
			require.NoError(t, err)

			src.WriteToV2(&msg)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			msg.Version = nil
			err = proto.Unmarshal(dst.Marshal(), &msg)
			require.Equal(t, &refs.Version{Major: 3, Minor: 14}, msg.Version)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst eacl.Table
			var msg apiacl.EACLTable

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			j, err = dst.MarshalJSON()
			require.NoError(t, err)
			err = protojson.Unmarshal(j, &msg)
			require.EqualValues(t, 2, msg.Version.Major)
			require.EqualValues(t, 13, msg.Version.Minor)

			msg.Version.Major, msg.Version.Minor = 3, 14

			b, err := protojson.Marshal(&msg)
			require.NoError(t, err)
			err = src.UnmarshalJSON(b)
			require.NoError(t, err)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			msg.Version = nil
			j, err = dst.MarshalJSON()
			require.NoError(t, err)
			err = protojson.Unmarshal(j, &msg)
			require.EqualValues(t, 3, msg.Version.Major)
			require.EqualValues(t, 14, msg.Version.Minor)
		})
	})
}

func TestTable_LimitToContainer(t *testing.T) {
	var tbl eacl.Table

	require.Zero(t, tbl.LimitedContainer())

	cnr := cidtest.ID()
	cnrOther := cidtest.ChangeID(cnr)

	tbl.LimitToContainer(cnr)
	require.Equal(t, cnr, tbl.LimitedContainer())

	tbl.LimitToContainer(cnrOther)
	require.Equal(t, cnrOther, tbl.LimitedContainer())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst eacl.Table

			dst.LimitToContainer(cnr)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.LimitedContainer())

			dst.LimitToContainer(cnrOther)
			src.LimitToContainer(cnr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, cnr, dst.LimitedContainer())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst eacl.Table
			var msg apiacl.EACLTable

			src.SetRecords(eacltest.NRecords(2)) // just to satisfy decoder

			dst.LimitToContainer(cnr)

			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Zero(t, dst.LimitedContainer())

			dst.LimitToContainer(cnrOther)
			src.LimitToContainer(cnr)
			src.WriteToV2(&msg)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, cnr, dst.LimitedContainer())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst eacl.Table

			dst.LimitToContainer(cnr)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.LimitedContainer())

			dst.LimitToContainer(cnrOther)
			src.LimitToContainer(cnr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, cnr, dst.LimitedContainer())
		})
	})
}

func TestTable_Records(t *testing.T) {
	var tbl eacl.Table

	require.Zero(t, tbl.Records())

	rs := eacltest.NRecords(3)
	tbl.SetRecords(rs)
	require.Equal(t, rs, tbl.Records())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst eacl.Table

			dst.SetRecords(eacltest.NRecords(2))

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Nil(t, dst.Records())

			dst.SetRecords(eacltest.NRecords(3))
			rs := eacltest.NRecords(3)
			src.SetRecords(rs)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, rs, dst.Records())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst eacl.Table
			var msg apiacl.EACLTable

			dst.SetRecords(eacltest.NRecords(3))
			rs := eacltest.NRecords(3)
			src.SetRecords(rs)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, rs, dst.Records())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst eacl.Table

			dst.SetRecords(eacltest.NRecords(2))

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Nil(t, dst.Records())

			dst.SetRecords(eacltest.NRecords(3))
			rs := eacltest.NRecords(3)
			src.SetRecords(rs)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, rs, dst.Records())
		})
	})
}

func TestTable_CopyTo(t *testing.T) {
	src := eacltest.Table()

	dst := eacltest.Table()
	src.CopyTo(&dst)
	require.Equal(t, src, dst)

	originAction := src.Records()[0].Action()
	otherAction := originAction + 1
	src.Records()[0].SetAction(otherAction)
	require.Equal(t, otherAction, src.Records()[0].Action())
	require.Equal(t, originAction, dst.Records()[0].Action())
}

func TestTable_SignedData(t *testing.T) {
	tbl := eacltest.Table()
	require.Equal(t, tbl.Marshal(), tbl.SignedData())
}

func TestTable_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("records", func(t *testing.T) {
			tbl := eacltest.Table()
			tbl.SetRecords(nil)
			var m apiacl.EACLTable
			tbl.WriteToV2(&m)
			require.ErrorContains(t, tbl.ReadFromV2(&m), "missing records")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			tbl := eacltest.Table()
			var m apiacl.EACLTable
			tbl.WriteToV2(&m)

			m.ContainerId.Value = []byte("not_a_container_ID")
			require.ErrorContains(t, tbl.ReadFromV2(&m), "invalid container")
		})
		t.Run("records", func(t *testing.T) {
			t.Run("targets", func(t *testing.T) {
				rs := eacltest.NRecords(2)
				rs[1].SetTargets(eacltest.NTargets(3))
				tbl := eacltest.Table()
				tbl.SetRecords(rs)
				var m apiacl.EACLTable
				tbl.WriteToV2(&m)

				m.Records[1].Targets[2].Role, m.Records[1].Targets[2].Keys = 0, nil
				require.ErrorContains(t, tbl.ReadFromV2(&m), "invalid record #1: invalid target #2: role and public keys are not mutually exclusive")
				m.Records[1].Targets[2].Role, m.Records[1].Targets[2].Keys = 1, make([][]byte, 1)
				require.ErrorContains(t, tbl.ReadFromV2(&m), "invalid record #1: invalid target #2: role and public keys are not mutually exclusive")
				m.Records[1].Targets = nil
				require.ErrorContains(t, tbl.ReadFromV2(&m), "invalid record #1: missing target subjects")
			})
			t.Run("filters", func(t *testing.T) {
				rs := eacltest.NRecords(2)
				rs[1].SetFilters(eacltest.NFilters(3))
				tbl := eacltest.Table()
				tbl.SetRecords(rs)
				var m apiacl.EACLTable
				tbl.WriteToV2(&m)

				m.Records[1].Filters[2].Key = ""
				require.ErrorContains(t, tbl.ReadFromV2(&m), "invalid record #1: invalid filter #2: missing key")
			})
		})
	})
}

func TestTable_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var tbl eacl.Table
		msg := []byte("definitely_not_protobuf")
		err := tbl.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			tbl := eacltest.Table()
			var m apiacl.EACLTable
			tbl.WriteToV2(&m)
			m.ContainerId.Value = []byte("not_a_container_ID")
			b, err := proto.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tbl.Unmarshal(b), "invalid container")
		})
	})
}

func TestTable_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var tbl eacl.Table
		msg := []byte("definitely_not_protojson")
		err := tbl.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			tbl := eacltest.Table()
			var m apiacl.EACLTable
			tbl.WriteToV2(&m)
			m.ContainerId.Value = []byte("not_a_container_ID")
			b, err := protojson.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tbl.UnmarshalJSON(b), "invalid container")
		})
	})
}
