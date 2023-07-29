package user

import (
	"crypto/ecdsa"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
)

// Signer represents a NeoFS user authorized by a digital signature.
type Signer interface {
	// Signer signs data on behalf of the user.
	neofscrypto.Signer
	// UserID returns ID of the associated user.
	UserID() ID
}

type signer struct {
	neofscrypto.Signer
	usr ID
}

func (s signer) UserID() ID {
	return s.usr
}

// NewSigner combines provided [neofscrypto.Signer] and [ID] into [Signer].
//
// See also [NewAutoIDSigner].
func NewSigner(s neofscrypto.Signer, usr ID) Signer {
	return signer{
		Signer: s,
		usr:    usr,
	}
}

func newAutoResolvedSigner(s neofscrypto.Signer, pubKey ecdsa.PublicKey) Signer {
	var id ID
	id.SetScriptHash((*keys.PublicKey)(&pubKey).GetScriptHash())

	return NewSigner(s, id)
}

// NewAutoIDSigner returns [Signer] with neofscrypto.ECDSA_SHA512
// signature scheme and user [ID] automatically resolved from the ECDSA public
// key.
//
// See also [NewAutoIDSignerRFC6979].
func NewAutoIDSigner(key ecdsa.PrivateKey) Signer {
	return newAutoResolvedSigner(neofsecdsa.Signer(key), key.PublicKey)
}

// NewAutoIDSignerRFC6979 is an analogue of [NewAutoIDSigner] but with
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] signature scheme.
func NewAutoIDSignerRFC6979(key ecdsa.PrivateKey) Signer {
	return newAutoResolvedSigner(neofsecdsa.SignerRFC6979(key), key.PublicKey)
}

// ResolveFromECDSAPublicKey resolves [ID] from the given [ecdsa.PublicKey].
func ResolveFromECDSAPublicKey(pk ecdsa.PublicKey) ID {
	var id ID
	id.SetScriptHash((*keys.PublicKey)(&pk).GetScriptHash())

	return id
}
