package eacl

import (
	"bytes"
	"errors"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
)

// Target describes the parties that are subject to a specific access rule.
type Target struct {
	role Role
	keys [][]byte
}

// CopyTo writes deep copy of the [Target] to dst.
func (t Target) CopyTo(dst *Target) {
	dst.role = t.role

	if t.keys != nil {
		dst.keys = make([][]byte, len(t.keys))
		for i := range t.keys {
			dst.keys[i] = bytes.Clone(t.keys[i])
		}
	} else {
		dst.keys = nil
	}
}

func isEmptyTarget(t Target) bool {
	return t.role == 0 && len(t.keys) == 0
}

func targetToAPI(t Target) *apiacl.EACLRecord_Target {
	if isEmptyTarget(t) {
		return nil
	}
	return &apiacl.EACLRecord_Target{
		Role: apiacl.Role(t.role),
		Keys: t.keys,
	}
}

func (t *Target) readFromV2(m *apiacl.EACLRecord_Target, checkFieldPresence bool) error {
	if checkFieldPresence && (m.Role == 0 == (len(m.Keys) == 0)) {
		return errors.New("role and public keys are not mutually exclusive")
	}

	t.role = Role(m.Role)
	t.keys = m.Keys

	return nil
}

// PublicKeys returns list of public keys to identify target subjects. Overlaps
// [Target.Role].
//
// Each element of the resulting slice is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (t Target) PublicKeys() [][]byte {
	return t.keys
}

// SetPublicKeys sets list of binary public keys to identify target subjects.
// Overlaps [Target.SetRole].
//
// Each element of the keys parameter is a slice of bytes is a serialized compressed public key.
// See [elliptic.MarshalCompressed].
func (t *Target) SetPublicKeys(keys [][]byte) {
	t.keys = keys
}

// SetRole sets role to identify group of target subjects. Overlaps with
// [Target.SetPublicKeys].
//
// See also [Target.Role].
func (t *Target) SetRole(r Role) {
	t.role = r
}

// Role returns role to identify group of target subjects. Overlaps with
// [Target.PublicKeys].
//
// See also [Target.SetRole].
func (t Target) Role() Role {
	return t.role
}
