package user

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// IDFromKey forms the ID using script hash calculated for the given key.
func IDFromKey(id *ID, key []byte) error {
	var pk keys.PublicKey
	if err := pk.DecodeBytes(key); err != nil {
		return fmt.Errorf("IDFromKey %v", err)
	}

	id.SetScriptHash(pk.GetScriptHash())
	return nil
}

// IDFromSigner forms the ID using script hash calculated for the given key.
func IDFromSigner(id *ID, signer neofscrypto.Signer) error {
	public := signer.Public()

	key := make([]byte, public.MaxEncodedSize())
	key = key[:public.Encode(key)]

	return IDFromKey(id, key)
}
