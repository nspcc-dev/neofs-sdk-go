package user

import (
	"crypto/ecdsa"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
)

// Signer is an interface of entities that can be used for signing operations
// in NeoFS. It is the same as [neofscrypto.Signer], but has an extra method to retrieve [ID].
type Signer interface {
	neofscrypto.Signer

	UserID() ID
}

// SignerRFC6979 wraps [ecdsa.PrivateKey] and represents signer based on deterministic
// ECDSA with SHA-256 hashing (RFC 6979). Provides [Signer] interface.
//
// Instances SHOULD be initialized with [NewSignerRFC6979] or [NewSignerRFC6979WithID].
type SignerRFC6979 struct {
	neofsecdsa.SignerRFC6979
	userID ID
}

// NewSignerRFC6979 is a constructor for [SignerRFC6979].
func NewSignerRFC6979(pk ecdsa.PrivateKey) *SignerRFC6979 {
	var id ID
	id.SetScriptHash((*keys.PublicKey)(&pk.PublicKey).GetScriptHash())

	return &SignerRFC6979{
		userID:        id,
		SignerRFC6979: neofsecdsa.SignerRFC6979(pk),
	}
}

// NewSignerRFC6979WithID is a constructor for [SignerRFC6979] where you may specify [ID] associated with this signer.
func NewSignerRFC6979WithID(pk ecdsa.PrivateKey, id ID) *SignerRFC6979 {
	return &SignerRFC6979{
		SignerRFC6979: neofsecdsa.SignerRFC6979(pk),
		userID:        id,
	}
}

// UserID returns the [ID] using script hash calculated for the given key.
func (s SignerRFC6979) UserID() ID {
	return s.userID
}

// StaticSigner emulates real sign and contains already precalculated hash.
// Provides [Signer] interface.
type StaticSigner struct {
	neofscrypto.StaticSigner
	id ID
}

// NewStaticSignerWithID creates new StaticSigner with specified [ID].
func NewStaticSignerWithID(scheme neofscrypto.Scheme, sig []byte, pubKey neofscrypto.PublicKey, id ID) *StaticSigner {
	return &StaticSigner{
		StaticSigner: *neofscrypto.NewStaticSigner(scheme, sig, pubKey),
		id:           id,
	}
}

// UserID returns underlying [ID].
func (s *StaticSigner) UserID() ID {
	return s.id
}

// ResolveFromECDSAPublicKey resolves [ID] from the given [ecdsa.PublicKey].
func ResolveFromECDSAPublicKey(pk ecdsa.PublicKey) ID {
	var id ID
	id.SetScriptHash((*keys.PublicKey)(&pk).GetScriptHash())

	return id
}
