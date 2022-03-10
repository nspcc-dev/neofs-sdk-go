package walletconnect

import (
	"encoding/hex"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/stretchr/testify/require"
)

const wif = "KxyjQ8eUa4FHt3Gvioyt1Wz29cTUrE4eTqX3yFSk1YFCsPL8uNsY"

func TestSignMessage(t *testing.T) {
	p1, err := keys.NewPrivateKey()
	require.NoError(t, err)

	msg := []byte("NEO")
	result, err := signMessage(p1, msg)
	require.NoError(t, err)
	require.Equal(t, p1.PublicKey().Bytes(), result.PublicKey)
	require.Equal(t, 16, len(result.Salt))
	require.Equal(t, keys.SignatureLen, len(result.Data))
	require.Equal(t, 4+1+16*2+3+2, len(result.Message))

	require.True(t, verifyMessage(p1.PublicKey(), result))

	t.Run("invalid public key", func(t *testing.T) {
		p2, err := keys.NewPrivateKey()
		require.NoError(t, err)
		require.False(t, verifyMessage(p2.PublicKey(), result))
	})
	t.Run("invalid signature", func(t *testing.T) {
		result := result
		result.Data[0] ^= 0xFF
		require.False(t, verifyMessage(p1.PublicKey(), result))
	})
}

func TestSign(t *testing.T) {
	p1, err := keys.NewPrivateKey()
	require.NoError(t, err)

	msg := []byte("NEO")
	sign, err := Sign(p1, msg)
	require.NoError(t, err)
	require.True(t, Verify(p1.PublicKey(), msg, sign))

	t.Run("invalid public key", func(t *testing.T) {
		p2, err := keys.NewPrivateKey()
		require.NoError(t, err)
		require.False(t, Verify(p2.PublicKey(), msg, sign))
	})
	t.Run("invalid signature", func(t *testing.T) {
		sign[0] ^= 0xFF
		require.False(t, Verify(p1.PublicKey(), msg, sign))
	})
}

func TestVerifyNeonWallet(t *testing.T) {
	// Test values from this GIF https://github.com/CityOfZion/neon-wallet/pull/2390 .
	pub, err := keys.NewPublicKeyFromString("02ce6228ba2cb2fc235be93aff9cd5fc0851702eb9791552f60db062f01e3d83f6")
	require.NoError(t, err)
	data, err := hex.DecodeString("90ab1886ca0bece59b982d9ade8f5598065d651362fb9ce45ad66d0474b89c0b80913c8f0118a282acbdf200a429ba2d81bc52534a53ab41a2c6dfe2f0b4fb1b")
	require.NoError(t, err)
	salt, err := hex.DecodeString("d41e348afccc2f3ee45cd9f5128b16dc")
	require.NoError(t, err)
	msg, err := hex.DecodeString("010001f05c6434316533343861666363633266336565343563643966353132386231366463436172616c686f2c206d756c65712c206f2062616775697520656820697373756d65726d6f2074616978206c696761646f206e61206d697373e36f3f0000")
	require.NoError(t, err)

	sm := SignedMessage{
		Data:      data,
		Message:   msg,
		PublicKey: pub.Bytes(),
		Salt:      salt,
	}
	require.True(t, verifyMessage(pub, sm))
}
