package eacl

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Target describes the NeoFS parties that are subject to a specific access
// rule.
//
// Target should be created using one of the constructors.
type Target struct {
	role  Role
	subjs [][]byte
}

// NewTargetByRole returns Target for specified role. Use NewTargetByRole in [Record]
// to direct it to subjects with the given role in NeoFS.
func NewTargetByRole(role Role) Target { return Target{role: role} }

// NewTargetByAccounts returns Target for specified set of NeoFS accounts. Use
// NewTargetByAccounts in [Record] to direct access rule to the given subjects in
// NeoFS.
func NewTargetByAccounts(accs []user.ID) Target {
	var res Target
	res.SetAccounts(accs)
	return res
}

// NewTargetByScriptHashes is an alternative to [NewTargetByAccounts] which
// allows to pass accounts as their script hashes.
func NewTargetByScriptHashes(hs []util.Uint160) Target {
	b := make([][]byte, len(hs))
	for i := range hs {
		h := user.NewFromScriptHash(hs[i])
		b[i] = h[:]
	}
	return Target{subjs: b}
}

func ecdsaKeysToPtrs(keys []ecdsa.PublicKey) []*ecdsa.PublicKey {
	keysPtr := make([]*ecdsa.PublicKey, len(keys))

	for i := range keys {
		keysPtr[i] = &keys[i]
	}

	return keysPtr
}

// CopyTo writes deep copy of the [Target] to dst.
func (t Target) CopyTo(dst *Target) {
	dst.role = t.role

	dst.subjs = make([][]byte, len(t.subjs))
	for i := range t.subjs {
		dst.subjs[i] = bytes.Clone(t.subjs[i])
	}
}

// BinaryKeys returns list of public keys to identify
// target subject in a binary format.
//
// Each element of the resulting slice is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
// Deprecated: use [Target.Accounts] instead.
func (t *Target) BinaryKeys() [][]byte {
	var r [][]byte

	for _, key := range t.subjs {
		if len(key) == 33 {
			r = append(r, key)
		}
	}

	return r
}

// SetRawSubjects sets target subjects in a binary format. Each element must be
// either 25-byte NeoFS user ID (see [user.ID]) or 33-byte compressed ECDSA
// public key. Use constructors to work with particular types. SetRawSubjects
// should only be used if you do not want to decode the data and take
// responsibility for its correctness.
func (t *Target) SetRawSubjects(subjs [][]byte) {
	t.subjs = subjs
}

// RawSubjects returns list of public keys or [user.ID] to identify target subject in a binary format.
//
// If element length is 33, it is a serialized compressed public key. See [elliptic.MarshalCompressed], [keys.PublicKey.GetScriptHash].
// If element length is 25, it is a [user.ID]. Use `id := user.ID(element)`.
//
// Using this method is your responsibility.
func (t Target) RawSubjects() [][]byte {
	return t.subjs
}

// SetBinaryKeys sets list of binary public keys to identify
// target subject.
//
// Each element of the keys parameter is a slice of bytes is a serialized compressed public key.
// See [elliptic.MarshalCompressed].
// Deprecated: use [Target.SetAccounts] instead.
func (t *Target) SetBinaryKeys(keys [][]byte) { t.SetRawSubjects(keys) }

// Accounts returns list of accounts to identify target subject.
//
// Use `user := user.ID(slice)` to decode it into a type-specific structure.
func (t Target) Accounts() []user.ID {
	var r []user.ID

	for _, key := range t.subjs {
		if len(key) == user.IDSize {
			r = append(r, user.ID(key))
		}
	}

	return r
}

// SetAccounts sets list of accounts to identify target subject.
func (t *Target) SetAccounts(accounts []user.ID) {
	subjs := make([][]byte, len(accounts))

	for i, acc := range accounts {
		subjs[i] = bytes.Clone(acc[:])
	}
	t.SetRawSubjects(subjs)
}

// SetTargetECDSAKeys converts ECDSA public keys to a binary format and stores
// them in Target.
// Deprecated: use [NewTargetByAccounts] or [Target.SetAccounts] along with
// [user.NewFromECDSAPublicKey] instead.
func SetTargetECDSAKeys(t *Target, pubs ...*ecdsa.PublicKey) {
	binKeys := t.BinaryKeys()
	ln := len(pubs)

	if cap(binKeys) >= ln {
		binKeys = binKeys[:0]
	} else {
		binKeys = make([][]byte, 0, ln)
	}

	for i := range ln {
		binKeys = append(binKeys, (*keys.PublicKey)(pubs[i]).Bytes())
	}

	t.SetBinaryKeys(binKeys)
}

// SetTargetAccounts sets accounts in Target.
// Deprecated: use [NewTargetByScriptHashes] instead.
func SetTargetAccounts(t *Target, accs ...util.Uint160) {
	account := make([]user.ID, len(accs))
	ln := len(accs)

	for i := range ln {
		account[i] = user.NewFromScriptHash(accs[i])
	}

	t.SetAccounts(account)
}

// TargetECDSAKeys interprets binary public keys of Target
// as ECDSA public keys. If any key has a different format,
// the corresponding element will be nil.
// Deprecated: use [Target.RawSubjects] with [keys.PublicKey.DecodeBytes] instead.
func TargetECDSAKeys(t *Target) []*ecdsa.PublicKey {
	binKeys := t.BinaryKeys()
	ln := len(binKeys)

	pubs := make([]*ecdsa.PublicKey, ln)

	for i := range ln {
		p := new(keys.PublicKey)
		if p.DecodeBytes(binKeys[i]) == nil {
			pubs[i] = (*ecdsa.PublicKey)(p)
		}
	}

	return pubs
}

// SetRole sets target subject's role class.
func (t *Target) SetRole(r Role) {
	t.role = r
}

// Role returns target subject's role class.
func (t Target) Role() Role {
	return t.role
}

// ToV2 converts Target to v2 acl.EACLRecord.Target message.
//
// Nil Target converts to nil.
// Deprecated: do not use it.
func (t *Target) ToV2() *v2acl.Target {
	if t != nil {
		return t.toProtoMessage()
	}
	return nil
}

func (t Target) toProtoMessage() *v2acl.Target {
	target := new(v2acl.Target)
	target.SetRole(v2acl.Role(t.role))
	target.SetKeys(t.subjs)

	return target
}

func (t *Target) fromProtoMessage(m *v2acl.Target) error {
	t.role = Role(m.GetRole())
	t.subjs = m.GetKeys()
	return nil
}

// NewTarget creates, initializes and returns blank Target instance.
//
// Defaults:
//   - role: RoleUnspecified;
//   - keys: nil.
//
// Deprecated: use [NewTargetByRole] or [TargetByPublicKeys] instead.
func NewTarget() *Target { return new(Target) }

// NewTargetFromV2 converts v2 acl.EACLRecord.Target message to Target.
// Deprecated: do not use it.
func NewTargetFromV2(target *v2acl.Target) *Target {
	t := new(Target)
	_ = t.fromProtoMessage(target)
	return t
}

// Marshal marshals Target into a protobuf binary form.
func (t Target) Marshal() []byte {
	return t.toProtoMessage().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Target.
func (t *Target) Unmarshal(data []byte) error {
	m := new(v2acl.Target)
	if err := m.Unmarshal(data); err != nil {
		return err
	}
	return t.fromProtoMessage(m)
}

// MarshalJSON encodes Target to protobuf JSON format.
func (t Target) MarshalJSON() ([]byte, error) {
	return t.toProtoMessage().MarshalJSON()
}

// UnmarshalJSON decodes Target from protobuf JSON format.
func (t *Target) UnmarshalJSON(data []byte) error {
	m := new(v2acl.Target)
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	return t.fromProtoMessage(m)
}
