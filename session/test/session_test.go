package sessiontest_test

import (
	"testing"

	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/stretchr/testify/require"
)

func TestContainer(t *testing.T) {
	v := sessiontest.Container()
	require.NotEqual(t, v, sessiontest.Container())
	require.True(t, v.VerifySignature())

	var v2 session.Container
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apisession.SessionToken
	v.WriteToV2(&m)
	var v3 session.Container
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 session.Container
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestContainerUnsigned(t *testing.T) {
	v := sessiontest.ContainerUnsigned()
	require.NotEqual(t, v, sessiontest.ContainerUnsigned())
	require.False(t, v.VerifySignature())

	var v2 session.Container
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v3 session.Container
	require.NoError(t, v3.UnmarshalJSON(j))
	require.Equal(t, v, v3)
}

func TestObject(t *testing.T) {
	v := sessiontest.Object()
	require.NotEqual(t, v, sessiontest.Object())
	require.True(t, v.VerifySignature())

	var v2 session.Object
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apisession.SessionToken
	v.WriteToV2(&m)
	var v3 session.Object
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 session.Object
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestObjectUnsigned(t *testing.T) {
	v := sessiontest.ObjectUnsigned()
	require.NotEqual(t, v, sessiontest.ObjectUnsigned())
	require.False(t, v.VerifySignature())

	var v2 session.Object
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v3 session.Object
	require.NoError(t, v3.UnmarshalJSON(j))
	require.Equal(t, v, v3)
}
