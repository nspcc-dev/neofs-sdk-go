package eacl

import (
	"errors"
	"fmt"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
)

// Record represents an access rule operating in NeoFS access management. The
// rule is applied when some party requests access to a certain NeoFS resource.
// Record is a structural descriptor: see [Table] for detailed behavior.
type Record struct {
	action    Action
	operation acl.Op
	filters   []Filter
	targets   []Target
}

func isEmptyRecord(r Record) bool {
	return r.action == 0 && r.operation == 0 && len(r.filters) == 0 && len(r.targets) == 0
}

func recordToAPI(r Record) *apiacl.EACLRecord {
	if isEmptyRecord(r) {
		return nil
	}

	m := &apiacl.EACLRecord{
		Operation: apiacl.Operation(r.operation),
		Action:    apiacl.Action(r.action),
	}

	if r.filters != nil {
		m.Filters = make([]*apiacl.EACLRecord_Filter, len(r.filters))
		for i := range r.filters {
			m.Filters[i] = filterToAPI(r.filters[i])
		}
	} else {
		m.Filters = nil
	}

	if r.targets != nil {
		m.Targets = make([]*apiacl.EACLRecord_Target, len(r.targets))
		for i := range r.targets {
			m.Targets[i] = targetToAPI(r.targets[i])
		}
	} else {
		m.Targets = nil
	}

	return m
}

func (r *Record) readFromV2(m *apiacl.EACLRecord, checkFieldPresence bool) error {
	var err error
	if len(m.Targets) > 0 {
		r.targets = make([]Target, len(m.Targets))
		for i := range m.Targets {
			if m.Targets[i] != nil {
				err = r.targets[i].readFromV2(m.Targets[i], checkFieldPresence)
				if err != nil {
					return fmt.Errorf("invalid target #%d: %w", i, err)
				}
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing target subjects")
	}

	if m.Filters != nil {
		r.filters = make([]Filter, len(m.Filters))
		for i := range m.Filters {
			if m.Filters[i] != nil {
				err = r.filters[i].readFromV2(m.Filters[i], checkFieldPresence)
				if err != nil {
					return fmt.Errorf("invalid filter #%d: %w", i, err)
				}
			}
		}
	} else {
		r.filters = nil
	}

	r.action = Action(m.Action)
	r.operation = acl.Op(m.Operation)

	return nil
}

// CopyTo writes deep copy of the [Record] to dst.
func (r Record) CopyTo(dst *Record) {
	dst.action = r.action
	dst.operation = r.operation

	if r.filters != nil {
		dst.filters = make([]Filter, len(r.filters))
		copy(dst.filters, r.filters)
	} else {
		dst.filters = nil
	}

	if r.targets != nil {
		dst.targets = make([]Target, len(r.targets))
		for i := range r.targets {
			r.targets[i].CopyTo(&dst.targets[i])
		}
	} else {
		dst.targets = nil
	}
}

// Targets returns list of target subjects for which this access rule applies.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Record.SetTargets].
func (r Record) Targets() []Target {
	return r.targets
}

// SetTargets sets list of target subjects for which this access rule applies.
//
// See also [Record.Targets].
func (r *Record) SetTargets(targets []Target) {
	r.targets = targets
}

// Filters returns list of filters to match the requested resource to this
// access rule. Zero rule has no filters which makes it applicable to any
// resource.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Record.SetFilters].
func (r Record) Filters() []Filter {
	return r.filters
}

// SetFilters returns list of filters to match the requested resource to this
// access rule. Empty list applies the rule to all resources.
//
// See also [Record.Filters].
func (r *Record) SetFilters(fs []Filter) {
	r.filters = fs
}

// Operation returns operation executed by the subject to match.
//
// See also [Record.SetOperation].
func (r Record) Operation() acl.Op {
	return r.operation
}

// SetOperation sets operation executed by the subject to match.
//
// See also [Record.SetOperation].
func (r *Record) SetOperation(operation acl.Op) {
	r.operation = operation
}

// Action returns action on the target subject when the access rule matches.
//
// See also [Record.SetAction].
func (r Record) Action() Action {
	return r.action
}

// SetAction sets action on the target subject when the access rule matches.
//
// See also [Record.Action].
func (r *Record) SetAction(action Action) {
	r.action = action
}
