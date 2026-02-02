package object

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
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

func sessionTokenV2(cnr cid.ID) *sessionv2.Token {
	var sess sessionv2.Token
	ctx, _ := sessionv2.NewContext(cnr, []sessionv2.Verb{sessionv2.VerbObjectPut})
	_ = sess.SetContexts([]sessionv2.Context{ctx})

	return &sess
}

func parenObject(cnr cid.ID, owner user.ID) *Object {
	return New(cnr, owner)
}

func TestObject_CopyTo(t *testing.T) {
	usr := usertest.User()

	var (
		attr Attribute
		cnr  = cidtest.ID()
		own  = usertest.ID()
		obj  = New(cnr, own)
	)

	attr.SetKey("key")
	attr.SetValue("value")

	obj.SetAttributes(attr)
	obj.SetPayload([]byte{1, 2, 3})
	obj.SetSessionToken(sessionToken(cnr))
	obj.SetSessionTokenV2(sessionTokenV2(cnr))
	obj.SetCreationEpoch(10)
	obj.SetParent(parenObject(cnr, own))
	obj.SetChildren(oidtest.ID(), oidtest.ID(), oidtest.ID())

	var splitID SplitID
	splitID.SetUUID(uuid.New())
	obj.SetSplitID(&splitID)

	v := version.Current()
	obj.SetVersion(&v)

	require.NoError(t, obj.CalculateAndSetID())
	require.NoError(t, obj.Sign(usr))

	t.Run("copy", func(t *testing.T) {
		var dst Object
		obj.CopyTo(&dst)

		checkObjectEquals(t, *obj, dst)
	})

	t.Run("change id", func(t *testing.T) {
		var dst Object
		obj.CopyTo(&dst)

		dstHeader := dst.ProtoMessage().GetHeader()
		require.NotNil(t, dstHeader)
		dstHeader.ObjectType = protoobject.ObjectType_TOMBSTONE

		objHeader := obj.ProtoMessage().GetHeader()

		require.NotEqual(t, dstHeader.GetObjectType(), objHeader.GetObjectType())
	})

	t.Run("overwrite id", func(t *testing.T) {
		var local Object
		require.True(t, local.GetID().IsZero())

		var dst Object
		require.NoError(t, dst.CalculateAndSetID())
		require.False(t, dst.GetID().IsZero())

		local.CopyTo(&dst)

		require.True(t, local.GetID().IsZero())
		require.True(t, dst.GetID().IsZero())

		checkObjectEquals(t, local, dst)

		require.NoError(t, dst.CalculateAndSetID())
		require.False(t, dst.GetID().IsZero())
		require.True(t, local.GetID().IsZero())
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
		require.NoError(t, dst.Sign(usr))
		require.NotNil(t, dst.Signature())

		local.CopyTo(&dst)
		require.Nil(t, local.Signature())
		require.Nil(t, dst.Signature())

		checkObjectEquals(t, local, dst)

		require.NoError(t, dst.CalculateAndSetID())
		require.NoError(t, dst.Sign(usr))
		require.NotNil(t, dst.Signature())
		require.Nil(t, local.Signature())
	})

	t.Run("overwrite header", func(t *testing.T) {
		var local Object
		require.Nil(t, local.ProtoMessage().GetHeader())

		var dst Object
		dst.SetAttributes(attr)
		require.NotNil(t, dst.ProtoMessage().GetHeader())

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)

		require.Nil(t, local.ProtoMessage().GetHeader())
		require.Nil(t, dst.ProtoMessage().GetHeader())

		dst.SetAttributes(attr)
		require.NotNil(t, dst.ProtoMessage().GetHeader())
		require.Nil(t, local.ProtoMessage().GetHeader())
	})

	t.Run("header, rewrite container id to nil", func(t *testing.T) {
		var local Object
		var localHeader protoobject.Header
		local.ProtoMessage().Header = &localHeader

		var dstContID refs.ContainerID
		dstContID.Value = []byte{1}

		var dstHeader protoobject.Header
		dstHeader.ContainerId = &dstContID

		var dst Object
		dst.ProtoMessage().Header = &dstHeader

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, change container id", func(t *testing.T) {
		c := cidtest.ID()

		var local, dst Object
		local.header.cnr = c
		dst.header.cnr = c
		require.Equal(t, local.GetContainerID(), dst.GetContainerID())

		local.CopyTo(&dst)
		local.header.cnr[0]++
		require.NotEqual(t, local.GetContainerID(), dst.GetContainerID())
		require.Equal(t, c, dst.GetContainerID())
	})

	t.Run("header, rewrite payload hash", func(t *testing.T) {
		localHeader := protoobject.Header{
			PayloadHash: &refs.Checksum{
				Type: refs.ChecksumType_TZ,
				Sum:  []byte{1},
			},
		}

		var local Object
		local.ProtoMessage().Header = &localHeader

		var dst Object
		local.CopyTo(&dst)

		checkObjectEquals(t, local, dst)
	})

	t.Run("header, rewrite homo hash", func(t *testing.T) {
		localHeader := protoobject.Header{
			HomomorphicHash: &refs.Checksum{
				Type: refs.ChecksumType_TZ,
				Sum:  []byte{1},
			},
		}

		var local Object
		local.ProtoMessage().Header = &localHeader

		var dst Object
		local.CopyTo(&dst)

		checkObjectEquals(t, local, dst)
	})

	t.Run("header, rewrite split header", func(t *testing.T) {
		localHeader := protoobject.Header{Split: new(protoobject.Header_Split)}

		var local Object
		local.ProtoMessage().Header = &localHeader

		var dst Object
		dst.SetChildren(oidtest.ID(), oidtest.ID())

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session owner", func(t *testing.T) {
		var local Object
		sess := sessionToken(cnr)
		sess.SetIssuer(usr.UserID())

		local.SetSessionToken(sess)

		var dst Object

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetOwnerId(),
			dst.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetOwnerId(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session owner to nil", func(t *testing.T) {
		var local Object
		local.SetSessionToken(sessionToken(cnr))

		sess := sessionToken(cnr)
		sess.SetIssuer(usr.UserID())

		var dst Object
		dst.SetSessionToken(sess)

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetOwnerId(),
			dst.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetOwnerId(),
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
			local.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetLifetime(),
			dst.ProtoMessage().GetHeader().GetSessionToken().GetBody().GetLifetime(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, overwrite session body", func(t *testing.T) {
		var local Object
		sessLocal := sessionToken(cnr)
		local.SetSessionToken(sessLocal)

		local.ProtoMessage().GetHeader().GetSessionToken().Body = nil

		sessDst := sessionToken(cnr)
		sessDst.SetID(uuid.New())

		var dst Object
		dst.SetSessionToken(sessDst)

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionToken().GetBody(),
			dst.ProtoMessage().GetHeader().GetSessionToken().GetBody(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session v2 owner", func(t *testing.T) {
		var local Object
		sess := sessionTokenV2(cnr)
		sess.SetIssuer(usr.UserID())

		local.SetSessionTokenV2(sess)

		var dst Object

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetIssuer(),
			dst.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetIssuer(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session v2 owner to nil", func(t *testing.T) {
		var local Object
		local.SetSessionTokenV2(sessionTokenV2(cnr))

		sess := sessionTokenV2(cnr)
		sess.SetIssuer(usr.UserID())

		var dst Object
		dst.SetSessionTokenV2(sess)

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetIssuer(),
			dst.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetIssuer(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, set session v2 lifetime", func(t *testing.T) {
		var local Object
		sess := sessionTokenV2(cnr)
		sess.SetExp(time.Unix(1234, 0))

		local.SetSessionTokenV2(sess)

		var dst Object

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetLifetime(),
			dst.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody().GetLifetime(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})

	t.Run("header, overwrite session v2 body", func(t *testing.T) {
		var local Object
		sessLocal := sessionTokenV2(cnr)
		local.SetSessionTokenV2(sessLocal)

		local.ProtoMessage().GetHeader().GetSessionTokenV2().Body = nil

		sessDst := sessionTokenV2(cnr)
		require.NoError(t, sessDst.SetAppData([]byte("some data")))

		var dst Object
		dst.SetSessionTokenV2(sessDst)

		require.NotEqual(t,
			local.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody(),
			dst.ProtoMessage().GetHeader().GetSessionTokenV2().GetBody(),
		)

		local.CopyTo(&dst)
		checkObjectEquals(t, local, dst)
	})
}

func checkObjectEquals(t *testing.T, local, dst Object) {
	require.Equal(t, local, dst)
	require.True(t, bytes.Equal(local.Marshal(), dst.Marshal()))
}
