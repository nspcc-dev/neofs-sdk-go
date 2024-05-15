package session_test

import (
	mrand "math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

var knownContainerVerbs = map[session.ContainerVerb]apisession.ContainerSessionContext_Verb{
	session.VerbContainerPut:     apisession.ContainerSessionContext_PUT,
	session.VerbContainerDelete:  apisession.ContainerSessionContext_DELETE,
	session.VerbContainerSetEACL: apisession.ContainerSessionContext_SETEACL,
}

func setRequiredContainerAPIFields(c *session.Container) { setRequiredTokenAPIFields(c) }

func TestContainerDecoding(t *testing.T) {
	testDecoding(t, sessiontest.Container, []invalidAPITestCase{
		{name: "body/context/wrong", err: "wrong context field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context = new(apisession.SessionToken_Body_Object)
		}},
		{name: "body/context/wrapped field/nil", err: "invalid context: missing container or wildcard flag", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = nil
		}},
		{name: "body/context/wrapped field/empty", err: "invalid context: missing container or wildcard flag", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = new(apisession.ContainerSessionContext)
		}},
		{name: "body/context/container conflict", err: "invalid context: container conflicts with wildcard flag", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = &apisession.ContainerSessionContext{
				Wildcard:    true,
				ContainerId: new(refs.ContainerID),
			}
		}},
		{name: "body/context/container/value/nil", err: "invalid context: invalid container ID: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = &apisession.ContainerSessionContext{
				ContainerId: &refs.ContainerID{Value: nil},
			}
		}},
		{name: "body/context/container/value/empty", err: "invalid context: invalid container ID: missing value field", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = &apisession.ContainerSessionContext{
				ContainerId: &refs.ContainerID{Value: []byte{}},
			}
		}},
		{name: "body/context/container/value/wrong length", err: "invalid context: invalid container ID: invalid value length 31", corrupt: func(st *apisession.SessionToken) {
			st.Body.Context.(*apisession.SessionToken_Body_Container).Container = &apisession.ContainerSessionContext{
				ContainerId: &refs.ContainerID{Value: make([]byte, 31)},
			}
		}},
	})
}

func TestContainer_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var c session.Container
		msg := []byte("definitely_not_protobuf")
		err := c.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
}

func TestContainer_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var c session.Container
		msg := []byte("definitely_not_protojson")
		err := c.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
}

func TestContainer_CopyTo(t *testing.T) {
	testCopyTo(t, sessiontest.Container())
}

func TestContainer_UnmarshalSignedData(t *testing.T) {
	testSignedData(t, sessiontest.ContainerUnsigned())
}

func TestContainer_SetAuthKey(t *testing.T) {
	testAuthKey(t, setRequiredContainerAPIFields)
}

func TestContainer_Sign(t *testing.T) {
	testSign(t, setRequiredContainerAPIFields)
}

func TestContainer_SetSignature(t *testing.T) {
	testSetSignature(t, setRequiredContainerAPIFields)
}

func TestContainer_SetIssuer(t *testing.T) {
	testIssuer(t, setRequiredContainerAPIFields)
}

func TestContainer_SetID(t *testing.T) {
	testID(t, setRequiredContainerAPIFields)
}

func TestContainer_InvalidAt(t *testing.T) {
	testInvalidAt(t, session.Container{})
}

func TestContainer_SetExp(t *testing.T) {
	testLifetimeField(t, session.Container.Exp, (*session.Container).SetExp, (*apisession.SessionToken_Body_TokenLifetime).GetExp, setRequiredContainerAPIFields)
}

func TestContainer_SetNbf(t *testing.T) {
	testLifetimeField(t, session.Container.Nbf, (*session.Container).SetNbf, (*apisession.SessionToken_Body_TokenLifetime).GetNbf, setRequiredContainerAPIFields)
}

func TestContainer_SetIat(t *testing.T) {
	testLifetimeField(t, session.Container.Iat, (*session.Container).SetIat, (*apisession.SessionToken_Body_TokenLifetime).GetIat, setRequiredContainerAPIFields)
}

func TestIssuedBy(t *testing.T) {
	var c session.Container

	usr, otherUsr := usertest.TwoUsers()
	c.SetIssuer(usr.ID)
	require.True(t, session.IssuedBy(c, usr.ID))
	require.False(t, session.IssuedBy(c, otherUsr.ID))

	require.NoError(t, c.Sign(otherUsr))
	require.False(t, session.IssuedBy(c, usr.ID))
	require.True(t, session.IssuedBy(c, otherUsr.ID))
}

func TestContainer_VerifySessionDataSignature(t *testing.T) {
	var c session.Container
	usr, otherUsr := usertest.TwoUsers()
	someData := []byte("Hello, world!")

	usrSig, err := usr.SignerRFC6979.Sign(someData)
	require.NoError(t, err)
	otherUsrSig, err := otherUsr.SignerRFC6979.Sign(someData)
	require.NoError(t, err)

	require.False(t, c.VerifySessionDataSignature(someData, usrSig))
	require.False(t, c.VerifySessionDataSignature(someData, otherUsrSig))

	c.SetAuthKey(usr.Public())
	require.True(t, c.VerifySessionDataSignature(someData, usrSig))
	require.False(t, c.VerifySessionDataSignature(someData, otherUsrSig))

	c.SetAuthKey(otherUsr.Public())
	require.False(t, c.VerifySessionDataSignature(someData, usrSig))
	require.True(t, c.VerifySessionDataSignature(someData, otherUsrSig))
}

func TestContainer_ApplyOnlyTo(t *testing.T) {
	var c session.Container

	cnr := cidtest.ID()
	cnrOther := cidtest.ChangeID(cnr)
	require.True(t, c.AppliedTo(cnr))
	require.True(t, c.AppliedTo(cnrOther))

	c.ApplyOnlyTo(cnr)
	require.True(t, c.AppliedTo(cnr))
	require.False(t, c.AppliedTo(cnrOther))

	c.ApplyOnlyTo(cnrOther)
	require.False(t, c.AppliedTo(cnr))
	require.True(t, c.AppliedTo(cnrOther))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst session.Container

			dst.ApplyOnlyTo(cnr)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AppliedTo(cnr))
			require.True(t, dst.AppliedTo(cnrOther))

			src.ApplyOnlyTo(cnr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AppliedTo(cnr))
			require.False(t, dst.AppliedTo(cnrOther))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst session.Container
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredContainerAPIFields(&src)

			dst.ApplyOnlyTo(cnr)
			src.WriteToV2(&msg)
			require.True(t, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.Wildcard)
			require.Nil(t, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.ContainerId)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AppliedTo(cnr))
			require.True(t, dst.AppliedTo(cnrOther))

			src.ApplyOnlyTo(cnr)
			src.WriteToV2(&msg)
			require.False(t, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.Wildcard)
			require.Equal(t, &refs.ContainerID{Value: cnr[:]}, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.ContainerId)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AppliedTo(cnr))
			require.False(t, dst.AppliedTo(cnrOther))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst session.Container

			dst.ApplyOnlyTo(cnr)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AppliedTo(cnr))
			require.True(t, dst.AppliedTo(cnrOther))

			src.ApplyOnlyTo(cnr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AppliedTo(cnr))
			require.False(t, dst.AppliedTo(cnrOther))
		})
	})
}

func TestContainer_ForVerb(t *testing.T) {
	var c session.Container

	for verb := range knownContainerVerbs {
		require.False(t, c.AssertVerb(verb))
		c.ForVerb(verb)
		require.True(t, c.AssertVerb(verb))
	}

	verb := session.ContainerVerb(mrand.Uint32() % 256)
	verbOther := verb + 1
	c.ForVerb(verb)
	require.True(t, c.AssertVerb(verb))
	require.False(t, c.AssertVerb(verbOther))

	c.ForVerb(verbOther)
	require.False(t, c.AssertVerb(verb))
	require.True(t, c.AssertVerb(verbOther))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst session.Container

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
			var src, dst session.Container
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredContainerAPIFields(&src)

			dst.ForVerb(verb)
			src.WriteToV2(&msg)
			require.Zero(t, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.Verb)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			dst.ForVerb(verbOther)
			src.ForVerb(verb)
			src.WriteToV2(&msg)
			require.EqualValues(t, verb, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.Verb)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AssertVerb(verb))
			require.False(t, dst.AssertVerb(verbOther))

			for verb, apiVerb := range knownContainerVerbs {
				src.ForVerb(verb)
				src.WriteToV2(&msg)
				require.EqualValues(t, apiVerb, msg.Body.Context.(*apisession.SessionToken_Body_Container).Container.Verb)
				require.NoError(t, dst.ReadFromV2(&msg))
				require.True(t, dst.AssertVerb(verb))
			}
		})
		t.Run("json", func(t *testing.T) {
			var src, dst session.Container

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
