package container_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	sigtest "github.com/nspcc-dev/neofs-sdk-go/signature/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
)

func TestNewContainer(t *testing.T) {
	c := container.New()

	nonce := uuid.New()

	ownerID := ownertest.GenerateID()
	policy := netmaptest.PlacementPolicy()

	c.SetBasicACL(acl.PublicBasicRule)

	attrs := containertest.Attributes()
	c.SetAttributes(attrs)

	c.SetPlacementPolicy(policy)
	c.SetNonceUUID(nonce)
	c.SetOwnerID(ownerID)

	ver := versiontest.Version()
	c.SetVersion(ver)

	v2 := c.ToV2()
	newContainer := container.NewContainerFromV2(v2)

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

		c2 := container.New()
		require.NoError(t, c2.Unmarshal(data))

		require.Equal(t, c, c2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := c.MarshalJSON()
		require.NoError(t, err)

		c2 := container.New()
		require.NoError(t, c2.UnmarshalJSON(data))

		require.Equal(t, c, c2)
	})
}

func TestContainer_SessionToken(t *testing.T) {
	tok := sessiontest.Generate()

	cnr := container.New()

	cnr.SetSessionToken(tok)

	require.Equal(t, tok, cnr.SessionToken())
}

func TestContainer_Signature(t *testing.T) {
	sig := sigtest.Signature()

	cnr := container.New()
	cnr.SetSignature(sig)

	require.Equal(t, sig, cnr.Signature())
}

func TestContainer_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *container.Container

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		cnt := container.New()

		// check initial values
		require.Nil(t, cnt.SessionToken())
		require.Nil(t, cnt.Signature())
		require.Nil(t, cnt.Version())
		require.Nil(t, cnt.Attributes())
		require.Nil(t, cnt.PlacementPolicy())
		require.Nil(t, cnt.OwnerID())

		require.EqualValues(t, acl.PrivateBasicRule, cnt.BasicACL())

		nonce, err := cnt.NonceUUID()
		require.NoError(t, err)
		require.NotNil(t, nonce)

		// convert to v2 message
		cntV2 := cnt.ToV2()

		nonceV2, err := uuid.FromBytes(cntV2.GetNonce())
		require.NoError(t, err)

		require.Equal(t, nonce.String(), nonceV2.String())

		require.Nil(t, cntV2.GetVersion())
		require.Nil(t, cntV2.GetAttributes())
		require.Nil(t, cntV2.GetPlacementPolicy())
		require.Nil(t, cntV2.GetOwnerID())

		require.Equal(t, uint32(acl.PrivateBasicRule), cntV2.GetBasicACL())
	})
}
