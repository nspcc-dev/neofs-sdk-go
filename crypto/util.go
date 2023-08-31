package neofscrypto

import "encoding/hex"

// StringifyKeyBinary returns string with HEX representation of source.
// Format can be changed and it's unsafe to rely on it beyond human-readable output.
//
// Parameter src is a serialized compressed public key. See [elliptic.MarshalCompressed].
func StringifyKeyBinary(src []byte) string {
	return hex.EncodeToString(src)
}

// PublicKeyBytes returns binary-encoded PublicKey. Use [PublicKey.Encode] to
// avoid new slice allocation.
func PublicKeyBytes(pubKey PublicKey) []byte {
	b := make([]byte, pubKey.MaxEncodedSize())
	return b[:pubKey.Encode(b)]
}
