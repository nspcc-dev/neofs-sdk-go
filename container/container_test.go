package container_test

import (
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func TestPlacementPolicyEncoding(t *testing.T) {
	v := containertest.Container()

	t.Run("binary", func(t *testing.T) {
		var v2 container.Container
		require.NoError(t, v2.Unmarshal(v.Marshal()))

		require.Equal(t, v, v2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := v.MarshalJSON()
		require.NoError(t, err)

		var v2 container.Container
		require.NoError(t, v2.UnmarshalJSON(data))

		require.Equal(t, v, v2)
	})
}

func TestContainer_Init(t *testing.T) {
	val := containertest.Container()

	val.Init()

	var ver = val.Version()
	require.Equal(t, version.Current(), ver)

	var msg v2container.Container
	val.WriteToV2(&msg)

	binNonce := msg.GetNonce()

	var nonce uuid.UUID
	require.NoError(t, nonce.UnmarshalBinary(binNonce))
	require.EqualValues(t, 4, nonce.Version())

	verV2 := msg.GetVersion()
	require.NotNil(t, verV2)

	require.NoError(t, ver.ReadFromV2(*verV2))

	require.Equal(t, version.Current(), ver)

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, val, val2)
}

func TestContainer_Owner(t *testing.T) {
	var val container.Container

	require.Zero(t, val.Owner())

	val = containertest.Container()

	owner := usertest.ID()

	val.SetOwner(owner)

	var msg v2container.Container
	val.WriteToV2(&msg)

	var msgOwner refs.OwnerID
	owner.WriteToV2(&msgOwner)

	require.Equal(t, &msgOwner, msg.GetOwnerID())

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.True(t, val2.Owner() == owner)
}

func TestContainer_BasicACL(t *testing.T) {
	var val container.Container

	require.Zero(t, val.BasicACL())

	val = containertest.Container()

	basicACL := containertest.BasicACL()
	val.SetBasicACL(basicACL)

	var msg v2container.Container
	val.WriteToV2(&msg)

	require.EqualValues(t, basicACL.Bits(), msg.GetBasicACL())

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, basicACL, val2.BasicACL())
}

func TestContainer_PlacementPolicy(t *testing.T) {
	var val container.Container

	require.Zero(t, val.PlacementPolicy())

	val = containertest.Container()

	pp := netmaptest.PlacementPolicy()
	val.SetPlacementPolicy(pp)

	var msgPolicy v2netmap.PlacementPolicy
	pp.WriteToV2(&msgPolicy)

	var msg v2container.Container
	val.WriteToV2(&msg)

	require.Equal(t, &msgPolicy, msg.GetPlacementPolicy())

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, pp, val2.PlacementPolicy())
}

func assertContainsAttribute(t *testing.T, m v2container.Container, key, val string) {
	var msgAttr v2container.Attribute

	msgAttr.SetKey(key)
	msgAttr.SetValue(val)
	require.Contains(t, m.GetAttributes(), msgAttr)
}

func TestContainer_Attribute(t *testing.T) {
	const attrKey1, attrKey2 = "key1", "key2"
	const attrVal1, attrVal2 = "val1", "val2"

	val := containertest.Container()

	val.SetAttribute(attrKey1, attrVal1)
	val.SetAttribute(attrKey2, attrVal2)

	var msg v2container.Container
	val.WriteToV2(&msg)

	require.GreaterOrEqual(t, len(msg.GetAttributes()), 2)
	assertContainsAttribute(t, msg, attrKey1, attrVal1)
	assertContainsAttribute(t, msg, attrKey2, attrVal2)

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, attrVal1, val2.Attribute(attrKey1))
	require.Equal(t, attrVal2, val2.Attribute(attrKey2))

	m := map[string]string{}

	val2.IterateAttributes(func(key, val string) {
		m[key] = val
	})

	require.GreaterOrEqual(t, len(m), 2)
	require.Equal(t, attrVal1, m[attrKey1])
	require.Equal(t, attrVal2, m[attrKey2])

	val2.SetAttribute(attrKey1, attrVal1+"_")
	require.Equal(t, attrVal1+"_", val2.Attribute(attrKey1))
}

func TestContainer_IterateUserAttributes(t *testing.T) {
	var cnr container.Container
	mSys := make(map[string]string)
	mUsr := make(map[string]string)

	for i := range 10 {
		si := strconv.Itoa(i)

		keyUsr := "key" + si
		valUsr := "val" + si
		keySys := "__NEOFS__" + si
		valSys := "sys-val" + si

		mUsr[keyUsr] = valUsr
		mSys[keySys] = valSys

		cnr.SetAttribute(keySys, valSys)
		cnr.SetAttribute(keyUsr, valUsr)
	}

	cnr.IterateUserAttributes(func(key, val string) {
		_, isSys := mSys[key]
		require.False(t, isSys, key)
		require.Equal(t, mUsr[key], val, key)
		delete(mUsr, key)
	})

	require.Empty(t, mUsr)
}

func TestSetName(t *testing.T) {
	var val container.Container

	require.Panics(t, func() {
		val.SetName("")
	})

	val = containertest.Container()

	const name = "some name"

	val.SetName(name)

	var msg v2container.Container
	val.WriteToV2(&msg)

	assertContainsAttribute(t, msg, "Name", name)

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, name, val2.Name())
}

func TestSetCreationTime(t *testing.T) {
	var val container.Container

	require.Zero(t, val.CreatedAt().Unix())

	val = containertest.Container()

	creat := time.Now()

	val.SetCreationTime(creat)

	var msg v2container.Container
	val.WriteToV2(&msg)

	assertContainsAttribute(t, msg, "Timestamp", strconv.FormatInt(creat.Unix(), 10))

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, creat.Unix(), val2.CreatedAt().Unix())
}

func TestDisableHomomorphicHashing(t *testing.T) {
	var val container.Container

	require.False(t, val.IsHomomorphicHashingDisabled())

	val = containertest.Container()

	val.DisableHomomorphicHashing()

	var msg v2container.Container
	val.WriteToV2(&msg)

	assertContainsAttribute(t, msg, v2container.SysAttributePrefix+"DISABLE_HOMOMORPHIC_HASHING", "true")

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.True(t, val2.IsHomomorphicHashingDisabled())
}

func TestWriteDomain(t *testing.T) {
	var val container.Container

	require.Zero(t, val.ReadDomain().Name())

	val = containertest.Container()

	const name = "domain name"

	var d container.Domain
	d.SetName(name)

	val.WriteDomain(d)

	var msg v2container.Container
	val.WriteToV2(&msg)

	assertContainsAttribute(t, msg, v2container.SysAttributeName, name)
	assertContainsAttribute(t, msg, v2container.SysAttributeZone, "container")

	const zone = "domain zone"

	d.SetZone(zone)

	val.WriteDomain(d)

	val.WriteToV2(&msg)

	assertContainsAttribute(t, msg, v2container.SysAttributeZone, zone)

	var val2 container.Container
	require.NoError(t, val2.ReadFromV2(msg))

	require.Equal(t, d, val2.ReadDomain())
}

func TestCalculateID(t *testing.T) {
	val := containertest.Container()

	require.False(t, val.AssertID(cidtest.ID()))

	var id cid.ID
	val.CalculateID(&id)

	var msg refs.ContainerID
	id.WriteToV2(&msg)

	h := sha256.Sum256(val.Marshal())
	require.Equal(t, h[:], msg.GetValue())

	var id2 cid.ID
	require.NoError(t, id2.ReadFromV2(msg))

	require.True(t, val.AssertID(id2))
}

func TestCalculateSignature(t *testing.T) {
	val := containertest.Container()

	var sig neofscrypto.Signature

	signer := neofscryptotest.Signer()
	require.Error(t, val.CalculateSignature(&sig, signer))
	require.NoError(t, val.CalculateSignature(&sig, signer.RFC6979))

	var msg refs.Signature
	sig.WriteToV2(&msg)

	var sig2 neofscrypto.Signature
	require.NoError(t, sig2.ReadFromV2(msg))

	require.True(t, val.VerifySignature(sig2))
}
