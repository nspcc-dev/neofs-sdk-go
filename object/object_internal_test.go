package object

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func sessionToken(cnr cid.ID) *session.Object {
	var sess session.Object
	sess.SetID(uuid.New())
	sess.ForVerb(session.VerbObjectPut)
	sess.BindContainer(cnr)

	return &sess
}

func parenObject(cnr cid.ID, owner user.ID) *Object {
	var obj Object

	obj.InitCreation(RequiredFields{
		Container: cnr,
		Owner:     owner,
	})

	return &obj
}

func TestObject_CopyTo(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)

	var obj Object
	cnr := cidtest.ID()
	own := usertest.ID(t)

	obj.InitCreation(RequiredFields{
		Container: cnr,
		Owner:     own,
	})

	var attr Attribute
	attr.SetKey("key")
	attr.SetValue("value")

	obj.SetAttributes(attr)
	obj.SetPayload([]byte{1, 2, 3})
	obj.SetSessionToken(sessionToken(cnr))
	obj.SetCreationEpoch(10)
	obj.SetParent(parenObject(cnr, own))
	obj.SetChildren(oidtest.ID(), oidtest.ID(), oidtest.ID())

	var splitID SplitID
	splitID.SetUUID(uuid.New())
	obj.SetSplitID(&splitID)

	v := version.Current()
	obj.SetVersion(&v)

	require.NoError(t, obj.CalculateAndSetID())
	require.NoError(t, obj.Sign(signer))

	t.Run("copy", func(t *testing.T) {
		var dst Object
		obj.CopyTo(&dst)

		checkObjectEquals(t, obj, dst)
	})

	t.Run("change id", func(t *testing.T) {
		var dst Object
		obj.CopyTo(&dst)

		dstHeader := dst.ToV2().GetHeader()
		require.NotNil(t, dstHeader)
		dstHeader.SetObjectType(object.TypeTombstone)

		objHeader := obj.ToV2().GetHeader()

		require.NotEqual(t, dstHeader.GetObjectType(), objHeader.GetObjectType())
	})

	t.Run("overwrite id", func(t *testing.T) {
		var local Object
		_, isSet := local.ID()
		require.False(t, isSet)

		var dst Object
		require.NoError(t, dst.CalculateAndSetID())
		_, isSet = dst.ID()
		require.True(t, isSet)

		local.CopyTo(&dst)

		_, isSet = local.ID()
		require.False(t, isSet)
		_, isSet = dst.ID()
		require.False(t, isSet)

		checkObjectEquals(t, local, dst)

		require.NoError(t, dst.CalculateAndSetID())
		_, isSet = dst.ID()
		require.True(t, isSet)

		_, isSet = local.ID()
		require.False(t, isSet)
	})

	t.Run("change payload", func(t *testing.T) {
		var dst Object
		obj.CopyTo(&dst)

		require.True(t, bytes.Equal(dst.Payload(), obj.Payload()))

		p := dst.Payload()
		p[0] = 4

		require.False(t, bytes.Equal(dst.Payload(), obj.Payload()))
	})

	t.Run("overwrite signature", func(t *testing.T) {
		var local Object
		require.Nil(t, local.Signature())

		var dst Object
		require.NoError(t, dst.CalculateAndSetID())
		require.NoError(t, dst.Sign(signer))
		require.NotNil(t, dst.Signature())

		local.CopyTo(&dst)
		require.Nil(t, local.Signature())
		require.Nil(t, dst.Signature())

		checkObjectEquals(t, local, dst)

		require.NoError(t, dst.CalculateAndSetID())
		require.NoError(t, dst.Sign(signer))
		require.NotNil(t, dst.Signature())
		require.Nil(t, local.Signature())
	})

	t.Run("overwrite header", func(t *testing.T) {
		var local Object
		require.Nil(t, local.ToV2().GetHeader())

		var dst Object
		dst.SetAttributes(attr)
		require.NotNil(t, dst.ToV2().GetHeader())

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)

		require.Nil(t, local.ToV2().GetHeader())
		require.Nil(t, dst.ToV2().GetHeader())

		dst.SetAttributes(attr)
		require.NotNil(t, dst.ToV2().GetHeader())
		require.Nil(t, local.ToV2().GetHeader())
	})

	t.Run("header, rewrite container id to nil", func(t *testing.T) {
		var local Object
		var localHeader object.Header
		local.ToV2().SetHeader(&localHeader)

		var dstContID refs.ContainerID
		dstContID.SetValue([]byte{1})

		var dstHeader object.Header
		dstHeader.SetContainerID(&dstContID)

		var dst Object
		dst.ToV2().SetHeader(&dstHeader)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, change container id", func(t *testing.T) {
		var localContID refs.ContainerID
		localContID.SetValue([]byte{1})

		var localHeader object.Header
		localHeader.SetContainerID(&localContID)

		var local Object
		local.ToV2().SetHeader(&localHeader)

		var dstContID refs.ContainerID
		dstContID.SetValue([]byte{2})

		var dstHeader object.Header
		dstHeader.SetContainerID(&dstContID)

		var dst Object
		dst.ToV2().SetHeader(&dstHeader)

		require.NotEqual(t, local.ToV2().GetHeader().GetContainerID(), dst.ToV2().GetHeader().GetContainerID())

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)

		local.ToV2().GetHeader().GetContainerID().SetValue([]byte{3})
		require.NotEqual(t, local.ToV2().GetHeader().GetContainerID(), dst.ToV2().GetHeader().GetContainerID())
	})

	t.Run("header, rewrite payload hash", func(t *testing.T) {
		var cs refs.Checksum
		cs.SetType(refs.TillichZemor)
		cs.SetSum([]byte{1})

		var localHeader object.Header
		localHeader.SetPayloadHash(&cs)

		var local Object
		local.ToV2().SetHeader(&localHeader)

		var dst Object
		local.CopyTo(&dst)

		checkObjectEquals(t, local, dst)
	})

	t.Run("header, rewrite homo hash", func(t *testing.T) {
		var cs refs.Checksum
		cs.SetType(refs.TillichZemor)
		cs.SetSum([]byte{1})

		var localHeader object.Header
		localHeader.SetHomomorphicHash(&cs)

		var local Object
		local.ToV2().SetHeader(&localHeader)

		var dst Object
		local.CopyTo(&dst)

		checkObjectEquals(t, local, dst)
	})

	t.Run("header, rewrite split header", func(t *testing.T) {
		var spl object.SplitHeader

		var localHeader object.Header
		localHeader.SetSplit(&spl)

		var local Object
		local.ToV2().SetHeader(&localHeader)

		var dst Object
		dst.SetChildren(oidtest.ID(), oidtest.ID())

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session owner", func(t *testing.T) {
		var local Object
		sess := sessionToken(cnr)
		sess.SetIssuer(signer.UserID())

		local.SetSessionToken(sess)

		var dst Object

		require.NotEqual(t,
			local.ToV2().GetHeader().GetSessionToken().GetBody().GetOwnerID(),
			dst.ToV2().GetHeader().GetSessionToken().GetBody().GetOwnerID(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session owner to nil", func(t *testing.T) {
		var local Object
		local.SetSessionToken(sessionToken(cnr))

		sess := sessionToken(cnr)
		sess.SetIssuer(signer.UserID())

		var dst Object
		dst.SetSessionToken(sess)

		require.NotEqual(t,
			local.ToV2().GetHeader().GetSessionToken().GetBody().GetOwnerID(),
			dst.ToV2().GetHeader().GetSessionToken().GetBody().GetOwnerID(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session lifetime", func(t *testing.T) {
		var local Object
		sess := sessionToken(cnr)
		sess.SetExp(1234)

		local.SetSessionToken(sess)

		var dst Object

		require.NotEqual(t,
			local.ToV2().GetHeader().GetSessionToken().GetBody().GetLifetime(),
			dst.ToV2().GetHeader().GetSessionToken().GetBody().GetLifetime(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, overwrite session body", func(t *testing.T) {
		var local Object
		sessLocal := sessionToken(cnr)
		local.SetSessionToken(sessLocal)

		local.ToV2().GetHeader().GetSessionToken().SetBody(nil)

		sessDst := sessionToken(cnr)
		sessDst.SetID(uuid.New())

		var dst Object
		dst.SetSessionToken(sessDst)

		require.NotEqual(t,
			local.ToV2().GetHeader().GetSessionToken().GetBody(),
			dst.ToV2().GetHeader().GetSessionToken().GetBody(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})
}

func checkObjectEquals(t *testing.T, local, dst Object) {
	bts, err := local.Marshal()
	require.NoError(t, err)

	bts2, err := dst.Marshal()
	require.NoError(t, err)

	require.Equal(t, local, dst)
	require.True(t, bytes.Equal(bts, bts2))
}
