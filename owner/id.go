package owner

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// ID represents v2-compatible owner identifier.
type ID refs.OwnerID

var errInvalidIDString = errors.New("incorrect format of the string owner ID")

// ErrEmptyPublicKey when public key passed to Verify method is nil.
var ErrEmptyPublicKey = errors.New("empty public key")

// NEO3WalletSize contains size of neo3 wallet.
const NEO3WalletSize = 25

// NewIDFromV2 wraps v2 OwnerID message to ID.
//
// Nil refs.OwnerID converts to nil.
func NewIDFromV2(idV2 *refs.OwnerID) *ID {
	return (*ID)(idV2)
}

// NewID creates and initializes blank ID.
//
// Works similar as NewIDFromV2(new(OwnerID)).
//
// Defaults:
//  - value: nil.
func NewID() *ID {
	return NewIDFromV2(new(refs.OwnerID))
}

// SetPublicKey sets owner identifier value to the provided NEO3 public key.
func (id *ID) SetPublicKey(pub *ecdsa.PublicKey) {
	(*refs.OwnerID)(id).SetValue(PublicKeyToIDBytes(pub))
}

// SetScriptHash sets owner identifier value to the provided NEO3 script hash.
func (id *ID) SetScriptHash(u util.Uint160) {
	(*refs.OwnerID)(id).SetValue(ScriptHashToIDBytes(u))
}

// ToV2 returns the v2 owner ID message.
//
// Nil ID converts to nil.
func (id *ID) ToV2() *refs.OwnerID {
	return (*refs.OwnerID)(id)
}

// String implements fmt.Stringer.
func (id *ID) String() string {
	return base58.Encode((*refs.OwnerID)(id).GetValue())
}

// Equal defines a comparison relation on ID's.
//
// ID's are equal if they have the same binary representation.
func (id *ID) Equal(id2 *ID) bool {
	return bytes.Equal(
		(*refs.ObjectID)(id).GetValue(),
		(*refs.ObjectID)(id2).GetValue(),
	)
}

// NewIDFromPublicKey creates new owner identity from ECDSA public key.
func NewIDFromPublicKey(pub *ecdsa.PublicKey) *ID {
	id := NewID()
	id.SetPublicKey(pub)

	return id
}

// NewIDFromN3Account creates new owner identity from N3 wallet account.
func NewIDFromN3Account(acc *wallet.Account) *ID {
	return NewIDFromPublicKey(
		(*ecdsa.PublicKey)(acc.PrivateKey().PublicKey()))
}

// Parse converts base58 string representation into ID.
func (id *ID) Parse(s string) error {
	data, err := base58.Decode(s)
	if err != nil {
		return fmt.Errorf("could not parse owner.ID from string: %w", err)
	} else if !valid(data) {
		return errInvalidIDString
	}

	(*refs.OwnerID)(id).SetValue(data)

	return nil
}

// Valid returns true if id is a valid owner id.
// The rules for v2 are the following:
// 1. Must be 25 bytes in length.
// 2. Must have N3 address prefix-byte.
// 3. Last 4 bytes must contain the address checksum.
func (id *ID) Valid() bool {
	rawID := id.ToV2().GetValue()
	return valid(rawID)
}

func valid(rawID []byte) bool {
	if len(rawID) != NEO3WalletSize {
		return false
	}
	if rawID[0] != address.NEO3Prefix {
		return false
	}

	const boundIndex = NEO3WalletSize - 4
	return bytes.Equal(rawID[boundIndex:], hash.Checksum(rawID[:boundIndex]))
}

// Marshal marshals ID into a protobuf binary form.
func (id *ID) Marshal() []byte {
	return (*refs.OwnerID)(id).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of ID.
func (id *ID) Unmarshal(data []byte) error {
	return (*refs.OwnerID)(id).Unmarshal(data)
}

// MarshalJSON encodes ID to protobuf JSON format.
func (id *ID) MarshalJSON() ([]byte, error) {
	return (*refs.OwnerID)(id).MarshalJSON()
}

// UnmarshalJSON decodes ID from protobuf JSON format.
func (id *ID) UnmarshalJSON(data []byte) error {
	return (*refs.OwnerID)(id).UnmarshalJSON(data)
}

// PublicKeyToIDBytes converts public key to a byte slice of NEO3WalletSize length.
// It is similar to decoding a NEO3 address but is inlined to skip base58 encoding-decoding step
// make it clear that no errors can occur.
func PublicKeyToIDBytes(pub *ecdsa.PublicKey) []byte {
	sh := (*keys.PublicKey)(pub).GetScriptHash()
	return ScriptHashToIDBytes(sh)
}

// ScriptHashToIDBytes converts NEO3 script hash to a byte slice of NEO3WalletSize length.
func ScriptHashToIDBytes(sh util.Uint160) []byte {
	b := make([]byte, NEO3WalletSize)
	b[0] = address.Prefix
	copy(b[1:], sh.BytesBE())
	copy(b[21:], hash.Checksum(b[:21]))
	return b
}
