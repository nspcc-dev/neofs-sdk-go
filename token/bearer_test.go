package token_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	tokentest "github.com/nspcc-dev/neofs-sdk-go/token/test"
	"github.com/stretchr/testify/require"
)

func TestBearerToken_Issuer(t *testing.T) {
	bearerToken := token.NewBearerToken()

	t.Run("non signed token", func(t *testing.T) {
		require.Nil(t, bearerToken.Issuer())
	})

	t.Run("signed token", func(t *testing.T) {
		p, err := keys.NewPrivateKey()
		require.NoError(t, err)

		ownerID := owner.NewIDFromPublicKey((*ecdsa.PublicKey)(p.PublicKey()))

		bearerToken.SetEACLTable(eacl.NewTable())
		require.NoError(t, bearerToken.SignToken(&p.PrivateKey))
		require.True(t, ownerID.Equal(bearerToken.Issuer()))
	})
}

func TestFilterEncoding(t *testing.T) {
	f := tokentest.BearerToken()

	t.Run("binary", func(t *testing.T) {
		data, err := f.Marshal()
		require.NoError(t, err)

		f2 := token.NewBearerToken()
		require.NoError(t, f2.Unmarshal(data))

		require.Equal(t, f, f2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := f.MarshalJSON()
		require.NoError(t, err)

		d2 := token.NewBearerToken()
		require.NoError(t, d2.UnmarshalJSON(data))

		require.Equal(t, f, d2)
	})
}

func TestBearerToken_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *token.BearerToken

		require.Nil(t, x.ToV2())
	})
}

func TestNewBearerToken(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		tkn := token.NewBearerToken()

		// convert to v2 message
		tknV2 := tkn.ToV2()

		require.NotNil(t, tknV2.GetBody())
		require.Zero(t, tknV2.GetBody().GetLifetime().GetExp())
		require.Zero(t, tknV2.GetBody().GetLifetime().GetNbf())
		require.Zero(t, tknV2.GetBody().GetLifetime().GetIat())
		require.Nil(t, tknV2.GetBody().GetEACL())
		require.Nil(t, tknV2.GetBody().GetOwnerID())
		require.Nil(t, tknV2.GetSignature())
	})
}
