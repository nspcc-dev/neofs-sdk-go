package user

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// ErrOwnerExtract is returned when failed to extract account info from key.
var ErrOwnerExtract = errors.New("decode owner failed")

// IDFromKey forms the ID using script hash calculated for the given key.
func IDFromKey(id *ID, key []byte) error {
	var pk keys.PublicKey
	if err := pk.DecodeBytes(key); err != nil {
		return fmt.Errorf("%w: %v", ErrOwnerExtract, err)
	}

	id.SetScriptHash(pk.GetScriptHash())
	return nil
}
