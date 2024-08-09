package eacl_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
	"testing"

	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

// Crypto.
var (
	anyECDSAPublicKeys = []ecdsa.PublicKey{
		{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218}),
			Y: new(big.Int).SetBytes([]byte{16, 170, 6, 224, 77, 3, 245, 72, 144, 58, 69, 14, 160, 35, 57, 108, 111, 27, 224, 129, 88, 230, 68, 48, 17, 10, 207, 118, 199, 120, 184, 119}),
		},
		{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136}),
			Y: new(big.Int).SetBytes([]byte{118, 201, 238, 8, 178, 41, 96, 3, 163, 197, 31, 58, 106, 218, 104, 47, 106, 153, 180, 68, 109, 243, 62, 31, 159, 17, 104, 134, 134, 97, 117, 52}),
		},
	}
	// // corresponds to anyECDSAPublicKeys.
	anyECDSAPublicKeysPtr []*ecdsa.PublicKey // set by init
	// corresponds to anyECDSAPublicKeys.
	anyValidECDSABinPublicKeys = [][]byte{
		{3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218},
		{2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136},
	}
)

// Other NeoFS stuff.
var (
	anyValidObjectID           = oid.ID{86, 149, 134, 57, 161, 211, 240, 124, 106, 146, 201, 140, 249, 50, 158, 38, 82, 140, 5, 160, 180, 117, 106, 214, 47, 255, 166, 89, 55, 99, 178, 66}
	anyValidObjectIDString     = "6pzJXAjxTH6i38Yk9dFWAPY6wrUpSLi4DUwZf82EompD"
	anyValidProtoVersion       = version.New(4835, 1532621)
	anyValidProtoVersionString = "v4835.1532621"
	anyValidContainerID        = cid.ID{243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227}
	anyValidContainerIDString  = "HRK1fwY2PMS4mgy292eZeDBr5XyH6mHbXzx7ds3n81vz"
	anyValidUserID             = user.ID{53, 52, 176, 7, 17, 140, 220, 179, 170, 128, 138, 214, 130, 87, 179, 211, 36, 197, 16, 38, 50, 88, 207, 120, 145}
	anyValidUserIDString       = "NQiZCLuU4EeP47mMFW6FWEu7Pfm4MDFVBn"
	anyData                    = []byte("Hello, world!")
	anyUserSet                 = []user.ID{
		{53, 121, 249, 124, 139, 174, 30, 193, 143, 226, 163, 208, 188, 194, 173, 123, 60, 84, 224, 229, 4, 14, 206, 19, 117},
		{53, 246, 34, 60, 106, 147, 200, 106, 111, 144, 9, 61, 86, 46, 111, 148, 91, 65, 206, 216, 139, 168, 188, 23, 102},
		{53, 164, 213, 123, 6, 115, 90, 134, 224, 150, 72, 192, 236, 220, 188, 131, 102, 5, 152, 164, 166, 222, 119, 72, 228},
	}
	anyValidChecksums = []checksum.Checksum{
		checksum.New(0, anyData),
		checksum.New(4893983, anyData),
		checksum.NewSHA256(sha256.Sum256(anyData)),
		checksum.NewTillichZemor(tz.Sum(anyData)),
	}
	// corresponds to anyValidChecksums.
	anyValidStringChecksums = []string{
		"CHECKSUM_TYPE_UNSPECIFIED:48656c6c6f2c20776f726c6421",
		"4893983:48656c6c6f2c20776f726c6421",
		"SHA256:315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3",
		"TZ:0000014249f10795c0240eddca8a6ebf000001c9c4dc98b017fd92ad62979c8c0000008d94cd98a457b983e937838dcd000000dbc8689e75c7dd8925ad0df727",
	}
	anyValidObjectTypes = []object.Type{
		32852,
		object.TypeRegular,
		object.TypeTombstone,
		object.TypeStorageGroup,
		object.TypeLock,
		object.TypeLink,
	}
	anyValidStringObjectTypes = []string{
		"32852",
		"REGULAR",
		"TOMBSTONE",
		"STORAGE_GROUP",
		"LOCK",
		"LINK",
	}
)

// EACL.
var (
	// not const to avoid separate block.
	anyValidMatcher       = eacl.Match(548643430)
	anyValidHeaderType    = eacl.FilterHeaderType(40968380)
	anyValidRole          = eacl.Role(690857412)
	anyValidAction        = eacl.Action(9875285)
	anyValidOp            = eacl.Operation(3462843)
	anyValidBinPublicKeys = [][]byte{
		[]byte("key_1940340825"),
		[]byte("key_879439723842"),
	}
	anyValidFilters = []eacl.Filter{
		eacl.ConstructFilter(eacl.FilterHeaderType(4509681), "key_54093643", eacl.Match(949385), "val_34811040"),
		eacl.ConstructFilter(eacl.FilterHeaderType(582984), "key_1298432", eacl.Match(7539428), "val_8243258"),
	}
	anyValidTargets = []eacl.Target{
		eacl.NewTargetByRole(anyValidRole),
		{}, // set by init
	}
	anyValidRecords = []eacl.Record{
		eacl.ConstructRecord(eacl.Action(5692342), eacl.Operation(12943052), []eacl.Target{anyValidTargets[0]}, anyValidFilters[0]),
		eacl.ConstructRecord(eacl.Action(43658603), eacl.Operation(12943052), anyValidTargets, anyValidFilters...),
	}
	anyValidEACL = eacl.NewTableForContainer(anyValidContainerID, anyValidRecords)
)

func init() {
	rawSubjs := [][]byte{
		anyValidECDSABinPublicKeys[0],
		anyUserSet[0][:],
		anyValidECDSABinPublicKeys[1],
		anyUserSet[1][:],
		anyUserSet[2][:],
	}
	anyValidTargets[1].SetRawSubjects(rawSubjs)
}

// Protobuf.
var (
	// corresponds to anyValidFilters.
	anyValidBinFilters = [][]byte{
		{8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57, 51, 54, 52, 51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48},
		{8, 200, 202, 35, 16, 228, 149, 204, 3, 26, 11, 107, 101, 121, 95, 49, 50, 57, 56, 52, 51, 50, 34, 11, 118, 97, 108, 95, 56, 50, 52, 51, 50, 53, 56},
	}
	// corresponds to anyValidTargets.
	anyValidBinTargets = [][]byte{
		{8, 196, 203, 182, 201, 2},
		{18, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219,
			209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18, 25, 53, 121, 249, 124, 139, 174, 30, 193, 143, 226, 163,
			208, 188, 194, 173, 123, 60, 84, 224, 229, 4, 14, 206, 19, 117, 18, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233,
			126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 18, 25, 53, 246, 34,
			60, 106, 147, 200, 106, 111, 144, 9, 61, 86, 46, 111, 148, 91, 65, 206, 216, 139, 168, 188, 23, 102, 18, 25, 53, 164,
			213, 123, 6, 115, 90, 134, 224, 150, 72, 192, 236, 220, 188, 131, 102, 5, 152, 164, 166, 222, 119, 72, 228},
	}
	// corresponds to anyValidRecords.
	anyValidBinRecords = [][]byte{
		{8, 204, 253, 149, 6, 16, 182, 183, 219, 2, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53,
			52, 48, 57, 51, 54, 52, 51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 34, 6, 8, 196, 203, 182, 201, 2},
		{8, 204, 253, 149, 6, 16, 235, 218, 232, 20, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53,
			52, 48, 57, 51, 54, 52, 51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 26, 35, 8, 200, 202, 35, 16,
			228, 149, 204, 3, 26, 11, 107, 101, 121, 95, 49, 50, 57, 56, 52, 51, 50, 34, 11, 118, 97, 108, 95, 56, 50, 52, 51, 50,
			53, 56, 34, 6, 8, 196, 203, 182, 201, 2, 34, 151, 1, 18, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21,
			173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18, 25, 53,
			121, 249, 124, 139, 174, 30, 193, 143, 226, 163, 208, 188, 194, 173, 123, 60, 84, 224, 229, 4, 14, 206, 19, 117, 18,
			33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95,
			85, 78, 45, 197, 136, 18, 25, 53, 246, 34, 60, 106, 147, 200, 106, 111, 144, 9, 61, 86, 46, 111, 148, 91, 65, 206, 216, 139,
			168, 188, 23, 102, 18, 25, 53, 164, 213, 123, 6, 115, 90, 134, 224, 150, 72, 192, 236, 220, 188, 131, 102, 5, 152,
			164, 166, 222, 119, 72, 228},
	}
	anyValidEACLBytes = []byte{10, 4, 8, 2, 16, 16, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
		62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 26, 57, 8, 204, 253, 149, 6, 16, 182, 183,
		219, 2, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57, 51, 54, 52, 51, 34, 12, 118,
		97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 34, 6, 8, 196, 203, 182, 201, 2, 26, 248, 1, 8, 204, 253, 149, 6, 16, 235,
		218, 232, 20, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57, 51, 54, 52, 51,
		34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 26, 35, 8, 200, 202, 35, 16, 228, 149, 204, 3, 26, 11, 107,
		101, 121, 95, 49, 50, 57, 56, 52, 51, 50, 34, 11, 118, 97, 108, 95, 56, 50, 52, 51, 50, 53, 56, 34, 6, 8, 196, 203, 182, 201,
		2, 34, 151, 1, 18, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58,
		223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18, 25, 53, 121, 249, 124, 139, 174, 30, 193, 143, 226, 163,
		208, 188, 194, 173, 123, 60, 84, 224, 229, 4, 14, 206, 19, 117, 18, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126,
		241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 18, 25, 53, 246, 34, 60, 106, 147,
		200, 106, 111, 144, 9, 61, 86, 46, 111, 148, 91, 65, 206, 216, 139, 168, 188, 23, 102, 18, 25, 53, 164, 213, 123, 6, 115, 90,
		134, 224, 150, 72, 192, 236, 220, 188, 131, 102, 5, 152, 164, 166, 222, 119, 72, 228}
)

// Protojson.
var (
	// corresponds to anyValidFilters.
	anyValidJSONFilters = []string{`
{
 "headerType": 4509681,
 "matchType": 949385,
 "key": "key_54093643",
 "value": "val_34811040"
}
`, `
{
 "headerType": 582984,
 "matchType": 7539428,
 "key": "key_1298432",
 "value": "val_8243258"
}
`}
	// corresponds to anyValidTargets.
	anyValidJSONTargets = []string{`
{
 "role": 690857412,
 "keys": []
}
`, `
{
 "role": "ROLE_UNSPECIFIED",
 "keys": [
  "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
  "NXn5fIuuHsGP4qPQvMKtezxU4OUEDs4TdQ==",
  "ApUrMsRbsT6D6X7xsQ1OYF53RzezCDXxTwIBX1VOLcWI",
  "NfYiPGqTyGpvkAk9Vi5vlFtBztiLqLwXZg==",
  "NaTVewZzWobglkjA7Ny8g2YFmKSm3ndI5A=="
 ]
}
`}
	// corresponds to anyValidRecords.
	anyValidJSONRecords = []string{`
{
 "operation": 5692342,
 "action": 12943052,
 "filters": [
  {
   "headerType": 4509681,
   "matchType": 949385,
   "key": "key_54093643",
   "value": "val_34811040"
  }
 ],
 "targets": [
  {
   "role": 690857412,
   "keys": []
  }
 ]
}
`, `
{
 "operation": 43658603,
 "action": 12943052,
 "filters": [
  {
   "headerType": 4509681,
   "matchType": 949385,
   "key": "key_54093643",
   "value": "val_34811040"
  },
  {
   "headerType": 582984,
   "matchType": 7539428,
   "key": "key_1298432",
   "value": "val_8243258"
  }
 ],
 "targets": [
  {
   "role": 690857412,
   "keys": []
  },
  {
   "role": "ROLE_UNSPECIFIED",
   "keys": [
    "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
    "NXn5fIuuHsGP4qPQvMKtezxU4OUEDs4TdQ==",
    "ApUrMsRbsT6D6X7xsQ1OYF53RzezCDXxTwIBX1VOLcWI",
    "NfYiPGqTyGpvkAk9Vi5vlFtBztiLqLwXZg==",
    "NaTVewZzWobglkjA7Ny8g2YFmKSm3ndI5A=="
   ]
  }
 ]
}
`}
	anyValidEACLJSON = `
{
 "version": {
  "major": 2,
  "minor": 16
 },
 "containerID": {
  "value": "8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="
 },
 "records": [
  {
   "operation": 5692342,
   "action": 12943052,
   "filters": [
    {
     "headerType": 4509681,
     "matchType": 949385,
     "key": "key_54093643",
     "value": "val_34811040"
    }
   ],
   "targets": [
    {
     "role": 690857412,
     "keys": []
    }
   ]
  },
  {
   "operation": 43658603,
   "action": 12943052,
   "filters": [
    {
     "headerType": 43658603,
     "matchType": 949385,
     "key": "key_54093643",
     "value": "val_34811040"
    },
    {
     "headerType": 582984,
     "matchType": 7539428,
     "key": "key_1298432",
     "value": "val_8243258"
    }
   ],
   "targets": [
    {
     "role": 690857412,
     "keys": []
    },
    {
     "role": "ROLE_UNSPECIFIED",
     "keys": [
      "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
      "NXn5fIuuHsGP4qPQvMKtezxU4OUEDs4TdQ==",
      "ApUrMsRbsT6D6X7xsQ1OYF53RzezCDXxTwIBX1VOLcWI",
      "NfYiPGqTyGpvkAk9Vi5vlFtBztiLqLwXZg==",
      "NaTVewZzWobglkjA7Ny8g2YFmKSm3ndI5A=="
     ]
    }
   ]
  }
 ]
}
`
)

func init() {
	for i := range anyECDSAPublicKeys {
		anyECDSAPublicKeysPtr = append(anyECDSAPublicKeysPtr, &anyECDSAPublicKeys[i])
	}
}

func assertProtoTargetsEqual(t testing.TB, ts []eacl.Target, ms []protoacl.Target) {
	require.Len(t, ms, len(ts))
	for i := range ts {
		require.EqualValues(t, ts[i].Role(), ms[i].GetRole(), i)
		require.Equal(t, ts[i].RawSubjects(), ms[i].GetKeys(), i)
	}
}

func assertProtoFiltersEqual(t testing.TB, fs []eacl.Filter, ms []protoacl.HeaderFilter) {
	require.Len(t, ms, len(fs))
	for i := range fs {
		require.EqualValues(t, fs[i].From(), ms[i].GetHeaderType(), i)
		require.Equal(t, fs[i].Key(), ms[i].GetKey(), i)
		require.EqualValues(t, fs[i].Matcher(), ms[i].GetMatchType(), i)
		require.Equal(t, fs[i].Value(), ms[i].GetValue(), i)
	}
}

func assertProtoRecordsEqual(t testing.TB, rs []eacl.Record, ms []protoacl.Record) {
	require.Len(t, ms, len(rs))
	for i := range rs {
		require.EqualValues(t, rs[i].Action(), ms[i].GetAction(), i)
		require.EqualValues(t, rs[i].Operation(), ms[i].GetOperation(), i)
		assertProtoTargetsEqual(t, rs[i].Targets(), ms[i].GetTargets())
		assertProtoFiltersEqual(t, rs[i].Filters(), ms[i].GetFilters())
	}
}
