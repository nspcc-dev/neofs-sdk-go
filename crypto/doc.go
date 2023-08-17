/*
Package neofscrypto collects NeoFS cryptographic primitives.

Signer type unifies entities for signing NeoFS data.

SDK natively supports several signature schemes that are implemented
in nested packages.

PublicKey allows to verify signatures.

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.
*/
package neofscrypto
