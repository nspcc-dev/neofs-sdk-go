package oid_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestAddressComparable(t *testing.T) {
	a1 := oidtest.Address()
	require.True(t, a1 == a1)
	a2 := oidtest.ChangeAddress(a1)
	require.NotEqual(t, a1, a2)
	require.False(t, a1 == a2)
}

func TestAddress_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			a := oidtest.Address()
			var m refs.Address

			a.WriteToV2(&m)
			m.ContainerId = nil
			require.ErrorContains(t, a.ReadFromV2(&m), "missing container ID")
		})
		t.Run("object", func(t *testing.T) {
			a := oidtest.Address()
			var m refs.Address

			a.WriteToV2(&m)
			m.ObjectId = nil
			require.ErrorContains(t, a.ReadFromV2(&m), "missing object ID")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			a := oidtest.Address()
			var m refs.Address

			a.WriteToV2(&m)
			m.ContainerId.Value = make([]byte, 31)
			require.ErrorContains(t, a.ReadFromV2(&m), "invalid container ID: invalid value length 31")
			m.ContainerId.Value = make([]byte, 33)
			require.ErrorContains(t, a.ReadFromV2(&m), "invalid container ID: invalid value length 33")
		})
		t.Run("object", func(t *testing.T) {
			a := oidtest.Address()
			var m refs.Address

			a.WriteToV2(&m)
			m.ObjectId.Value = make([]byte, 31)
			require.ErrorContains(t, a.ReadFromV2(&m), "invalid object ID: invalid value length 31")
			m.ObjectId.Value = make([]byte, 33)
			require.ErrorContains(t, a.ReadFromV2(&m), "invalid object ID: invalid value length 33")
		})
	})
}

func TestNodeInfo_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var a oid.Address
		msg := []byte("definitely_not_protojson")
		err := a.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
	t.Run("invalid fields", func(t *testing.T) {
		testCases := []struct {
			name string
			err  string
			json string
		}{{name: "missing container", err: "missing container ID", json: `
{
 "objectID": {
  "value": "a86CDbsGIuktRdOUdsUkdW4iaNsfjX5LwUncE1zHk+s="
 }
}`},
			{name: "missing object", err: "missing object ID", json: `
{
 "containerID": {
  "value": "lW85+mf1fm2JnD2P/sVrSijaf2U2G6v+PEYC6EeGk9s="
 }
}
`},
			{name: "invalid container length", err: "invalid container ID: invalid value length 31", json: `
{
 "containerID": {
  "value": "LfmuVsOC2bNfuxvuGrdnHwIM+QhDMO8eD22Vlgl2JQ=="
 },
 "objectID": {
  "value": "a86CDbsGIuktRdOUdsUkdW4iaNsfjX5LwUncE1zHk+s="
 }
}`},
			{name: "invalid object length", err: "invalid object ID: invalid value length 33", json: `
{
 "containerID": {
  "value": "lW85+mf1fm2JnD2P/sVrSijaf2U2G6v+PEYC6EeGk9s="
 },
 "objectID": {
  "value": "3007wQX0PGK+/ERYq1Xj/Lg6qMj2jsnDorgzB/apoi6v"
 }
}`},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				var a oid.Address
				require.ErrorContains(t, a.UnmarshalJSON([]byte(testCase.json)), testCase.err)
			})
		}
	})
}

func TestAddress_DecodeString(t *testing.T) {
	var a oid.Address

	const zeroAddrString = "11111111111111111111111111111111/11111111111111111111111111111111"
	require.Equal(t, zeroAddrString, a.EncodeToString())
	a = oidtest.ChangeAddress(a)
	require.NoError(t, a.DecodeString(zeroAddrString))
	require.Equal(t, zeroAddrString, a.EncodeToString())
	require.Zero(t, a)

	var bCnr = [32]byte{231, 129, 236, 104, 74, 71, 155, 100, 72, 209, 186, 80, 2, 184, 9, 161, 10, 76, 18, 203, 126, 94, 101, 42, 157, 211, 66, 99, 247, 143, 226, 23}
	var bObj = [32]byte{67, 239, 220, 249, 222, 147, 14, 92, 52, 46, 242, 209, 101, 80, 248, 39, 206, 189, 29, 55, 8, 3, 70, 205, 213, 7, 46, 54, 192, 232, 35, 247}
	const str = "Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN/5aCTDRcH248TManyLPJYkGnqsCGAfiyMw6rgvNCkHu98"
	require.NoError(t, a.DecodeString(str))
	require.Equal(t, str, a.EncodeToString())
	require.EqualValues(t, bCnr, a.Container())
	require.EqualValues(t, bObj, a.Object())

	var bCnrOther = [32]byte{190, 131, 185, 144, 207, 179, 2, 201, 93, 205, 169, 242, 167, 89, 56, 112, 48, 5, 13, 128, 58, 179, 92, 119, 37, 234, 236, 35, 9, 89, 73, 97}
	var bObjOther = [32]byte{77, 244, 70, 159, 204, 190, 29, 22, 105, 203, 94, 30, 169, 236, 97, 176, 179, 51, 89, 138, 164, 69, 157, 131, 190, 246, 16, 93, 93, 249, 66, 95}
	const strOther = "DpgxiqnrkpZzYwmT58AzY9w51V41P5dWNKuVGm7oeEak/6FJS2jh2cKmtHL54tQSREuJ3bUG2pkbvChJPyJ3ZchSW"
	require.NoError(t, a.DecodeString(strOther))
	require.Equal(t, strOther, a.EncodeToString())
	require.EqualValues(t, bCnrOther, a.Container())
	require.EqualValues(t, bObjOther, a.Object())

	t.Run("invalid", func(t *testing.T) {
		var a oid.Address
		for _, testCase := range []struct{ input, err string }{
			{input: "", err: "missing delimiter"},
			{input: "no_delimiter", err: "missing delimiter"},
			{input: "/Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN", err: "decode container string: invalid value length 0"},
			{input: "Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN/", err: "decode object string: invalid value length 0"},
			{input: "qxAE9SLuDq7dARPAFaWG6vbuGoocwoTn19LK5YVqnS/Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN", err: "decode container string: invalid value length 31"},
			{input: "Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN/qxAE9SLuDq7dARPAFaWG6vbuGoocwoTn19LK5YVqnS", err: "decode object string: invalid value length 31"},
			{input: "HJJEkEKthnvMw7NsZNgzBEQ4tf9AffmaBYWxfBULvvbPW/Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN", err: "decode container string: invalid value length 33"},
			{input: "Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN/HJJEkEKthnvMw7NsZNgzBEQ4tf9AffmaBYWxfBULvvbPW", err: "decode object string: invalid value length 33"},
		} {
			require.ErrorContains(t, a.DecodeString(testCase.input), testCase.err, testCase)
		}
	})
	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst oid.Address
			var msg refs.Address

			require.NoError(t, dst.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 32), msg.ContainerId.Value)
			require.Equal(t, make([]byte, 32), msg.ObjectId.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst)
			require.Equal(t, zeroAddrString, dst.EncodeToString())

			require.NoError(t, src.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, bCnr[:], msg.ContainerId.Value)
			require.EqualValues(t, bObj[:], msg.ObjectId.Value)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, bCnr, dst.Container())
			require.EqualValues(t, bObj, dst.Object())
			require.Equal(t, str, dst.EncodeToString())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst oid.Address

			require.NoError(t, dst.DecodeString(str))

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(j))
			require.Zero(t, dst)
			require.Equal(t, zeroAddrString, dst.EncodeToString())

			require.NoError(t, src.DecodeString(str))

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.EqualValues(t, bCnr, dst.Container())
			require.EqualValues(t, bObj, dst.Object())
			require.Equal(t, str, dst.EncodeToString())
		})
	})
}

func TestAddress_SetContainer(t *testing.T) {
	var a oid.Address

	require.Zero(t, a.Container())

	cnr := cidtest.ID()
	a.SetContainer(cnr)
	require.Equal(t, cnr, a.Container())

	cnrOther := cidtest.ChangeID(cnr)
	a.SetContainer(cnrOther)
	require.Equal(t, cnrOther, a.Container())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst oid.Address
			var msg refs.Address

			// set required data just to satisfy decoder
			src.SetObject(oidtest.ID())

			dst.SetContainer(cnr)

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 32), msg.ContainerId.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Container())

			src.SetContainer(cnr)

			src.WriteToV2(&msg)
			require.Equal(t, cnr[:], msg.ContainerId.Value)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, cnr, dst.Container())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst oid.Address

			// set required data just to satisfy decoder
			src.SetObject(oidtest.ID())

			dst.SetContainer(cnr)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(j))
			require.Zero(t, dst.Container())

			src.SetContainer(cnr)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, cnr, dst.Container())
		})
	})
}

func TestAddress_SetObject(t *testing.T) {
	var a oid.Address

	require.Zero(t, a.Object())

	obj := oidtest.ID()
	a.SetObject(obj)
	require.Equal(t, obj, a.Object())

	objOther := oidtest.ChangeID(obj)
	a.SetObject(objOther)
	require.Equal(t, objOther, a.Object())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst oid.Address
			var msg refs.Address

			// set required data just to satisfy decoder
			src.SetContainer(cidtest.ID())

			dst.SetObject(obj)

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 32), msg.ObjectId.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Object())

			src.SetObject(obj)

			src.WriteToV2(&msg)
			require.Equal(t, obj[:], msg.ObjectId.Value)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, obj, dst.Object())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst oid.Address

			// set required data just to satisfy decoder
			src.SetContainer(cidtest.ID())

			dst.SetObject(obj)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(j))
			require.Zero(t, dst.Object())

			src.SetObject(obj)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, obj, dst.Object())
		})
	})
}
