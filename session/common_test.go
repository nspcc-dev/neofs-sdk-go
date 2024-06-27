package session

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func Test_commonData_copyTo(t *testing.T) {
	var sig refs.Signature

	sig.SetKey([]byte("key"))
	sig.SetSign([]byte("sign"))
	sig.SetScheme(refs.ECDSA_SHA512)

	usr := usertest.User()

	data := commonData{
		idSet:       true,
		id:          uuid.New(),
		issuerSet:   true,
		issuer:      usr.UserID(),
		lifetimeSet: true,
		iat:         1,
		nbf:         2,
		exp:         3,
		authKey:     []byte{1, 2, 3, 4},
		sigSet:      true,
		sig:         sig,
	}

	t.Run("copy", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}

		require.Equal(t, data, dst)
		require.True(t, bytes.Equal(data.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.Equal(t, data.issuerSet, dst.issuerSet)
		require.Equal(t, data.issuer.String(), dst.issuer.String())
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

		require.Equal(t, data.issuerSet, dst.issuerSet)
		require.True(t, data.issuer.Equals(dst.issuer))

		dst.SetIssuer(usertest.OtherID(usr.ID))

		require.Equal(t, data.issuerSet, dst.issuerSet)
		require.False(t, data.issuer.Equals(dst.issuer))
	})

	t.Run("overwrite issuer", func(t *testing.T) {
		var local commonData
		require.False(t, local.issuerSet)

		var dst commonData
		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.True(t, dst.issuerSet)

		local.copyTo(&dst)
		require.False(t, local.issuerSet)
		require.False(t, dst.issuerSet)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.Equal(t, local.issuerSet, dst.issuerSet)
		require.True(t, local.issuer.Equals(dst.issuer))

		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.False(t, local.issuerSet)
		require.True(t, dst.issuerSet)

		require.False(t, local.issuer.Equals(dst.issuer))
	})

	t.Run("change lifetime", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		require.Equal(t, data.lifetimeSet, dst.lifetimeSet)
		require.Equal(t, data.iat, dst.iat)
		require.Equal(t, data.nbf, dst.nbf)
		require.Equal(t, data.exp, dst.exp)

		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		require.Equal(t, data.lifetimeSet, dst.lifetimeSet)
		require.NotEqual(t, data.iat, dst.iat)
		require.NotEqual(t, data.nbf, dst.nbf)
		require.NotEqual(t, data.exp, dst.exp)
	})

	t.Run("overwrite lifetime", func(t *testing.T) {
		// lifetime is not set
		local := commonData{}
		require.False(t, local.lifetimeSet)

		// lifetime is set
		var dst commonData
		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)
		require.True(t, dst.lifetimeSet)

		local.copyTo(&dst)
		require.False(t, local.lifetimeSet)
		require.False(t, dst.lifetimeSet)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		// check both are equal
		require.Equal(t, local.lifetimeSet, dst.lifetimeSet)
		require.Equal(t, local.iat, dst.iat)
		require.Equal(t, local.nbf, dst.nbf)
		require.Equal(t, local.exp, dst.exp)

		// update lifetime
		dst.SetExp(100)
		dst.SetIat(200)
		dst.SetNbf(300)

		// check that affects only dst
		require.False(t, local.lifetimeSet)
		require.True(t, dst.lifetimeSet)
		require.NotEqual(t, local.iat, dst.iat)
		require.NotEqual(t, local.nbf, dst.nbf)
		require.NotEqual(t, local.exp, dst.exp)
	})

	t.Run("change sig", func(t *testing.T) {
		var dst commonData
		data.copyTo(&dst)

		require.Equal(t, data.sigSet, dst.sigSet)
		require.Equal(t, data.sig.GetScheme(), dst.sig.GetScheme())
		require.True(t, bytes.Equal(data.sig.GetKey(), dst.sig.GetKey()))
		require.True(t, bytes.Equal(data.sig.GetSign(), dst.sig.GetSign()))

		dst.sig.SetKey([]byte{1, 2, 3})
		dst.sig.SetScheme(100)
		dst.sig.SetSign([]byte{10, 11, 12})

		require.Equal(t, data.issuerSet, dst.issuerSet)
		require.NotEqual(t, data.sig.GetScheme(), dst.sig.GetScheme())
		require.False(t, bytes.Equal(data.sig.GetKey(), dst.sig.GetKey()))
		require.False(t, bytes.Equal(data.sig.GetSign(), dst.sig.GetSign()))
	})

	t.Run("overwrite sig", func(t *testing.T) {
		local := commonData{}
		require.False(t, local.sigSet)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}

		var dst commonData
		require.NoError(t, dst.sign(usr, emptyWriter))
		require.True(t, dst.sigSet)

		local.copyTo(&dst)
		require.False(t, local.sigSet)
		require.False(t, dst.sigSet)

		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.NoError(t, dst.sign(usr, emptyWriter))
		require.False(t, local.sigSet)
		require.True(t, dst.sigSet)
	})
}
