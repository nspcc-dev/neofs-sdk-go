package neofsecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignerWalletConnect_Sign(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	signer := (*SignerWalletConnect)(key)

	data := []byte("some_message")
	sig, err := signer.Sign(data)
	require.NoError(t, err)
	require.Len(t, sig, 64+16)
	require.True(t, signer.Public().Verify(data, sig))
}

func TestPublicKey_Verify(t *testing.T) {
	const data = "Hello, world!"
	const pubKeyHex = "02b6027dffdc35e66d4469ae68c8382a8eb7e40b4dd01d812b0470a7d4864fb858"
	const sigHex = "ae6ea59b2028f696d9c2626aba9677800a0fcd5441aa7f66d085c9c070de61faa11d6002e821d4407bb8b8690571d6ac7e79706510d6c89175705d156a234c53d3af772b1846f96034fd4fa740744553"

	b, err := hex.DecodeString(pubKeyHex)
	require.NoError(t, err)

	var pubKey PublicKeyWalletConnect
	require.NoError(t, pubKey.Decode(b))

	sig, err := hex.DecodeString(sigHex)
	require.NoError(t, err)

	require.True(t, pubKey.Verify([]byte(data), sig))
}

func TestVerifyWalletConnect(t *testing.T) {
	for _, testCase := range []struct {
		msgHex    string
		pubKeyHex string
		sigHex    string
		saltHex   string
	}{
		{ // Test values from this GIF https://github.com/CityOfZion/neon-wallet/pull/2390 .
			msgHex:    "436172616c686f2c206d756c65712c206f2062616775697520656820697373756d65726d6f2074616978206c696761646f206e61206d697373e36f3f",
			pubKeyHex: "02ce6228ba2cb2fc235be93aff9cd5fc0851702eb9791552f60db062f01e3d83f6",
			saltHex:   "d41e348afccc2f3ee45cd9f5128b16dc",
			sigHex:    "90ab1886ca0bece59b982d9ade8f5598065d651362fb9ce45ad66d0474b89c0b80913c8f0118a282acbdf200a429ba2d81bc52534a53ab41a2c6dfe2f0b4fb1b",
		},
		{ // Test value from wallet connect integration test
			msgHex:    "313233343536", // ascii string "123456"
			pubKeyHex: "03bd9108c0b49f657e9eee50d1399022bd1e436118e5b7529a1b7cd606652f578f",
			saltHex:   "2c5b189569e92cce12e1c640f23e83ba",
			sigHex:    "510caa8cb6db5dedf04d215a064208d64be7496916d890df59aee132db8f2b07532e06f7ea664c4a99e3bcb74b43a35eb9653891b5f8701d2aef9e7526703eaa",
		},
		{ // Test value from wallet connect integration test
			msgHex:    "",
			pubKeyHex: "03bd9108c0b49f657e9eee50d1399022bd1e436118e5b7529a1b7cd606652f578f",
			saltHex:   "58c86b2e74215b4f36b47d731236be3b",
			sigHex:    "1e13f248962d8b3b60708b55ddf448d6d6a28c6b43887212a38b00bf6bab695e61261e54451c6e3d5f1f000e5534d166c7ca30f662a296d3a9aafa6d8c173c01",
		},
	} {
		pubKeyBin, err := hex.DecodeString(testCase.pubKeyHex)
		require.NoError(t, err)
		sig, err := hex.DecodeString(testCase.sigHex)
		require.NoError(t, err)
		salt, err := hex.DecodeString(testCase.saltHex)
		require.NoError(t, err)
		msg, err := hex.DecodeString(testCase.msgHex)
		require.NoError(t, err)

		pubKey := ecdsa.PublicKey{Curve: elliptic.P256()}
		pubKey.X, pubKey.Y = elliptic.UnmarshalCompressed(pubKey.Curve, pubKeyBin)
		require.NotNil(t, pubKey.X)
		require.True(t, verifyWalletConnect(&pubKey, msg, sig, salt), testCase)
	}

}
