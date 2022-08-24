/*
Package neofsecdsa collects ECDSA primitives for NeoFS cryptography.

Signer and PublicKey support ECDSA signature algorithm with SHA-512 hashing.
SignerRFC6979 and PublicKeyRFC6979 implement signature algorithm described in RFC 6979.
All these types provide corresponding interfaces from neofscrypto package.

Package import causes registration of next signature schemes via neofscrypto.RegisterScheme:
  - neofscrypto.ECDSA_SHA512
  - neofscrypto.ECDSA_DETERMINISTIC_SHA256
*/
package neofsecdsa
