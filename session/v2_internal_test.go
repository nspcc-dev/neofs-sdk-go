package session

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestV2_CopyTo(t *testing.T) {
	sig := neofscrypto.NewSignatureFromRawKey(neofscrypto.ECDSA_SHA512, []byte("key"), []byte("sign"))

	usr := usertest.User()

	data := TokenV2{
		Lifetime: Lifetime{
			iat: 1,
			nbf: 2,
			exp: 3,
		},
		version: 2,
		idSet:   true,
		id:      uuid.New(),
		issuer:  Target{ownerID: usr.UserID()},
		subjects: []Target{
			{nnsName: "alice.neofs"},
		},
		contexts: []ContextV2{
			{
				container: cidtest.ID(),
				verbs:     []VerbV2{VerbV2ObjectGet, VerbV2ObjectPut},
				objects:   oidtest.IDs(3),
			},
		},
		delegationChain: []DelegationInfo{
			{
				Lifetime: Lifetime{
					iat: 1,
					nbf: 2,
					exp: 3,
				},
				issuer: Target{ownerID: usr.UserID()},
				subjects: []Target{
					{nnsName: "bob.neofs"},
				},
				verbs:  []VerbV2{VerbV2ObjectGet, VerbV2ObjectPut},
				sigSet: true,
				sig:    sig,
			},
		},
		sigSet: true,
		sig:    sig,
	}

	t.Run("copy", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data, dst)
		require.Equal(t, data.Marshal(), dst.Marshal())
		require.Equal(t, data.issuer, dst.issuer)
	})

	t.Run("change version", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.version, dst.version)

		dst.version = 100

		require.NotEqual(t, data.version, dst.version)
	})

	t.Run("overwrite version", func(t *testing.T) {
		var local TokenV2
		require.Zero(t, local.version)

		var dst TokenV2
		dst.version = 100
		require.NotZero(t, dst.version)

		local.CopyTo(&dst)
		require.Zero(t, local.version)
		require.Zero(t, dst.version)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.version = 100
		require.Zero(t, local.version)
		require.NotZero(t, dst.version)

		require.NotEqual(t, local.version, dst.version)
	})

	t.Run("change id", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.idSet, dst.idSet)
		require.Equal(t, data.id.String(), dst.id.String())

		dst.SetID(uuid.New())

		require.Equal(t, data.idSet, dst.idSet)
		require.NotEqual(t, data.id.String(), dst.id.String())
	})

	t.Run("overwrite id", func(t *testing.T) {
		// id is not set
		local := TokenV2{}
		require.False(t, local.idSet)

		// id is set
		var dst TokenV2
		dst.SetID(uuid.New())
		require.True(t, dst.idSet)

		// overwrite ID data
		local.CopyTo(&dst)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.False(t, local.idSet)
		require.False(t, dst.idSet)

		// update id
		dst.SetID(uuid.New())

		// check that affects only dst
		require.False(t, local.idSet)
		require.True(t, dst.idSet)
	})

	t.Run("change issuer", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.issuer, dst.issuer)

		dst.SetIssuer(Target{ownerID: usertest.OtherID(usr.ID)})

		require.NotEqual(t, data.issuer, dst.issuer)
	})

	t.Run("overwrite issuer", func(t *testing.T) {
		var local TokenV2
		require.Zero(t, local.issuer)

		var dst TokenV2
		dst.SetIssuer(Target{ownerID: usertest.OtherID(usr.ID)})
		require.NotZero(t, dst.issuer)

		local.CopyTo(&dst)
		require.Zero(t, local.issuer)
		require.Zero(t, dst.issuer)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.Equal(t, local.issuer, dst.issuer)

		dst.SetIssuer(Target{ownerID: usertest.OtherID(usr.ID)})
		require.Zero(t, local.issuer)
		require.NotZero(t, dst.issuer)

		require.NotEqual(t, local.issuer, dst.issuer)
	})

	t.Run("change subjects", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.subjects, dst.subjects)

		// Modify the copy's subjects
		dst.subjects = append(dst.subjects, Target{nnsName: "charlie.neofs"})

		require.NotEqual(t, len(data.subjects), len(dst.subjects))
	})

	t.Run("overwrite subjects", func(t *testing.T) {
		var local TokenV2
		require.Nil(t, local.subjects)

		var dst TokenV2
		dst.subjects = []Target{{nnsName: "test.neofs"}}
		require.NotNil(t, dst.subjects)

		local.CopyTo(&dst)
		require.Nil(t, local.subjects)
		require.Nil(t, dst.subjects)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.subjects = []Target{{nnsName: "test.neofs"}}
		require.Nil(t, local.subjects)
		require.NotNil(t, dst.subjects)
	})

	t.Run("change lifetime", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.iat, dst.iat)
		require.Equal(t, data.nbf, dst.nbf)
		require.Equal(t, data.exp, dst.exp)

		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		require.NotEqual(t, data.iat, dst.iat)
		require.NotEqual(t, data.nbf, dst.nbf)
		require.NotEqual(t, data.exp, dst.exp)
	})

	t.Run("overwrite lifetime", func(t *testing.T) {
		// lifetime is not set
		local := TokenV2{}

		// lifetime is set
		var dst TokenV2
		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		local.CopyTo(&dst)

		require.Equal(t, local.Marshal(), dst.Marshal())

		// check both are equal
		require.Equal(t, local.iat, dst.iat)
		require.Equal(t, local.nbf, dst.nbf)
		require.Equal(t, local.exp, dst.exp)

		// update lifetime
		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		// check that affects only dst
		require.NotEqual(t, local.iat, dst.iat)
		require.NotEqual(t, local.nbf, dst.nbf)
		require.NotEqual(t, local.exp, dst.exp)
	})

	t.Run("change contexts", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, len(data.contexts), len(dst.contexts))

		dst.contexts = append(dst.contexts, ContextV2{
			container: cidtest.ID(),
			verbs:     []VerbV2{VerbV2ObjectDelete},
		})

		require.NotEqual(t, len(data.contexts), len(dst.contexts))
	})

	t.Run("overwrite contexts", func(t *testing.T) {
		var local TokenV2
		require.Nil(t, local.contexts)

		var dst TokenV2
		dst.contexts = []ContextV2{{container: cidtest.ID()}}
		require.NotNil(t, dst.contexts)

		local.CopyTo(&dst)
		require.Nil(t, local.contexts)
		require.Nil(t, dst.contexts)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.contexts = []ContextV2{{container: cidtest.ID()}}
		require.Nil(t, local.contexts)
		require.NotNil(t, dst.contexts)
	})

	t.Run("change delegationChain", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, len(data.delegationChain), len(dst.delegationChain))

		dst.delegationChain = append(dst.delegationChain, DelegationInfo{
			issuer:   Target{ownerID: usertest.OtherID(usr.ID)},
			subjects: []Target{{nnsName: "charlie.neofs"}},
			verbs:    []VerbV2{VerbV2ObjectDelete},
		})

		require.NotEqual(t, len(data.delegationChain), len(dst.delegationChain))
	})

	t.Run("overwrite delegationChain", func(t *testing.T) {
		var local TokenV2
		require.Nil(t, local.delegationChain)

		var dst TokenV2
		dst.delegationChain = []DelegationInfo{{
			issuer:   Target{ownerID: usr.UserID()},
			subjects: []Target{{nnsName: "test.neofs"}},
		}}
		require.NotNil(t, dst.delegationChain)

		local.CopyTo(&dst)
		require.Nil(t, local.delegationChain)
		require.Nil(t, dst.delegationChain)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.delegationChain = []DelegationInfo{{
			issuer:   Target{ownerID: usr.UserID()},
			subjects: []Target{{nnsName: "test.neofs"}},
		}}
		require.Nil(t, local.delegationChain)
		require.NotNil(t, dst.delegationChain)
	})

	t.Run("change sig", func(t *testing.T) {
		var dst TokenV2
		data.CopyTo(&dst)

		require.Equal(t, data.sig, dst.sig)

		dst.sig.SetPublicKeyBytes([]byte{1, 2, 3})
		dst.sig.SetScheme(100)
		dst.sig.SetValue([]byte{10, 11, 12})

		require.Equal(t, data.issuer, dst.issuer)
		require.NotEqual(t, data.sig.Scheme(), dst.sig.Scheme())
		require.False(t, bytes.Equal(data.sig.PublicKeyBytes(), dst.sig.PublicKeyBytes()))
		require.False(t, bytes.Equal(data.sig.Value(), dst.sig.Value()))
	})

	t.Run("overwrite sig", func(t *testing.T) {
		var local TokenV2
		require.False(t, local.sigSet)

		var dst TokenV2
		require.NoError(t, dst.Sign(usr))
		require.True(t, dst.sigSet)

		local.CopyTo(&dst)
		require.False(t, local.sigSet)
		require.False(t, dst.sigSet)

		require.True(t, bytes.Equal(local.Marshal(), dst.Marshal()))

		require.NoError(t, dst.Sign(usr))
		require.False(t, local.sigSet)
		require.True(t, dst.sigSet)
	})
}

func TestTarget_protoMessage(t *testing.T) {
	t.Run("zero target returns nil", func(t *testing.T) {
		var target Target
		require.Nil(t, target.protoMessage())
	})

	t.Run("ownerID target", func(t *testing.T) {
		userID := usertest.ID()
		target := NewTarget(userID)
		msg := target.protoMessage()
		require.NotNil(t, msg)
		require.NotNil(t, msg.GetOwnerId())
	})

	t.Run("NNS target", func(t *testing.T) {
		nnsName := "test.neo"
		target := NewTargetFromNNS(nnsName)
		msg := target.protoMessage()
		require.NotNil(t, msg)
		require.Equal(t, nnsName, msg.GetNnsName())
	})
}
