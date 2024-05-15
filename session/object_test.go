package session_test

import (
	mrand "math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/stretchr/testify/require"
)

var knownObjectVerbs = map[session.ObjectVerb]apisession.ObjectSessionContext_Verb{
	session.VerbObjectPut:       apisession.ObjectSessionContext_PUT,
	session.VerbObjectGet:       apisession.ObjectSessionContext_GET,
	session.VerbObjectHead:      apisession.ObjectSessionContext_HEAD,
	session.VerbObjectDelete:    apisession.ObjectSessionContext_DELETE,
	session.VerbObjectSearch:    apisession.ObjectSessionContext_SEARCH,
	session.VerbObjectRange:     apisession.ObjectSessionContext_RANGE,
	session.VerbObjectRangeHash: apisession.ObjectSessionContext_RANGEHASH,
}

func setRequiredObjectAPIFields(o *session.Object) {
	setRequiredTokenAPIFields(o)
	o.BindContainer(cidtest.ID())
}

func TestObjectDecoding(t *testing.T) {
	testDecoding(t, sessiontest.Object, []invalidAPITestCase{
		{name: "body/context/wrong", err: "wrong context field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context = new(apisession.SessionToken_Body_Container)
		}},
		{name: "body/context/target/container/value/nil", err: "invalid context: invalid target container: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = nil
		}},
		{name: "body/context/target/container/value/empty", err: "invalid context: invalid target container: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = []byte{}
		}},
		{name: "body/context/target/container/value/wrong length", err: "invalid context: invalid target container: invalid value length 31", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container.Value = make([]byte, 31)
		}},
		{name: "body/context/target/objects/value/nil", err: "invalid context: invalid target object #1: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
				{Value: make([]byte, 32)}, {Value: nil}}
		}},
		{name: "body/context/target/objects/value/empty", err: "invalid context: invalid target object #1: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
				{Value: make([]byte, 32)}, {Value: nil}}
		}},
		{name: "body/context/target/objects/value/wrong length", err: "invalid context: invalid target object #1: invalid value length 31", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects = []*refs.ObjectID{
				{Value: make([]byte, 32)}, {Value: make([]byte, 31)}}
		}},
	})
}

func TestObject_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var c session.Object
		msg := []byte("definitely_not_protobuf")
		err := c.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
}

func TestObject_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var c session.Object
		msg := []byte("definitely_not_protojson")
		err := c.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
}

func TestObject_CopyTo(t *testing.T) {
	testCopyTo(t, sessiontest.Object())
}

func TestObject_UnmarshalSignedData(t *testing.T) {
	testSignedData(t, sessiontest.ObjectUnsigned())
}

func TestObject_SetAuthKey(t *testing.T) {
	testAuthKey(t, setRequiredObjectAPIFields)
}

func TestObject_Sign(t *testing.T) {
	testSign(t, setRequiredObjectAPIFields)
}

func TestObject_SetSignature(t *testing.T) {
	testSetSignature(t, setRequiredObjectAPIFields)
}

func TestObject_SetIssuer(t *testing.T) {
	testIssuer(t, setRequiredObjectAPIFields)
}

func TestObject_SetID(t *testing.T) {
	testID(t, setRequiredObjectAPIFields)
}

func TestObject_InvalidAt(t *testing.T) {
	testInvalidAt(t, session.Object{})
}

func TestObject_SetExp(t *testing.T) {
	testLifetimeField(t, session.Object.Exp, (*session.Object).SetExp, (*apisession.SessionToken_Body_TokenLifetime).GetExp, setRequiredObjectAPIFields)
}

func TestObject_SetNbf(t *testing.T) {
	testLifetimeField(t, session.Object.Nbf, (*session.Object).SetNbf, (*apisession.SessionToken_Body_TokenLifetime).GetNbf, setRequiredObjectAPIFields)
}

func TestObject_SetIat(t *testing.T) {
	testLifetimeField(t, session.Object.Iat, (*session.Object).SetIat, (*apisession.SessionToken_Body_TokenLifetime).GetIat, setRequiredObjectAPIFields)
}

func TestObject_ExpiredAt(t *testing.T) {
	var o session.Object
	require.True(t, o.ExpiredAt(0))
	require.True(t, o.ExpiredAt(1))

	o.SetExp(1)
	require.False(t, o.ExpiredAt(0))
	require.False(t, o.ExpiredAt(1))
	require.True(t, o.ExpiredAt(2))
	require.True(t, o.ExpiredAt(3))

	o.SetExp(2)
	require.False(t, o.ExpiredAt(0))
	require.False(t, o.ExpiredAt(1))
	require.False(t, o.ExpiredAt(2))
	require.True(t, o.ExpiredAt(3))
}

func TestObject_ForVerb(t *testing.T) {
	var c session.Object

	for verb := range knownObjectVerbs {
		require.False(t, c.AssertVerb(verb))
		c.ForVerb(verb)
		require.True(t, c.AssertVerb(verb))
	}

	verb := session.ObjectVerb(mrand.Uint32() % 256)
	verbOther := verb + 1
	c.ForVerb(verb)
	require.True(t, c.AssertVerb(verb))
	require.False(t, c.AssertVerb(verbOther))

	c.ForVerb(verbOther)
	require.False(t, c.AssertVerb(verb))
	require.True(t, c.AssertVerb(verbOther))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst session.Object

			dst.ForVerb(verb)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			dst.ForVerb(verbOther)
			src.ForVerb(verb)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst session.Object
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredObjectAPIFields(&src)

			dst.ForVerb(verb)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Verb)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			dst.ForVerb(verbOther)
			src.ForVerb(verb)
			src.WriteToV2(&msg)
			require.EqualValues(t, verb, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Verb)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			for verb, apiVerb := range knownObjectVerbs {
				src.ForVerb(verb)
				src.WriteToV2(&msg)
				require.EqualValues(t, apiVerb, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Verb)
				require.NoError(t, dst.ReadFromV2(&msg))
				require.True(t, dst.AssertVerb(verb))
			}
		})
		t.Run("json", func(t *testing.T) {
			var src, dst session.Object

			dst.ForVerb(verb)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			dst.ForVerb(verbOther)
			src.ForVerb(verb)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))
		})
	})
}

func TestObject_BindContainer(t *testing.T) {
	var c session.Object

	cnr := cidtest.ID()
	cnrOther := cidtest.ChangeID(cnr)
	require.False(t, c.AssertContainer(cnr))
	require.False(t, c.AssertContainer(cnrOther))

	c.BindContainer(cnr)
	require.True(t, c.AssertContainer(cnr))
	require.False(t, c.AssertContainer(cnrOther))

	c.BindContainer(cnrOther)
	require.False(t, c.AssertContainer(cnr))
	require.True(t, c.AssertContainer(cnrOther))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst session.Object

			dst.BindContainer(cnr)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.AssertContainer(cnr))
			require.False(t, dst.AssertContainer(cnrOther))

			src.BindContainer(cnr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AssertContainer(cnr))
			require.False(t, dst.AssertContainer(cnrOther))

			src.BindContainer(cnrOther)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.AssertContainer(cnr))
			require.True(t, dst.AssertContainer(cnrOther))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst session.Object
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredTokenAPIFields(&src)

			src.BindContainer(cnr)
			src.WriteToV2(&msg)
			require.Equal(t, &refs.ContainerID{Value: cnr[:]}, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AssertContainer(cnr))
			require.False(t, dst.AssertContainer(cnrOther))

			src.BindContainer(cnrOther)
			src.WriteToV2(&msg)
			require.Equal(t, &refs.ContainerID{Value: cnrOther[:]}, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Container)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, dst.AssertContainer(cnr))
			require.True(t, dst.AssertContainer(cnrOther))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst session.Object

			dst.BindContainer(cnr)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.AssertContainer(cnr))
			require.False(t, dst.AssertContainer(cnrOther))

			src.BindContainer(cnr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AssertContainer(cnr))
			require.False(t, dst.AssertContainer(cnrOther))

			src.BindContainer(cnrOther)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.AssertContainer(cnr))
			require.True(t, dst.AssertContainer(cnrOther))
		})
	})
}

func TestObject_LimitByObjects(t *testing.T) {
	var c session.Object
	checkMatch := func(c session.Object, objs []oid.ID) {
		for i := range objs {
			require.True(t, c.AssertObject(objs[i]))
		}
	}
	checkMismatch := func(c session.Object, objs []oid.ID) {
		for i := range objs {
			require.False(t, c.AssertObject(objs[i]))
		}
	}

	objs := oidtest.NIDs(4)
	checkMatch(c, objs)

	c.LimitByObjects(objs[:2])
	checkMatch(c, objs[:2])
	checkMismatch(c, objs[2:])

	c.LimitByObjects(objs[2:])
	checkMismatch(c, objs[:2])
	checkMatch(c, objs[2:])

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst session.Object

			dst.LimitByObjects(oidtest.NIDs(3))
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			checkMatch(dst, objs)

			src.LimitByObjects(objs[:2])
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			checkMatch(dst, objs[:2])
			checkMismatch(dst, objs[2:])

			src.LimitByObjects(objs[2:])
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			checkMismatch(dst, objs[:2])
			checkMatch(dst, objs[2:])
		})
		t.Run("api", func(t *testing.T) {
			var src, dst session.Object
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredObjectAPIFields(&src)

			dst.LimitByObjects(oidtest.NIDs(3))
			src.WriteToV2(&msg)
			require.Zero(t, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects)
			require.NoError(t, dst.ReadFromV2(&msg))
			checkMatch(dst, objs)

			src.LimitByObjects(objs[:2])
			src.WriteToV2(&msg)
			require.ElementsMatch(t, []*refs.ObjectID{
				{Value: objs[0][:]},
				{Value: objs[1][:]},
			}, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects)
			require.NoError(t, dst.ReadFromV2(&msg))
			checkMatch(dst, objs[:2])
			checkMismatch(dst, objs[2:])

			src.LimitByObjects(objs[2:])
			src.WriteToV2(&msg)
			require.ElementsMatch(t, []*refs.ObjectID{
				{Value: objs[2][:]},
				{Value: objs[3][:]},
			}, msg.Body.Context.(*apisession.SessionToken_Body_Object).Object.Target.Objects)
			require.NoError(t, dst.ReadFromV2(&msg))
			checkMismatch(dst, objs[:2])
			checkMatch(dst, objs[2:])
		})
		t.Run("json", func(t *testing.T) {
			var src, dst session.Object

			dst.LimitByObjects(oidtest.NIDs(3))
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			checkMatch(dst, objs)

			src.LimitByObjects(objs[:2])
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			checkMatch(dst, objs[:2])
			checkMismatch(dst, objs[2:])

			src.LimitByObjects(objs[2:])
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			checkMismatch(dst, objs[:2])
			checkMatch(dst, objs[2:])
		})
	})
}
