package eacl

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/util/slice"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// Target describes the parties that are subject to a specific access rule.
type Target struct {
	roles []Role
	keys  [][]byte
}

// NewTarget returns Target that matches parties with at least one of the
// specified roles or public keys.
//
// At least one argument MUST be non-empty. Both arguments MUST NOT be mutated
// within resulting Target lifetime. All roles MUST be supported values from the
// corresponding enum except [RoleSystem] which is no longer supported. All keys
// MUST be non-nil.
//
// See also other helper constructors.
func NewTarget(roles []Role, publicKeys []neofscrypto.PublicKey) Target {
	var res Target
	res.roles = roles

	if publicKeys != nil {
		res.keys = make([][]byte, len(publicKeys))

		for i := range publicKeys {
			if publicKeys[i] == nil {
				panic(fmt.Sprintf("key #%d is nil", i))
			}

			res.keys[i] = neofscrypto.PublicKeyBytes(publicKeys[i])
		}
	}

	if msg := res.validate(); msg != "" {
		panic(msg)
	}

	return res
}

// returns message about docs violation or zero if everything is OK.
func (t Target) validate() string {
	if len(t.roles)+len(t.keys) == 0 {
		return "neither roles nor keys are presented"
	}

	for i := range t.keys {
		if len(t.keys[i]) == 0 {
			return fmt.Sprintf("invalid key #%d: key is empty", i)
		}
	}

	for i := range t.roles {
		switch t.roles[i] {
		default:
			panic(fmt.Sprintf("invalid role #%d: forbidden value %v of enum %T", i, t.roles[i], t.roles[i]))
		case
			RoleContainerOwner,
			RoleOthers:
		}
	}

	return ""
}

// copyTo writes deep copy of the [Target] to dst.
func (t Target) copyTo(dst *Target) {
	if t.roles != nil {
		dst.roles = make([]Role, len(t.roles))
		copy(dst.roles, t.roles)
	} else {
		dst.roles = nil
	}

	if t.keys != nil {
		dst.keys = make([][]byte, len(t.keys))
		for i := range t.keys {
			dst.keys[i] = slice.Copy(t.keys[i])
		}
	} else {
		dst.keys = nil
	}
}

// NewTargetWithRole returns Target for the given role only.
//
// See also [NewTarget].
func NewTargetWithRole(role Role) Target {
	return NewTarget([]Role{role}, nil)
}

// NewTargetWithKey returns Target for the given public key only.
//
// See also [NewTarget], [NewTargetWithKeys].
func NewTargetWithKey(key neofscrypto.PublicKey) Target {
	return NewTargetWithKeys([]neofscrypto.PublicKey{key})
}

// NewTargetWithKeys returns Target for the given list of public keys.
//
// See also [NewTarget], [NewTargetWithKey].
func NewTargetWithKeys(publicKeys []neofscrypto.PublicKey) Target {
	return NewTarget(nil, publicKeys)
}

// readFromV2 reads Target from the [v2acl.Target] messages. Returns an error if
// any message is malformed according to the NeoFS API V2 protocol. Behavior is
// forward-compatible:
//   - unknown enum values are considered valid
//   - unknown format of binary public keys is considered valid
//
// The argument MUST be non-empty.
func (t *Target) readFromV2(ms []v2acl.Target) error {
	if len(ms) == 0 {
		panic("empty slice of targets")
	}

	for i := range ms {
		role := ms[i].GetRole()
		keys := ms[i].GetKeys()

		if role == 0 && len(keys) == 0 {
			return fmt.Errorf("invalid target #%d: neither role nor public keys are set", i)
		}

		if role != 0 && len(keys) != 0 {
			return fmt.Errorf("invalid target #%d: both role and public keys are set", i)
		}

		for j := range keys {
			if len(keys[j]) == 0 {
				return fmt.Errorf("invalid target #%d: empty key #%d", i, j)
			}
		}

		if role != 0 {
			t.roles = append(t.roles, Role(role))
		} else {
			t.keys = append(t.keys, keys...)
		}
	}

	return nil
}

// toV2 writes Target to [v2acl.Target] messages of the NeoFS API protocol.
func (t Target) toV2() []v2acl.Target {
	var ms []v2acl.Target

	withKeys := len(t.keys) > 0
	if withKeys {
		ms = make([]v2acl.Target, len(t.roles)+1)
	} else {
		ms = make([]v2acl.Target, len(t.roles))
	}

	for i := range t.roles {
		ms[i].SetRole(v2acl.Role(t.roles[i]))
	}

	if withKeys {
		ms[len(ms)-1].SetKeys(t.keys)
	}

	return ms
}
