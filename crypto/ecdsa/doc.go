/*
Package neofsecdsa collects ECDSA primitives for NeoFS cryptography.

Signer and PublicKey provide corresponding interfaces from neofscrypto package.

Package import causes registration of next signature schemes via neofscrypto.RegisterScheme:
  - neofscrypto.ECDSA_SHA512
  - neofscrypto.ECDSA_DETERMINISTIC_SHA256

*/
package neofsecdsa
