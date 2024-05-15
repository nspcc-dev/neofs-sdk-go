package bearer_test

import (
	"testing"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestToken_SetEACLTable(t *testing.T) {
	var tok bearer.Token

	_, ok := tok.EACLTable()
	require.False(t, ok)

	tbl := eacltest.Table()
	tblOther := eacltest.Table()

	tok.SetEACLTable(tbl)
	res, ok := tok.EACLTable()
	require.True(t, ok)
	require.Equal(t, tbl, res)

	tok.SetEACLTable(tblOther)
	res, ok = tok.EACLTable()
	require.True(t, ok)
	require.Equal(t, tblOther, res)

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token

			dst.SetEACLTable(tbl)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			_, ok = dst.EACLTable()
			require.False(t, ok)

			dst.SetEACLTable(tblOther)
			src.SetEACLTable(tbl)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			res, ok = dst.EACLTable()
			require.True(t, ok)
			require.Equal(t, tbl, res)
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.SetEACLTable(tbl)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			res, ok = dst.EACLTable()
			require.True(t, ok)
			require.Equal(t, tbl, res)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token

			dst.SetEACLTable(tbl)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			_, ok = dst.EACLTable()
			require.False(t, ok)

			dst.SetEACLTable(tblOther)
			src.SetEACLTable(tbl)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			res, ok = dst.EACLTable()
			require.True(t, ok)
			require.Equal(t, tbl, res)
		})
	})
}

func TestToken_ForUser(t *testing.T) {
	var tok bearer.Token

	usr := usertest.ID()
	otherUsr := usertest.ChangeID(usr)

	require.True(t, tok.AssertUser(usr))
	require.True(t, tok.AssertUser(otherUsr))

	tok.ForUser(usr)
	require.True(t, tok.AssertUser(usr))
	require.False(t, tok.AssertUser(otherUsr))

	tok.ForUser(otherUsr)
	require.False(t, tok.AssertUser(usr))
	require.True(t, tok.AssertUser(otherUsr))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.ForUser(usr)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AssertUser(otherUsr))
			dst.WriteToV2(&msg)
			require.Nil(t, msg.GetBody().GetOwnerId())

			dst.ForUser(otherUsr)
			src.ForUser(usr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AssertUser(usr))
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.ForUser(usr)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.True(t, dst.AssertUser(usr))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.ForUser(usr)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AssertUser(otherUsr))
			dst.WriteToV2(&msg)
			require.Nil(t, msg.GetBody().GetOwnerId())

			dst.ForUser(otherUsr)
			src.ForUser(usr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AssertUser(usr))
		})
	})
}

func TestToken_SetExp(t *testing.T) {
	var tok bearer.Token

	require.True(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.True(t, tok.InvalidAt(15))

	tok.SetExp(13)
	require.False(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.True(t, tok.InvalidAt(15))

	tok.SetExp(14)
	require.False(t, tok.InvalidAt(13))
	require.False(t, tok.InvalidAt(14))
	require.True(t, tok.InvalidAt(15))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(13)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetExp())

			dst.SetExp(42)
			src.SetExp(13)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.InvalidAt(13))
			require.True(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 13, msg.Body.Lifetime.Exp)
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			src.SetNbf(0)
			src.SetIat(0)
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.SetExp(13)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.False(t, dst.InvalidAt(13))
			require.True(t, dst.InvalidAt(14))
			require.EqualValues(t, 13, msg.Body.Lifetime.Exp)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(13)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetExp())

			dst.SetExp(42)
			src.SetExp(13)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.InvalidAt(13))
			require.True(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 13, msg.Body.Lifetime.Exp)
		})
	})
}

func TestToken_SetIat(t *testing.T) {
	var tok bearer.Token

	require.True(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.True(t, tok.InvalidAt(15))

	tok.SetExp(15)
	tok.SetIat(14)
	require.True(t, tok.InvalidAt(13))
	require.False(t, tok.InvalidAt(14))
	require.False(t, tok.InvalidAt(15))

	tok.SetIat(15)
	require.True(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.False(t, tok.InvalidAt(15))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(15)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetIat())

			dst.SetIat(42)
			src.SetExp(15)
			src.SetIat(14)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 14, msg.Body.Lifetime.Iat)
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			src.SetNbf(0)
			src.SetExp(15)
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.SetIat(14)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			require.EqualValues(t, 14, msg.Body.Lifetime.Iat)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(15)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetIat())

			dst.SetIat(42)
			src.SetExp(15)
			src.SetIat(14)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 14, msg.Body.Lifetime.Iat)
		})
	})
}

func TestToken_SetNbf(t *testing.T) {
	var tok bearer.Token

	require.True(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.True(t, tok.InvalidAt(15))

	tok.SetExp(15)
	tok.SetNbf(14)
	require.True(t, tok.InvalidAt(13))
	require.False(t, tok.InvalidAt(14))
	require.False(t, tok.InvalidAt(15))

	tok.SetNbf(15)
	require.True(t, tok.InvalidAt(13))
	require.True(t, tok.InvalidAt(14))
	require.False(t, tok.InvalidAt(15))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(15)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetNbf())

			dst.SetNbf(42)
			src.SetExp(15)
			src.SetNbf(14)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 14, msg.Body.Lifetime.Nbf)
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			src.SetIat(0)
			src.SetExp(15)
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.SetNbf(14)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			require.EqualValues(t, 14, msg.Body.Lifetime.Nbf)
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token
			var msg apiacl.BearerToken

			dst.SetExp(15)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(0))
			dst.WriteToV2(&msg)
			require.Zero(t, msg.GetBody().GetLifetime().GetNbf())

			dst.SetNbf(42)
			src.SetExp(15)
			src.SetNbf(14)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.InvalidAt(13))
			require.False(t, dst.InvalidAt(14))
			dst.WriteToV2(&msg)
			require.EqualValues(t, 14, msg.Body.Lifetime.Nbf)
		})
	})
}

func TestToken_Issuer(t *testing.T) {
	var tok bearer.Token

	require.Zero(t, tok.Issuer())

	usr := usertest.ID()
	tok.SetIssuer(usr)
	require.Equal(t, usr, tok.Issuer())

	otherUsr := usertest.ChangeID(usr)
	tok.SetIssuer(otherUsr)
	require.Equal(t, otherUsr, tok.Issuer())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst bearer.Token

			dst.SetIssuer(usr)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.Issuer())

			dst.SetIssuer(otherUsr)
			src.SetIssuer(usr)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, usr, dst.Issuer())
		})
		t.Run("api", func(t *testing.T) {
			src := bearertest.Token()
			var dst bearer.Token
			var msg apiacl.BearerToken

			src.SetIssuer(usr)
			src.WriteToV2(&msg)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, usr, dst.Issuer())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst bearer.Token

			dst.SetIssuer(usr)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Issuer())

			dst.SetIssuer(otherUsr)
			src.SetIssuer(usr)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, usr, dst.Issuer())
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

func TestToken_Sign(t *testing.T) {
	var val bearer.Token

	require.False(t, val.VerifySignature())

	usr, _ := usertest.TwoUsers()

	val = bearertest.Token()

	require.NoError(t, val.Sign(usr))

	require.True(t, val.VerifySignature())

	var m apiacl.BearerToken
	val.WriteToV2(&m)

	require.NotZero(t, m.GetSignature().GetKey())
	require.NotZero(t, m.GetSignature().GetSign())

	val2 := bearertest.Token()

	require.NoError(t, val2.Unmarshal(val.Marshal()))
	require.True(t, val2.VerifySignature())

	jd, err := val.MarshalJSON()
	require.NoError(t, err)

	val2 = bearertest.Token()
	require.NoError(t, val2.UnmarshalJSON(jd))
	require.True(t, val2.VerifySignature())
}

func TestToken_SignedData(t *testing.T) {
	var val bearer.Token

	require.False(t, val.VerifySignature())

	signedData := val.SignedData()
	var dec bearer.Token
	require.NoError(t, dec.UnmarshalSignedData(signedData))
	require.Equal(t, val, dec)

	signer := test.RandomSignerRFC6979(t)
	val = bearertest.Token()
	val.SetIssuer(signer.UserID())

	test.SignedDataComponentUser(t, signer, &val)
}

func TestToken_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("signature", func(t *testing.T) {
			tok := bearertest.Token()
			var m apiacl.BearerToken
			tok.WriteToV2(&m)
			require.ErrorContains(t, tok.ReadFromV2(&m), "missing body signature")
		})
		t.Run("body", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)
			m.Body = nil
			require.ErrorContains(t, tok.ReadFromV2(&m), "missing token body")
		})
		t.Run("eACL", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)
			m.Body.EaclTable = nil
			require.ErrorContains(t, tok.ReadFromV2(&m), "missing eACL table")
		})
		t.Run("lifetime", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)
			m.Body.Lifetime = nil
			require.ErrorContains(t, tok.ReadFromV2(&m), "missing token lifetime")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("target user", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.OwnerId.Value = []byte("not_a_user")
			require.ErrorContains(t, tok.ReadFromV2(&m), "invalid target user")
		})
		t.Run("issuer", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.Issuer.Value = []byte("not_a_user")
			require.ErrorContains(t, tok.ReadFromV2(&m), "invalid issuer")
		})
		t.Run("signature", func(t *testing.T) {
			t.Run("public key", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Key = nil
				require.ErrorContains(t, tok.ReadFromV2(&m), "invalid body signature: missing public key")
				m.Signature.Key = []byte("not_a_key")
				require.ErrorContains(t, tok.ReadFromV2(&m), "invalid body signature: decode public key from binary")
			})
			t.Run("value", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Sign = nil
				require.ErrorContains(t, tok.ReadFromV2(&m), "invalid body signature: missing signature")
			})
		})
		t.Run("eACL", func(t *testing.T) {
			t.Run("records", func(t *testing.T) {
				t.Run("targets", func(t *testing.T) {
					rs := eacltest.NRecords(2)
					rs[1].SetTargets(eacltest.NTargets(3))
					tbl := eacltest.Table()
					tbl.SetRecords(rs)
					tok := bearertest.SignedToken(t)
					tok.SetEACLTable(tbl)
					var m apiacl.BearerToken
					tok.WriteToV2(&m)

					m.Body.EaclTable.Records[1].Targets[2].Role, m.Body.EaclTable.Records[1].Targets[2].Keys = 0, nil
					err := tok.ReadFromV2(&m)
					require.ErrorContains(t, err, "invalid eACL table: invalid record #1: invalid target #2: role and public keys are not mutually exclusive")
					m.Body.EaclTable.Records[1].Targets[2].Role, m.Body.EaclTable.Records[1].Targets[2].Keys = 1, make([][]byte, 1)
					err = tok.ReadFromV2(&m)
					require.ErrorContains(t, err, "invalid eACL table: invalid record #1: invalid target #2: role and public keys are not mutually exclusive")
					m.Body.EaclTable.Records[1].Targets = nil
					err = tok.ReadFromV2(&m)
					require.ErrorContains(t, err, "invalid eACL table: invalid record #1: missing target subjects")
				})
				t.Run("filters", func(t *testing.T) {
					rs := eacltest.NRecords(2)
					rs[1].SetFilters(eacltest.NFilters(3))
					tbl := eacltest.Table()
					tbl.SetRecords(rs)
					tok := bearertest.SignedToken(t)
					tok.SetEACLTable(tbl)
					var m apiacl.BearerToken
					tok.WriteToV2(&m)

					m.Body.EaclTable.Records[1].Filters[2].Key = ""
					err := tok.ReadFromV2(&m)
					require.ErrorContains(t, err, "invalid eACL table: invalid record #1: invalid filter #2: missing key")
				})
			})
		})
	})
}

func TestResolveIssuer(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)

	var val bearer.Token

	require.Zero(t, val.ResolveIssuer())

	require.NoError(t, val.Sign(signer))

	usr := signer.UserID()

	require.Equal(t, usr, val.ResolveIssuer())
	require.Equal(t, usr, val.Issuer())
}

func TestToken_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var tok bearer.Token
		msg := []byte("definitely_not_protobuf")
		err := tok.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("target user", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.OwnerId.Value = []byte("not_a_user")
			b, err := proto.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tok.Unmarshal(b), "invalid target user")
		})
		t.Run("issuer", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.Issuer.Value = []byte("not_a_user")
			b, err := proto.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tok.Unmarshal(b), "invalid issuer")
		})
		t.Run("signature", func(t *testing.T) {
			t.Run("public key", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Key = nil
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.Unmarshal(b), "invalid body signature: missing public key")

				m.Signature.Key = []byte("not_a_key")
				b, err = proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.Unmarshal(b), "invalid body signature: decode public key from binary")
			})
			t.Run("value", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Sign = nil
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.Unmarshal(b), "invalid body signature: missing signature")
			})
		})
		t.Run("eACL", func(t *testing.T) {
			t.Run("container", func(t *testing.T) {
				tbl := eacltest.Table()
				tok := bearertest.SignedToken(t)
				tok.SetEACLTable(tbl)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Body.EaclTable.ContainerId.Value = []byte("not_a_container_ID")
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.Unmarshal(b), "invalid eACL table: invalid container")
			})
		})
	})
}

func TestToken_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var tok bearer.Token
		msg := []byte("definitely_not_protojson")
		err := tok.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("target user", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.OwnerId.Value = []byte("not_a_user")
			b, err := protojson.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid target user")
		})
		t.Run("issuer", func(t *testing.T) {
			tok := bearertest.SignedToken(t)
			var m apiacl.BearerToken
			tok.WriteToV2(&m)

			m.Body.Issuer.Value = []byte("not_a_user")
			b, err := protojson.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid issuer")
		})
		t.Run("signature", func(t *testing.T) {
			t.Run("public key", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Key = nil
				b, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid body signature: missing public key")

				m.Signature.Key = []byte("not_a_key")
				b, err = protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid body signature: decode public key from binary")
			})
			t.Run("value", func(t *testing.T) {
				tok := bearertest.SignedToken(t)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Signature.Sign = nil
				b, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid body signature: missing signature")
			})
		})
		t.Run("eACL", func(t *testing.T) {
			t.Run("container", func(t *testing.T) {
				tbl := eacltest.Table()
				tok := bearertest.SignedToken(t)
				tok.SetEACLTable(tbl)
				var m apiacl.BearerToken
				tok.WriteToV2(&m)

				m.Body.EaclTable.ContainerId.Value = []byte("not_a_container_ID")
				b, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, tok.UnmarshalJSON(b), "invalid eACL table: invalid container")
			})
		})
	})
}
