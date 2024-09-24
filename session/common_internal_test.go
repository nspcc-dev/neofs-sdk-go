package session

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func Test_commonData_copyTo(t *testing.T) {
	sig := neofscrypto.NewSignatureFromRawKey(neofscrypto.ECDSA_SHA512, []byte("key"), []byte("sign"))

	usr := usertest.User()

	data := commonData{
		idSet:   true,
		id:      uuid.New(),
		issuer:  usr.UserID(),
		iat:     1,
		nbf:     2,
		exp:     3,
		authKey: []byte{1, 2, 3, 4},
		sigSet:  true,
		sig:     sig,
	}

	t.Run("copy", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}

		require.Equal(t, data, dst)
		require.True(t, bytes.Equal(data.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.Equal(t, data.issuer, dst.issuer)
	})

	t.Run("change id", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		require.Equal(t, data.idSet, dst.idSet)
		require.Equal(t, data.id.String(), dst.id.String())

		dst.SetID(uuid.New())

		require.Equal(t, data.idSet, dst.idSet)
		require.NotEqual(t, data.id.String(), dst.id.String())
	})

	t.Run("overwrite id", func(t *testing.T) {
		// id is not set
		local := commonData{}
		require.False(t, local.idSet)

		// id is set
		var dst commonData
		dst.SetID(uuid.New())
		require.True(t, dst.idSet)

		// overwrite ID data
		local.copyTo(&dst)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.False(t, local.idSet)
		require.False(t, dst.idSet)

		// update id
		dst.SetID(uuid.New())

		// check that affects only dst
		require.False(t, local.idSet)
		require.True(t, dst.idSet)
	})

	t.Run("change issuer", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		require.True(t, data.issuer == dst.issuer)

		dst.SetIssuer(usertest.OtherID(usr.ID))

		require.False(t, data.issuer == dst.issuer)
	})

	t.Run("overwrite issuer", func(t *testing.T) {
		var local commonData
		require.Zero(t, local.issuer)

		var dst commonData
		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.NotZero(t, dst.issuer)

		local.copyTo(&dst)
		require.Zero(t, local.issuer)
		require.Zero(t, dst.issuer)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.True(t, local.issuer == dst.issuer)

		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.Zero(t, local.issuer)
		require.NotZero(t, dst.issuer)

		require.False(t, local.issuer == dst.issuer)
	})

	t.Run("change lifetime", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

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
		local := commonData{}

		// lifetime is set
		var dst commonData
		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		local.copyTo(&dst)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

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

	t.Run("change sig", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		require.Equal(t, data.sigSet, dst.sigSet)
		require.Equal(t, data.sig.Scheme(), dst.sig.Scheme())
		require.True(t, bytes.Equal(data.sig.PublicKeyBytes(), dst.sig.PublicKeyBytes()))
		require.True(t, bytes.Equal(data.sig.Value(), dst.sig.Value()))

		dst.sig.SetPublicKeyBytes([]byte{1, 2, 3})
		dst.sig.SetScheme(100)
		dst.sig.SetValue([]byte{10, 11, 12})

		require.Equal(t, data.issuer, dst.issuer)
		require.NotEqual(t, data.sig.Scheme(), dst.sig.Scheme())
		require.False(t, bytes.Equal(data.sig.PublicKeyBytes(), dst.sig.PublicKeyBytes()))
		require.False(t, bytes.Equal(data.sig.Value(), dst.sig.Value()))
	})

	t.Run("overwrite sig", func(t *testing.T) {
		var local Container
		require.False(t, local.sigSet)

		var dst Container
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
