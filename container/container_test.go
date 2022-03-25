package container_test

import (
	"testing"

	"github.com/google/uuid"
	containerv2 "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	sigtest "github.com/nspcc-dev/neofs-sdk-go/signature/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func TestNewContainer(t *testing.T) {
	var c container.Container

	nonce := uuid.New()

	ownerID := ownertest.ID()
	policy := netmaptest.PlacementPolicy()

	c.SetBasicACL(acl.PublicBasicRule)

	attrs := containertest.Attributes()
	c.SetAttributes(attrs)

	c.SetPlacementPolicy(policy)
	c.SetNonceUUID(nonce)
	c.SetOwnerID(ownerID)

	ver := versiontest.Version()
	c.SetVersion(ver)

	var v2 containerv2.Container
	c.WriteToV2(&v2)

	var newContainer container.Container
	newContainer.ReadFromV2(v2)

	require.EqualValues(t, newContainer.PlacementPolicy(), policy)
	require.EqualValues(t, newContainer.Attributes(), attrs)
	require.EqualValues(t, newContainer.BasicACL(), acl.PublicBasicRule)

	newNonce, err := newContainer.NonceUUID()
	require.NoError(t, err)

	require.EqualValues(t, newNonce, nonce)
	require.EqualValues(t, newContainer.OwnerID(), ownerID)
	require.EqualValues(t, newContainer.Version(), ver)
}

func TestContainerEncoding(t *testing.T) {
	c := containertest.Container()

	t.Run("binary", func(t *testing.T) {
		data, err := c.Marshal()
		require.NoError(t, err)

		c2 := container.InitCreation()
		require.NoError(t, c2.Unmarshal(data))

		require.Equal(t, c, c2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := c.MarshalJSON()
		require.NoError(t, err)

		c2 := container.InitCreation()
		require.NoError(t, c2.UnmarshalJSON(data))

		require.Equal(t, c, c2)
	})
}

func TestContainer_SessionToken(t *testing.T) {
	tok := sessiontest.Token()

	var cnr container.Container

	cnr.SetSessionToken(tok)

	require.Equal(t, tok, cnr.SessionToken())
}

func TestContainer_Signature(t *testing.T) {
	sig := sigtest.Signature()

	var cnr container.Container
	cnr.SetSignature(sig)

	require.Equal(t, sig, cnr.Signature())
}

func TestContainer_ToV2(t *testing.T) {
	t.Run("zero to V2", func(t *testing.T) {
		var (
			x  = container.InitCreation()
			v2 containerv2.Container
		)

		x.WriteToV2(&v2)

		nonce, err := x.NonceUUID()
		require.NoError(t, err)

		require.Equal(t, nonce[:], v2.GetNonce())
		require.Empty(t, v2.GetAttributes())
		require.Equal(t, version.Current().ToV2(), v2.GetVersion())
		require.Nil(t, v2.GetOwnerID())
		require.Equal(t, acl.PrivateBasicRule, acl.BasicACL(v2.GetBasicACL()))
	})

	t.Run("default values", func(t *testing.T) {
		cnt := container.InitCreation()

		// check initial values
		require.Nil(t, cnt.SessionToken())
		require.Nil(t, cnt.Signature())
		require.Empty(t, cnt.Attributes())
		require.Nil(t, cnt.PlacementPolicy())
		require.Nil(t, cnt.OwnerID())

		require.EqualValues(t, acl.PrivateBasicRule, cnt.BasicACL())
		require.Equal(t, version.Current(), cnt.Version())

		nonce, err := cnt.NonceUUID()
		require.NoError(t, err)
		require.NotNil(t, nonce)

		// convert to v2 message
		var cntV2 containerv2.Container
		cnt.WriteToV2(&cntV2)

		nonceV2, err := uuid.FromBytes(cntV2.GetNonce())
		require.NoError(t, err)

		require.Equal(t, nonce.String(), nonceV2.String())

		require.Empty(t, cntV2.GetAttributes())
		require.Nil(t, cntV2.GetPlacementPolicy())
		require.Nil(t, cntV2.GetOwnerID())

		require.Equal(t, uint32(acl.PrivateBasicRule), cntV2.GetBasicACL())
		require.Equal(t, version.Current().ToV2(), cntV2.GetVersion())
	})
}
