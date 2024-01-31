package eacl_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

func TestNewFilter(t *testing.T) {
	const typ = eacl.HeaderFromRequest
	const key = "any_key"
	const matcher = eacl.MatchStringNotEqual
	const val = "any_value"

	t.Run("empty key", func(t *testing.T) {
		require.Panics(t, func() { eacl.NewFilter(typ, "", matcher, val) })
	})

	f := eacl.NewFilter(typ, key, matcher, val)

	require.Equal(t, typ, f.HeaderType())
	require.Equal(t, key, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, val, f.HeaderValue())
}

func TestNewFilterObjectAttribute(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const value = "any_value"

	t.Run("reserved", func(t *testing.T) {
		require.Panics(t, func() { eacl.NewFilterObjectAttribute("$Object:any", matcher, value) })
	})

	const key = "any_key"

	f := eacl.NewFilterObjectAttribute(key, matcher, value)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, key, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, value, f.HeaderValue())
}

func TestNewFilterObjectVersion(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	var ver version.Version

	ver.SetMajor(123)
	ver.SetMinor(456)

	f := eacl.NewFilterObjectVersion(matcher, ver)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectVersion, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, "v123.456", f.HeaderValue())
}

func TestNewFilterObjectID(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const strID = "CdrUYtHAuDDzFF8iw4mAgN2qqb8SDKPo8Gpyg12Ree2k"

	var id oid.ID
	require.NoError(t, id.DecodeString(strID))

	f := eacl.NewFilterObjectID(matcher, id)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectID, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, strID, f.HeaderValue())
}

func TestNewFilterContainerID(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const strCnr = "AT95KqvYRw3AC1cCmPJdxwYAcXDJGFLv89rZaZKsmJk3"

	var cnr cid.ID
	require.NoError(t, cnr.DecodeString(strCnr))

	f := eacl.NewFilterContainerID(matcher, cnr)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectContainerID, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, strCnr, f.HeaderValue())
}

func TestNewFilterOwnerID(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const strOwner = "AT95KqvYRw3AC1cCmPJdxwYAcXDJGFLv89rZaZKsmJk3"

	var owner user.ID
	require.NoError(t, owner.DecodeString(strOwner))

	f := eacl.NewFilterOwnerID(matcher, owner)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectOwnerID, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, strOwner, f.HeaderValue())
}

func TestNewFilterObjectCreationEpoch(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const epoch = 321

	f := eacl.NewFilterObjectCreationEpoch(matcher, epoch)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectCreationEpoch, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, "321", f.HeaderValue())
}

func TestNewFilterObjectPayloadSize(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const size = 987

	f := eacl.NewFilterObjectPayloadSize(matcher, size)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectPayloadSize, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, "987", f.HeaderValue())
}

func TestNewFilterObjectType(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const typ = object.TypeTombstone

	f := eacl.NewFilterObjectType(matcher, typ)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectType, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, "TOMBSTONE", f.HeaderValue())
}

func TestNewFilterObjectPayloadChecksum(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const strChecksum = "9789ec335ad13e956261e82c5f1acb7fdb7b03d766dc53565bb17ca663f002a6"

	d, err := hex.DecodeString(strChecksum)
	require.NoError(t, err)

	var cs [sha256.Size]byte
	require.Equal(t, len(cs), copy(cs[:], d))

	f := eacl.NewFilterObjectPayloadChecksum(matcher, cs)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectPayloadChecksum, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, strChecksum, f.HeaderValue())
}

func TestNewFilterObjectPayloadHomomorphicChecksum(t *testing.T) {
	const matcher = eacl.MatchStringNotEqual
	const strChecksum = "631e30ec9730f0580192cdaa6edbbeb704adfa4fd37dde65ceb9d0d357242338d14f7bea2a2dcf09b67b5e74e71bb8398954093059e972e5a551a28d28c6653a"

	d, err := hex.DecodeString(strChecksum)
	require.NoError(t, err)

	var cs [tz.Size]byte
	require.Equal(t, len(cs), copy(cs[:], d))

	f := eacl.NewFilterObjectPayloadHomomorphicChecksum(matcher, cs)

	require.Equal(t, eacl.HeaderFromObject, f.HeaderType())
	require.Equal(t, eacl.FilterObjectPayloadHomomorphicChecksum, f.HeaderKey())
	require.Equal(t, matcher, f.Matcher())
	require.Equal(t, strChecksum, f.HeaderValue())
}
