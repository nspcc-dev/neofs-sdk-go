package eacl

import (
	"bytes"
	"errors"
	"fmt"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// Record represents an access rule operating in NeoFS access management. The
// rule is applied when some party requests access to a certain NeoFS resource.
// Record is a structural descriptor: see [Table] for detailed behavior.
type Record struct {
	action    Action
	operation acl.Op
	filters   []Filter
	target    Target
}

const lastOp = acl.OpObjectHash + 1

// NewRecord returns Record leading to a given action when particular subject
// attempts to access some resource under the following conditions:
//   - the subject executes specified operation
//   - the subject matches the specified rule target
//   - filter list is empty or the resource matches all of them
//
// Both action and op MUST be supported values from corresponding enum. Target
// and all provided filters MUST be correctly constructed.
func NewRecord(action Action, op acl.Op, target Target, filters ...Filter) Record {
	r := Record{
		action:    action,
		operation: op,
		filters:   filters,
		target:    target,
	}

	if msg := r.validate(); msg != "" {
		panic(msg)
	}

	return r
}

// returns message about docs violation or zero if everything is OK.
func (r Record) validate() string {
	switch {
	case r.action <= 0 || r.action >= lastAction:
		return fmt.Sprintf("unsupported value %v from the enum %T", r.action, r.action)
	case r.operation <= 0 || r.operation >= lastOp:
		return fmt.Sprintf("unsupported value %v from the enum %T", r.operation, r.operation)
	}

	if msg := r.target.validate(); msg != "" {
		return fmt.Sprintf("invalid target: %s", msg)
	}

	for i := range r.filters {
		if msg := r.filters[i].validate(); msg != "" {
			panic(fmt.Sprintf("invalid filter #%d: %s", i, msg))
		}
	}

	return ""
}

// copyTo writes deep copy of the [Record] to dst.
func (r Record) copyTo(dst *Record) {
	dst.action = r.action
	dst.operation = r.operation

	if r.filters != nil {
		dst.filters = make([]Filter, len(r.filters))
		copy(dst.filters, r.filters)
	} else {
		dst.filters = nil
	}

	r.target.copyTo(&dst.target)
}

// readFromV2 reads Record from the [v2acl.Record] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. Behavior is
// forward-compatible:
//   - unknown enum values are considered valid
//   - unknown format of binary public keys is considered valid
func (r *Record) readFromV2(m v2acl.Record) error {
	targets := m.GetTargets()
	if len(targets) == 0 {
		return errors.New("missing targets")
	}

	err := r.target.readFromV2(targets)
	if err != nil {
		return err
	}

	filters := m.GetFilters()
	if filters != nil {
		r.filters = make([]Filter, len(filters))
		for i := range filters {
			err := r.filters[i].readFromV2(filters[i])
			if err != nil {
				return fmt.Errorf("invalid filter #%d: %w", i, err)
			}
		}
	} else {
		r.filters = nil
	}

	r.action = Action(m.GetAction())
	r.operation = acl.Op(m.GetOperation())

	return nil
}

// writeToV2 writes Record to [v2acl.Target] messages of the NeoFS API protocol.
func (r Record) writeToV2(m *v2acl.Record) {
	if r.filters != nil {
		filters := make([]v2acl.HeaderFilter, len(r.filters))
		for i := range r.filters {
			r.filters[i].writeToV2(&filters[i])
		}

		m.SetFilters(filters)
	} else {
		m.SetFilters(nil)
	}

	m.SetTargets(r.target.toV2())
	m.SetAction(v2acl.Action(r.action))
	m.SetOperation(v2acl.Operation(r.operation))
}

// IsForRole checks whether the access rule applies to the given role or not.
// Role MUST be one of the supported enum values.
func (r Record) IsForRole(role Role) bool {
	switch role {
	default:
		panic(fmt.Sprintf("unsupported value of enum %T: %v", role, role))
	case
		RoleContainerOwner,
		RoleSystem,
		RoleOthers:
	}

	for i := range r.target.roles {
		if r.target.roles[i] == role {
			return true
		}
	}

	return false
}

// TargetBinaryKeys returns binary-encoded public keys of subjects to which this
// access rule is matched.
//
// Return value MUST NOT be mutated, make a copy first if needed.
//
// See also [Record.IsForKey].
func (r Record) TargetBinaryKeys() [][]byte {
	return r.target.keys
}

// IsForKey checks whether the access rule applies to the given public key or
// not.
//
// See also [Record.TargetBinaryKeys].
func (r Record) IsForKey(pubKey neofscrypto.PublicKey) bool {
	bKey := neofscrypto.PublicKeyBytes(pubKey)

	for i := range r.target.keys {
		if bytes.Equal(r.target.keys[i], bKey) {
			return true
		}
	}

	return false
}

// Filters returns list of filters to match the requested resource to this
// access rule.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (r Record) Filters() []Filter {
	return r.filters
}

// Op returns operation executed by the subject to match.
//
// See also [Record.IsForOp].
func (r Record) Op() acl.Op {
	return r.operation
}

// IsForOp checks whether the access rule matches given operation executed by
// the subject or not.
//
// See also [Record.Op].
func (r Record) IsForOp(op acl.Op) bool {
	return r.operation == op
}

// Action returns action on the target subject when the access rule matches.
func (r Record) Action() Action {
	return r.action
}
