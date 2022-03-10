package neofsecdsa

import neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"

func init() {
	neofscrypto.RegisterScheme(neofscrypto.ECDSA_SHA512, func() neofscrypto.PublicKey {
		return new(PublicKey)
	})

	neofscrypto.RegisterScheme(neofscrypto.ECDSA_DETERMINISTIC_SHA256, func() neofscrypto.PublicKey {
		return new(PublicKeyRFC6979)
	})

	neofscrypto.RegisterScheme(neofscrypto.ECDSA_WALLETCONNECT, func() neofscrypto.PublicKey {
		return new(PublicKeyWalletConnect)
	})
}
