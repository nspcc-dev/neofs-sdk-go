package storagegroup_test

import (
	"crypto/sha256"
	"encoding/json"
	"math/rand/v2"
	"strconv"
	"testing"

	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	protosg "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	storagegrouptest "github.com/nspcc-dev/neofs-sdk-go/storagegroup/test"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

const (
	anyValidSize = uint64(15436265993370839342)
	anyValidExp  = uint64(13491075253593710190)
)

var (
	anyValidSHA256Hash = [sha256.Size]byte{49, 95, 91, 219, 118, 208, 120, 196, 59, 138, 192, 6, 78, 74, 1, 100,
		97, 43, 31, 206, 119, 200, 105, 52, 91, 252, 148, 199, 88, 148, 237, 211}
	anyValidTillichZemorHash = [tz.Size]byte{0, 0, 1, 66, 73, 241, 7, 149, 192, 36, 14, 221, 202, 138, 110, 191,
		0, 0, 1, 201, 196, 220, 152, 176, 23, 253, 146, 173, 98, 151, 156, 140, 0, 0, 0, 141, 148, 205, 152, 164,
		87, 185, 131, 233, 55, 131, 141, 205, 0, 0, 0, 219, 200, 104, 158, 117, 199, 221, 137, 37, 173, 13, 247, 39}
)

var anyValidMembers = []oid.ID{
	{138, 163, 217, 148, 183, 203, 248, 137, 98, 245, 243, 80, 22, 7, 219, 189,
		157, 190, 201, 32, 41, 255, 198, 245, 248, 206, 65, 33, 101, 127, 122, 216},
	{166, 174, 16, 34, 157, 146, 167, 232, 106, 101, 234, 123, 46, 85, 109, 169,
		62, 223, 253, 39, 172, 237, 222, 223, 134, 93, 176, 237, 93, 21, 9, 39},
}

var validChecksums = []checksum.Checksum{
	checksum.New(3259832435, []byte("Hello, world!")),
	checksum.NewSHA256(anyValidSHA256Hash),
	checksum.NewTillichZemor(anyValidTillichZemorHash),
}

// set by init.
var validStorageGroups = make([]storagegroup.StorageGroup, len(validChecksums))

func init() {
	for i := range validChecksums {
		validStorageGroups[i] = storagegroup.New(anyValidSize, validChecksums[i], anyValidMembers)
		validStorageGroups[i].SetExpirationEpoch(anyValidExp)
	}
}

// corresponds to validStorageGroups.
var validBinStorageGroups = [][]byte{
	{8, 174, 210, 190, 202, 173, 196, 168, 156, 214, 1, 18, 26, 8, 243, 176, 180, 146, 252, 255,
		255, 255, 255, 1, 18, 13, 72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100, 33, 24,
		238, 188, 144, 132, 238, 174, 251, 156, 187, 1, 34, 34, 10, 32, 138, 163, 217, 148, 183,
		203, 248, 137, 98, 245, 243, 80, 22, 7, 219, 189, 157, 190, 201, 32, 41, 255, 198, 245,
		248, 206, 65, 33, 101, 127, 122, 216, 34, 34, 10, 32, 166, 174, 16, 34, 157, 146, 167, 232,
		106, 101, 234, 123, 46, 85, 109, 169, 62, 223, 253, 39, 172, 237, 222, 223, 134, 93,
		176, 237, 93, 21, 9, 39},
	{8, 174, 210, 190, 202, 173, 196, 168, 156, 214, 1, 18, 36, 8, 2, 18, 32, 49, 95, 91, 219, 118,
		208, 120, 196, 59, 138, 192, 6, 78, 74, 1, 100, 97, 43, 31, 206, 119, 200, 105, 52, 91, 252,
		148, 199, 88, 148, 237, 211, 24, 238, 188, 144, 132, 238, 174, 251, 156, 187, 1, 34, 34, 10,
		32, 138, 163, 217, 148, 183, 203, 248, 137, 98, 245, 243, 80, 22, 7, 219, 189, 157, 190, 201,
		32, 41, 255, 198, 245, 248, 206, 65, 33, 101, 127, 122, 216, 34, 34, 10, 32, 166, 174, 16, 34,
		157, 146, 167, 232, 106, 101, 234, 123, 46, 85, 109, 169, 62, 223, 253, 39, 172, 237, 222, 223,
		134, 93, 176, 237, 93, 21, 9, 39},
	{8, 174, 210, 190, 202, 173, 196, 168, 156, 214, 1, 18, 68, 8, 1, 18, 64, 0, 0, 1, 66, 73, 241, 7,
		149, 192, 36, 14, 221, 202, 138, 110, 191, 0, 0, 1, 201, 196, 220, 152, 176, 23, 253, 146,
		173, 98, 151, 156, 140, 0, 0, 0, 141, 148, 205, 152, 164, 87, 185, 131, 233, 55, 131, 141, 205,
		0, 0, 0, 219, 200, 104, 158, 117, 199, 221, 137, 37, 173, 13, 247, 39, 24, 238, 188, 144, 132,
		238, 174, 251, 156, 187, 1, 34, 34, 10, 32, 138, 163, 217, 148, 183, 203, 248, 137, 98, 245,
		243, 80, 22, 7, 219, 189, 157, 190, 201, 32, 41, 255, 198, 245, 248, 206, 65, 33, 101, 127,
		122, 216, 34, 34, 10, 32, 166, 174, 16, 34, 157, 146, 167, 232, 106, 101, 234, 123, 46, 85, 109,
		169, 62, 223, 253, 39, 172, 237, 222, 223, 134, 93, 176, 237, 93, 21, 9, 39},
}

var validJSONStorageGroups = []string{`
{
 "validationDataSize": "15436265993370839342",
 "validationHash": {
  "type": -1035134861,
  "sum": "SGVsbG8sIHdvcmxkIQ=="
 },
 "expirationEpoch": "13491075253593710190",
 "members": [
  {
   "value": "iqPZlLfL+Ili9fNQFgfbvZ2+ySAp/8b1+M5BIWV/etg="
  },
  {
   "value": "pq4QIp2Sp+hqZep7LlVtqT7f/Ses7d7fhl2w7V0VCSc="
  }
 ]
}
`, `
{
 "validationDataSize": "15436265993370839342",
 "validationHash": {
  "type": "SHA256",
  "sum": "MV9b23bQeMQ7isAGTkoBZGErH853yGk0W/yUx1iU7dM="
 },
 "expirationEpoch": "13491075253593710190",
 "members": [
  {
   "value": "iqPZlLfL+Ili9fNQFgfbvZ2+ySAp/8b1+M5BIWV/etg="
  },
  {
   "value": "pq4QIp2Sp+hqZep7LlVtqT7f/Ses7d7fhl2w7V0VCSc="
  }
 ]
}
`, `
{
 "validationDataSize": "15436265993370839342",
 "validationHash": {
  "type": "TZ",
  "sum": "AAABQknxB5XAJA7dyopuvwAAAcnE3JiwF/2SrWKXnIwAAACNlM2YpFe5g+k3g43NAAAA28honnXH3YklrQ33Jw=="
 },
 "expirationEpoch": "13491075253593710190",
 "members": [
  {
   "value": "iqPZlLfL+Ili9fNQFgfbvZ2+ySAp/8b1+M5BIWV/etg="
  },
  {
   "value": "pq4QIp2Sp+hqZep7LlVtqT7f/Ses7d7fhl2w7V0VCSc="
  }
 ]
}
`,
}

type invalidProtoTestCase struct {
	name    string
	err     string
	corrupt func(*protosg.StorageGroup)
}

var invalidProtoTestcases = []invalidProtoTestCase{
	{name: "checksum/value/nil", err: "invalid hash: missing value", corrupt: func(sg *protosg.StorageGroup) {
		sg.SetValidationHash(new(refs.Checksum))
	}},
	{name: "checksum/value/empty", err: "invalid hash: missing value", corrupt: func(sg *protosg.StorageGroup) {
		var cs refs.Checksum
		cs.SetSum([]byte{})
		sg.SetValidationHash(&cs)
	}},
	{name: "members/value/nil", err: "invalid member #1: invalid length 0", corrupt: func(sg *protosg.StorageGroup) {
		members := make([]refs.ObjectID, 2)
		members[0].SetValue(anyValidMembers[0][:])
		sg.SetMembers(members)
	}},
	{name: "members/value/empty", err: "invalid member #1: invalid length 0", corrupt: func(sg *protosg.StorageGroup) {
		members := make([]refs.ObjectID, 2)
		members[0].SetValue(anyValidMembers[0][:])
		members[1].SetValue([]byte{})
		sg.SetMembers(members)
	}},
	{name: "members/value/undersize", err: "invalid member #1: invalid length 31", corrupt: func(sg *protosg.StorageGroup) {
		members := make([]refs.ObjectID, 2)
		members[0].SetValue(anyValidMembers[0][:])
		members[1].SetValue(anyValidMembers[1][:31])
		sg.SetMembers(members)
	}},
	{name: "members/value/oversize", err: "invalid member #1: invalid length 33", corrupt: func(sg *protosg.StorageGroup) {
		members := make([]refs.ObjectID, 2)
		members[0].SetValue(anyValidMembers[0][:])
		members[1].SetValue(append(anyValidMembers[1][:], 1))
		sg.SetMembers(members)
	}},
	{name: "members/duplicated", err: "duplicated member ALCAybSe17EF2b2e2TkfVVrMeQ6Gt6TW58rWkzzcGBoV", corrupt: func(sg *protosg.StorageGroup) {
		members := make([]refs.ObjectID, 3)
		members[0].SetValue(anyValidMembers[0][:])
		members[1].SetValue(anyValidMembers[1][:])
		members[2].SetValue(anyValidMembers[0][:])
		sg.SetMembers(members)
	}},
}

func TestStorageGroup_ReadFromV2(t *testing.T) {
	members := make([]refs.ObjectID, 2)
	members[0].SetValue(anyValidMembers[0][:])
	members[1].SetValue(anyValidMembers[1][:])
	var mcs refs.Checksum
	var m protosg.StorageGroup
	m.SetValidationDataSize(anyValidSize)
	//nolint:staticcheck
	m.SetExpirationEpoch(anyValidExp)
	m.SetMembers(members)
	var sg storagegroup.StorageGroup
	for i, tc := range []struct {
		typ refs.ChecksumType
		val []byte
	}{
		{43503860, []byte("Hello, world!")},
		{refs.SHA256, anyValidSHA256Hash[:]},
		{refs.TillichZemor, anyValidTillichZemorHash[:]},
	} {
		mcs.SetType(tc.typ)
		mcs.SetSum(tc.val)
		m.SetValidationHash(&mcs)
		require.NoError(t, sg.ReadFromV2(m), i)
		require.EqualValues(t, anyValidSize, sg.ValidationDataSize(), i)
		require.EqualValues(t, anyValidExp, sg.ExpirationEpoch(), i)
		require.Equal(t, anyValidMembers, sg.Members(), i)
		cs, ok := sg.ValidationDataHash()
		require.True(t, ok, i)
		require.Equal(t, tc.val, cs.Value(), i)
		switch typ := cs.Type(); tc.typ {
		default:
			require.EqualValues(t, tc.typ, typ, i)
		case refs.SHA256:
			require.Equal(t, checksum.SHA256, typ, i)
		case refs.TillichZemor:
			require.Equal(t, checksum.TillichZemor, typ, i)
		}
	}

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range append(invalidProtoTestcases, invalidProtoTestCase{
			name:    "missing checksum",
			err:     "missing hash",
			corrupt: func(sg *protosg.StorageGroup) { sg.SetValidationHash(nil) },
		}, invalidProtoTestCase{
			name:    "members/nil",
			err:     "missing members",
			corrupt: func(sg *protosg.StorageGroup) { sg.SetMembers(nil) },
		}, invalidProtoTestCase{
			name:    "members/empty",
			err:     "missing members",
			corrupt: func(sg *protosg.StorageGroup) { sg.SetMembers([]refs.ObjectID{}) },
		}) {
			t.Run(tc.name, func(t *testing.T) {
				var m protosg.StorageGroup
				storagegrouptest.StorageGroup().WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(storagegroup.StorageGroup).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestStorageGroup_WriteToV2(t *testing.T) {
	for i, sg := range validStorageGroups {
		var m protosg.StorageGroup
		sg.WriteToV2(&m)
		require.EqualValues(t, anyValidSize, m.GetValidationDataSize(), i)
		//nolint:staticcheck
		require.EqualValues(t, anyValidExp, m.GetExpirationEpoch(), i)
		members := m.GetMembers()
		require.Len(t, members, 2, i)
		require.Equal(t, anyValidMembers[0][:], members[0].GetValue(), i)
		require.EqualValues(t, anyValidMembers[1][:], members[1].GetValue(), i)
		mcs := m.GetValidationHash()
		require.Equal(t, validChecksums[i].Value(), mcs.GetSum(), i)
		switch typ := validChecksums[i].Type(); typ {
		default:
			require.EqualValues(t, typ, mcs.GetType())
		case checksum.SHA256:
			require.Equal(t, refs.SHA256, mcs.GetType())
		case checksum.TillichZemor:
			require.Equal(t, refs.TillichZemor, mcs.GetType())
		}
	}
}

func TestNew(t *testing.T) {
	size := rand.Uint64()
	cs := checksumtest.Checksum()
	members := oidtest.IDs(1 + rand.Int()%10)

	sg := storagegroup.New(size, cs, members)
	require.Equal(t, size, sg.ValidationDataSize())
	require.Equal(t, members, sg.Members())
	cs2, ok := sg.ValidationDataHash()
	require.True(t, ok)
	require.Equal(t, cs, cs2)
}

func TestStorageGroup_SetValidationDataSize(t *testing.T) {
	var sg storagegroup.StorageGroup
	require.Zero(t, sg.ValidationDataSize())

	sg.SetValidationDataSize(anyValidSize)
	require.EqualValues(t, anyValidSize, sg.ValidationDataSize())

	otherSz := anyValidSize + 1
	sg.SetValidationDataSize(otherSz)
	require.Equal(t, otherSz, sg.ValidationDataSize())
}

func TestStorageGroup_SetValidationDataHash(t *testing.T) {
	var sg storagegroup.StorageGroup
	_, ok := sg.ValidationDataHash()
	require.False(t, ok)

	for i, cs := range validChecksums {
		sg.SetValidationDataHash(cs)
		res, ok := sg.ValidationDataHash()
		require.True(t, ok, i)
		require.Equal(t, cs, res, i)
	}
}

func TestStorageGroup_SetExpirationEpoch(t *testing.T) {
	var sg storagegroup.StorageGroup
	require.Zero(t, sg.ExpirationEpoch())

	sg.SetExpirationEpoch(anyValidExp)
	require.EqualValues(t, anyValidExp, sg.ExpirationEpoch())

	otherExp := anyValidExp + 1
	sg.SetExpirationEpoch(otherExp)
	require.Equal(t, otherExp, sg.ExpirationEpoch())
}

func TestStorageGroup_SetMembers(t *testing.T) {
	var sg storagegroup.StorageGroup
	require.Zero(t, sg.Members())

	members := oidtest.IDs(3)
	sg.SetMembers(members)
	require.Equal(t, members, sg.Members())

	otherMembers := oidtest.IDs(5)
	sg.SetMembers(otherMembers)
	require.Equal(t, otherMembers, sg.Members())

	t.Run("double setting", func(t *testing.T) {
		var sg storagegroup.StorageGroup
		mm := oidtest.IDs(3) // cap is 3 at least
		require.NotPanics(t, func() { sg.SetMembers(mm) })
		// the previous cap is more that a new length; slicing should not lead to `out
		// of range` and apply update correctly
		require.NotPanics(t, func() { sg.SetMembers(mm[:1]) })
	})
}

func TestStorageGroup_Marshal(t *testing.T) {
	for i := range validStorageGroups {
		require.Equal(t, validBinStorageGroups[i], validStorageGroups[i].Marshal())
	}
}

func TestStorageGroup_Unmarshal(t *testing.T) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := new(storagegroup.StorageGroup).Unmarshal([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var sg storagegroup.StorageGroup
	for i := range validBinStorageGroups {
		err := sg.Unmarshal(validBinStorageGroups[i])
		require.NoError(t, err)
		require.Equal(t, validStorageGroups[i], sg)
		sg, err = storagegroup.Unmarshal(validBinStorageGroups[i])
		require.NoError(t, err)
		require.Equal(t, validStorageGroups[i], sg)
	}
}

func TestStorageGroup_MarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(storagegroup.StorageGroup).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var sg storagegroup.StorageGroup
	for i := range validStorageGroups {
		b, err := json.MarshalIndent(validStorageGroups[i], "", " ")
		require.NoError(t, err, i)
		require.NoError(t, sg.UnmarshalJSON(b), i)
		require.Equal(t, validStorageGroups[i], sg, i)
		sg, err = storagegroup.UnmarshalJSON(b)
		require.NoError(t, err, i)
		require.Equal(t, validStorageGroups[i], sg, i)
	}
}

func TestStorageGroup_UnmarshalJSON(t *testing.T) {
	var sg storagegroup.StorageGroup
	for i, j := range validJSONStorageGroups {
		require.NoError(t, sg.UnmarshalJSON([]byte(j)), i)
		require.Equal(t, validStorageGroups[i], sg, i)
		sg, err := storagegroup.UnmarshalJSON([]byte(j))
		require.NoError(t, err, i)
		require.Equal(t, validStorageGroups[i], sg, i)
	}
}

func newStorageGroupObject(exp uint64, sg storagegroup.StorageGroup) object.Object {
	var expAttr object.Attribute
	expAttr.SetKey("__NEOFS__EXPIRATION_EPOCH")
	expAttr.SetValue(strconv.FormatUint(exp, 10))

	var o object.Object
	o.SetAttributes(objecttest.Attribute(), expAttr, objecttest.Attribute())
	o.SetType(object.TypeStorageGroup)
	o.SetPayload(sg.Marshal())
	return o
}

func TestReadFromObject(t *testing.T) {
	sg := storagegroup.New(rand.Uint64(), checksumtest.Checksum(), oidtest.IDs(3))

	o := newStorageGroupObject(rand.Uint64(), sg)
	var sg2 storagegroup.StorageGroup
	require.NoError(t, storagegroup.ReadFromObject(&sg2, o))
	require.Equal(t, sg, sg2)

	t.Run("invalid object payload", func(t *testing.T) {
		o := newStorageGroupObject(rand.Uint64(), sg)
		o.SetPayload([]byte("Hello, world!"))

		err := storagegroup.ReadFromObject(new(storagegroup.StorageGroup), o)
		require.ErrorContains(t, err, "could not unmarshal storage group from object payload")
	})

	t.Run("invalid expiration attribute", func(t *testing.T) {
		o := newStorageGroupObject(rand.Uint64(), sg)
		var a object.Attribute
		a.SetKey("__NEOFS__EXPIRATION_EPOCH")
		a.SetValue("Hello, world!")
		o.SetAttributes(a)

		err := storagegroup.ReadFromObject(new(storagegroup.StorageGroup), o)
		require.ErrorContains(t, err, "could not get expiration from object")
	})

	t.Run("diff expiration in object and SG", func(t *testing.T) {
		sg := sg
		sg.SetExpirationEpoch(13)
		err := storagegroup.ReadFromObject(new(storagegroup.StorageGroup), newStorageGroupObject(42, sg))
		require.EqualError(t, err, "expiration does not match: from object: 42, from payload: 13")
	})

	t.Run("incorrect object type", func(t *testing.T) {
		var sg2 storagegroup.StorageGroup

		o.SetType(object.TypeTombstone)
		require.EqualError(t, storagegroup.ReadFromObject(&sg2, o), "object is not of StorageGroup type: TOMBSTONE")
	})
}

func TestWriteToObject(t *testing.T) {
	sg := storagegrouptest.StorageGroup()

	sgRaw := sg.Marshal()

	t.Run("empty object", func(t *testing.T) {
		var o object.Object
		storagegroup.WriteToObject(sg, &o)

		exp, found := expFromObj(t, o)
		require.True(t, found)

		require.Equal(t, sgRaw, o.Payload())
		require.Equal(t, sg.ExpirationEpoch(), exp)
		require.Equal(t, object.TypeStorageGroup, o.Type())
	})

	t.Run("obj already has exp attr", func(t *testing.T) {
		var o object.Object

		var attr object.Attribute
		attr.SetKey(protoobject.SysAttributeExpEpoch)
		attr.SetValue(strconv.FormatUint(sg.ExpirationEpoch()+1, 10))

		o.SetAttributes(object.Attribute{}, attr, object.Attribute{})

		storagegroup.WriteToObject(sg, &o)

		exp, found := expFromObj(t, o)
		require.True(t, found)

		require.Equal(t, sgRaw, o.Payload())
		require.Equal(t, sg.ExpirationEpoch(), exp)
		require.Equal(t, object.TypeStorageGroup, o.Type())
	})
}

func expFromObj(t *testing.T, o object.Object) (uint64, bool) {
	for _, attr := range o.Attributes() {
		if attr.Key() == protoobject.SysAttributeExpEpoch {
			exp, err := strconv.ParseUint(attr.Value(), 10, 64)
			require.NoError(t, err)

			return exp, true
		}
	}

	return 0, false
}
