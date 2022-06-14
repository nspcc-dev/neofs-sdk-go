package walletconnect

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
)

// SignedMessage contains mirrors `SignedMessage` struct from the WalletConnect API.
// https://neon.coz.io/wksdk/core/modules.html#SignedMessage
type SignedMessage struct {
	Data      []byte
	Message   []byte
	PublicKey []byte
	Salt      []byte
}

const saltSize = 16

// Sign signs message using WalletConnect API. The returned signature
// contains RFC6979 signature and 16-byte salt.
func Sign(p *keys.PrivateKey, msg []byte) ([]byte, error) {
	sm, err := signMessage(p, msg)
	if err != nil {
		return nil, err
	}
	return append(sm.Data, sm.Salt...), nil
}

// Verify verifies message using WalletConnect API. The returned signature
// contains RFC6979 signature and 16-byte salt.
func Verify(p *keys.PublicKey, data, sign []byte) bool {
	if len(sign) != keys.SignatureLen+saltSize {
		return false
	}

	salt := sign[keys.SignatureLen:]
	return verifyMessage(p, SignedMessage{
		Data:    sign[:keys.SignatureLen],
		Message: createMessageWithSalt(data, salt),
		Salt:    salt,
	})
}

// signMessage signs message with a private key and returns structure similar to
// `signMessage` of the WalletConnect API.
// https://github.com/CityOfZion/wallet-connect-sdk/blob/89c236b/packages/wallet-connect-sdk-core/src/index.ts#L496
// https://github.com/CityOfZion/neon-wallet/blob/1174a9388480e6bbc4f79eb13183c2a573f67ca8/app/context/WalletConnect/helpers.js#L133
func signMessage(p *keys.PrivateKey, msg []byte) (SignedMessage, error) {
	var salt [16]byte
	_, _ = rand.Read(salt[:])

	msg = createMessageWithSalt(msg, salt[:])
	return SignedMessage{
		Data:      p.Sign(msg),
		Message:   msg,
		PublicKey: p.PublicKey().Bytes(),
		Salt:      salt[:],
	}, nil
}

// verifyMessage verifies message with a private key and returns structure similar to
// `verifyMessage` of WalletConnect API.
// https://github.com/CityOfZion/wallet-connect-sdk/blob/89c236b/packages/wallet-connect-sdk-core/src/index.ts#L515
// https://github.com/CityOfZion/neon-wallet/blob/1174a9388480e6bbc4f79eb13183c2a573f67ca8/app/context/WalletConnect/helpers.js#L147
func verifyMessage(p *keys.PublicKey, m SignedMessage) bool {
	if p == nil {
		var err error
		p, err = keys.NewPublicKeyFromBytes(m.PublicKey, elliptic.P256())
		if err != nil {
			return false
		}
	}
	h := sha256.Sum256(m.Message)
	return p.Verify(m.Data, h[:])
}

func createMessageWithSalt(msg, salt []byte) []byte {
	// 4 byte prefix + length of the message with salt in bytes +
	// + salt + message + 2 byte postfix.
	w := io.NewBufBinWriter()
	saltedSize := len(salt)*2 + len(msg)
	w.Grow(4 + io.GetVarSize(saltedSize) + saltedSize + 2)

	w.WriteBytes([]byte{0x01, 0x00, 0x01, 0xf0}) // fixed prefix
	w.WriteVarUint(uint64(saltedSize))
	w.WriteBytes([]byte(hex.EncodeToString(salt[:]))) // for some reason we encode salt in hex
	w.WriteBytes(msg)
	w.WriteBytes([]byte{0x00, 0x00})

	return w.Bytes()
}
