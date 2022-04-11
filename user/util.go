package user

import (
	"crypto/ecdsa"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// IDFromKey forms the ID using script hash calculated for the given key.
func IDFromKey(id *ID, key ecdsa.PublicKey) {
	id.SetScriptHash((*keys.PublicKey)(&key).GetScriptHash())
}
