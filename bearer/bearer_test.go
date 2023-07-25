package bearer_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// compares binary representations of two eacl.Table instances.
func isEqualEACLTables(t1, t2 eacl.Table) bool {
	d1, err := t1.Marshal()
	if err != nil {
		panic(err)
	}

	d2, err := t2.Marshal()
	if err != nil {
		panic(err)
	}

	return bytes.Equal(d1, d2)
}

func TestToken_SetEACLTable(t *testing.T) {
	var val bearer.Token
	var m acl.BearerToken
	filled := bearertest.Token(t)

	val.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	val2 := filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))
	require.Zero(t, val2.EACLTable())

	val2 = filled

	jd, err := val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))
	require.Zero(t, val2.EACLTable())

	// set value

	eaclTable := *eacltest.Table(t)

	val.SetEACLTable(eaclTable)
	require.True(t, isEqualEACLTables(eaclTable, val.EACLTable()))

	val.WriteToV2(&m)
	eaclTableV2 := eaclTable.ToV2()
	require.Equal(t, eaclTableV2, m.GetBody().GetEACL())

	val2 = filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))
	require.True(t, isEqualEACLTables(eaclTable, val.EACLTable()))

	val2 = filled

	jd, err = val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))
	require.True(t, isEqualEACLTables(eaclTable, val.EACLTable()))
}

func TestToken_ForUser(t *testing.T) {
	var val bearer.Token
	var m acl.BearerToken
	filled := bearertest.Token(t)

	val.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	val2 := filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))

	val2.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	val2 = filled

	jd, err := val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))

	val2.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	// set value
	usr := *usertest.ID(t)

	var usrV2 refs.OwnerID
	usr.WriteToV2(&usrV2)

	val.ForUser(usr)

	val.WriteToV2(&m)
	require.Equal(t, usrV2, *m.GetBody().GetOwnerID())

	val2 = filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))

	val2.WriteToV2(&m)
	require.Equal(t, usrV2, *m.GetBody().GetOwnerID())

	val2 = filled

	jd, err = val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))

	val2.WriteToV2(&m)
	require.Equal(t, usrV2, *m.GetBody().GetOwnerID())
}

func testLifetimeClaim(t *testing.T, setter func(*bearer.Token, uint64), getter func(*acl.BearerToken) uint64) {
	var val bearer.Token
	var m acl.BearerToken
	filled := bearertest.Token(t)

	val.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	val2 := filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))

	val2.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	val2 = filled

	jd, err := val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))

	val2.WriteToV2(&m)
	require.Zero(t, m.GetBody())

	// set value
	exp := rand.Uint64()

	setter(&val, exp)

	val.WriteToV2(&m)
	require.Equal(t, exp, getter(&m))

	val2 = filled

	require.NoError(t, val2.Unmarshal(val.Marshal()))

	val2.WriteToV2(&m)
	require.Equal(t, exp, getter(&m))

	val2 = filled

	jd, err = val.MarshalJSON()
	require.NoError(t, err)

	require.NoError(t, val2.UnmarshalJSON(jd))

	val2.WriteToV2(&m)
	require.Equal(t, exp, getter(&m))
}

func TestToken_SetLifetime(t *testing.T) {
	t.Run("iat", func(t *testing.T) {
		testLifetimeClaim(t, (*bearer.Token).SetIat, func(token *acl.BearerToken) uint64 {
			return token.GetBody().GetLifetime().GetIat()
		})
	})

	t.Run("nbf", func(t *testing.T) {
		testLifetimeClaim(t, (*bearer.Token).SetNbf, func(token *acl.BearerToken) uint64 {
			return token.GetBody().GetLifetime().GetNbf()
		})
	})

	t.Run("exp", func(t *testing.T) {
		testLifetimeClaim(t, (*bearer.Token).SetExp, func(token *acl.BearerToken) uint64 {
			return token.GetBody().GetLifetime().GetExp()
		})
	})
}

func TestToken_InvalidAt(t *testing.T) {
	var val bearer.Token

	require.True(t, val.InvalidAt(0))
	require.True(t, val.InvalidAt(1))

	val.SetIat(1)
	val.SetNbf(2)
	val.SetExp(4)

	require.True(t, val.InvalidAt(0))
	require.True(t, val.InvalidAt(1))
	require.False(t, val.InvalidAt(2))
	require.False(t, val.InvalidAt(3))
	require.False(t, val.InvalidAt(4))
	require.True(t, val.InvalidAt(5))
}

func TestToken_AssertContainer(t *testing.T) {
	var val bearer.Token
	cnr := cidtest.ID()

	require.True(t, val.AssertContainer(cnr))

	eaclTable := *eacltest.Table(t)

	eaclTable.SetCID(cidtest.ID())
	val.SetEACLTable(eaclTable)
	require.False(t, val.AssertContainer(cnr))

	eaclTable.SetCID(cnr)
	val.SetEACLTable(eaclTable)
	require.True(t, val.AssertContainer(cnr))
}

func TestToken_AssertUser(t *testing.T) {
	var val bearer.Token
	usr := *usertest.ID(t)

	require.True(t, val.AssertUser(usr))

	val.ForUser(*usertest.ID(t))
	require.False(t, val.AssertUser(usr))

	val.ForUser(usr)
	require.True(t, val.AssertUser(usr))
}

func TestToken_Sign(t *testing.T) {
	var val bearer.Token

	require.False(t, val.VerifySignature())

	signer := test.RandomSigner(t)

	val = bearertest.Token(t)

	require.NoError(t, val.Sign(signer))

	require.True(t, val.VerifySignature())

	var m acl.BearerToken
	val.WriteToV2(&m)

	require.NotZero(t, m.GetSignature().GetKey())
	require.NotZero(t, m.GetSignature().GetSign())

	val2 := bearertest.Token(t)

	require.NoError(t, val2.Unmarshal(val.Marshal()))
	require.True(t, val2.VerifySignature())

	jd, err := val.MarshalJSON()
	require.NoError(t, err)

	val2 = bearertest.Token(t)
	require.NoError(t, val2.UnmarshalJSON(jd))
	require.True(t, val2.VerifySignature())
}

func TestToken_ReadFromV2(t *testing.T) {
	var val bearer.Token
	var m acl.BearerToken

	require.Error(t, val.ReadFromV2(m))

	var body acl.BearerTokenBody
	m.SetBody(&body)

	require.Error(t, val.ReadFromV2(m))

	eaclTable := eacltest.Table(t).ToV2()
	body.SetEACL(eaclTable)

	require.Error(t, val.ReadFromV2(m))

	var lifetime acl.TokenLifetime
	body.SetLifetime(&lifetime)

	require.Error(t, val.ReadFromV2(m))

	const iat, nbf, exp = 1, 2, 3
	lifetime.SetIat(iat)
	lifetime.SetNbf(nbf)
	lifetime.SetExp(exp)

	body.SetLifetime(&lifetime)

	require.Error(t, val.ReadFromV2(m))

	var sig refs.Signature
	m.SetSignature(&sig)

	require.NoError(t, val.ReadFromV2(m))

	var m2 acl.BearerToken

	val.WriteToV2(&m2)
	require.Equal(t, m, m2)

	usr, usr2 := *usertest.ID(t), *usertest.ID(t)

	require.True(t, val.AssertUser(usr))
	require.True(t, val.AssertUser(usr2))

	var usrV2 refs.OwnerID
	usr.WriteToV2(&usrV2)

	body.SetOwnerID(&usrV2)

	require.NoError(t, val.ReadFromV2(m))

	val.WriteToV2(&m2)
	require.Equal(t, m, m2)

	require.True(t, val.AssertUser(usr))
	require.False(t, val.AssertUser(usr2))

	signer := test.RandomSigner(t)

	var s neofscrypto.Signature

	require.NoError(t, s.CalculateMarshalled(signer, &body))

	s.WriteToV2(&sig)

	require.NoError(t, val.ReadFromV2(m))
	require.True(t, val.VerifySignature())
	require.Equal(t, sig.GetKey(), val.SigningKeyBytes())
}

func TestResolveIssuer(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)

	var val bearer.Token

	require.Zero(t, bearer.ResolveIssuer(val))

	var m acl.BearerToken

	var sig refs.Signature
	sig.SetKey([]byte("invalid key"))

	m.SetSignature(&sig)

	require.NoError(t, val.Unmarshal(m.StableMarshal(nil)))

	require.Zero(t, bearer.ResolveIssuer(val))

	require.NoError(t, val.Sign(signer))

	usr := signer.UserID()

	require.Equal(t, usr, bearer.ResolveIssuer(val))
}
