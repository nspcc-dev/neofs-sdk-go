package eacl

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

// Record represents an access rule operating in NeoFS access management. The
// rule is applied when some party requests access to a certain NeoFS resource.
//
// Record should be created using one of the constructors.
type Record struct {
	action    Action
	operation Operation
	filters   []Filter
	targets   []Target
	comment   string
}

// ConstructRecord constructs new Record representing access rule regulating
// action in relation to specified target subjects when they perform the given
// operation. Optional filters allow to limit the effect of a rule on specific
// resources.
func ConstructRecord(a Action, op Operation, ts []Target, fs ...Filter) Record {
	return Record{action: a, operation: op, filters: fs, targets: ts}
}

// CopyTo writes deep copy of the [Record] to dst.
func (r Record) CopyTo(dst *Record) {
	dst.action = r.action
	dst.operation = r.operation
	dst.filters = slices.Clone(r.filters)
	dst.comment = r.comment

	dst.targets = slices.Clone(r.targets)
	for i, t := range r.targets {
		t.CopyTo(&dst.targets[i])
	}
}

// Targets returns list of target subjects to which this access rule matches.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (r Record) Targets() []Target {
	return r.targets
}

// SetTargets sets list of target subjects to which this access rule matches.
func (r *Record) SetTargets(targets ...Target) {
	r.targets = targets
}

// Filters returns list of filters to match the requested resource to this
// access rule. Absence of filters means that Record is applicable to any
// resource.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (r Record) Filters() []Filter {
	return r.filters
}

// SetFilters returns list of filters to match the requested resource to this
// access rule. Empty list applies the Record to all resources.
func (r *Record) SetFilters(fs []Filter) {
	r.filters = fs
}

// Operation returns operation executed by the subject to match.
func (r Record) Operation() Operation {
	return r.operation
}

// SetOperation sets operation executed by the subject to match.
func (r *Record) SetOperation(operation Operation) {
	r.operation = operation
}

// Action returns action on the target subject when the access rule matches.
func (r Record) Action() Action {
	return r.action
}

// SetAction sets action on the target subject when the access rule matches.
func (r *Record) SetAction(action Action) {
	r.action = action
}

// Comment returns auxiliary data associated with the access rule. The comment
// does not affect permission evaluation.
func (r Record) Comment() string {
	return r.comment
}

// SetComment associates auxiliary data with the access rule. The comment does
// not affect permission evaluation. It returns an error if comment is not valid
// UTF-8 or contains zero bytes.
func (r *Record) SetComment(comment string) error {
	if err := validateComment(comment); err != nil {
		return err
	}

	r.comment = comment
	return nil
}

func validateComment(comment string) error {
	if !utf8.ValidString(comment) {
		return errors.New("invalid UTF-8 comment")
	}
	if strings.ContainsRune(comment, 0) {
		return errors.New("comment contains zero byte")
	}
	return nil
}

type stringEncoder interface {
	EncodeToString() string
}

func (r Record) toProtoMessage() *protoacl.EACLRecord {
	m := &protoacl.EACLRecord{
		Operation: protoacl.Operation(r.operation),
		Action:    protoacl.Action(r.action),
		Comment:   r.comment,
	}

	if r.targets != nil {
		m.Targets = make([]*protoacl.EACLRecord_Target, len(r.targets))
		for i := range r.targets {
			m.Targets[i] = r.targets[i].protoMessage()
		}
	}

	if r.filters != nil {
		m.Filters = make([]*protoacl.EACLRecord_Filter, len(r.filters))
		for i := range r.filters {
			m.Filters[i] = r.filters[i].protoMessage()
		}
	}

	return m
}

func (r *Record) fromProtoMessage(m *protoacl.EACLRecord) error {
	if m.Action < 0 {
		return fmt.Errorf("negative action %d", m.Action)
	}
	if m.Operation < 0 {
		return fmt.Errorf("negative op %d", m.Operation)
	}
	if err := validateComment(m.Comment); err != nil {
		return err
	}

	mt := m.Targets
	r.targets = make([]Target, len(mt))
	for i := range mt {
		if mt[i] == nil {
			return fmt.Errorf("nil target #%d", i)
		}
		if err := r.targets[i].fromProtoMessage(mt[i]); err != nil {
			return fmt.Errorf("invalid subject descriptor #%d: %w", i, err)
		}
	}

	mf := m.Filters
	r.filters = make([]Filter, len(mf))
	for i := range mf {
		if mf[i] == nil {
			return fmt.Errorf("nil filter #%d", i)
		}
		if err := r.filters[i].fromProtoMessage(mf[i]); err != nil {
			return fmt.Errorf("invalid filter #%d: %w", i, err)
		}
	}

	r.action = Action(m.Action)
	r.operation = Operation(m.Operation)
	r.comment = m.Comment

	return nil
}

// Marshal marshals Record into a protobuf binary form.
func (r Record) Marshal() []byte {
	return protoencoding.MarshalMessage(r.toProtoMessage())
}

// Unmarshal unmarshals protobuf binary representation of Record.
func (r *Record) Unmarshal(data []byte) error {
	m := new(protoacl.EACLRecord)
	if err := protoencoding.UnmarshalMessage(data, m); err != nil {
		return err
	}
	return r.fromProtoMessage(m)
}

// MarshalJSON encodes Record to protobuf JSON format.
func (r Record) MarshalJSON() ([]byte, error) {
	return protoencoding.MarshalMessageJSON(r.toProtoMessage())
}

// UnmarshalJSON decodes Record from protobuf JSON format.
func (r *Record) UnmarshalJSON(data []byte) error {
	m := new(protoacl.EACLRecord)
	if err := protoencoding.UnmarshalMessageJSON(data, m); err != nil {
		return err
	}
	return r.fromProtoMessage(m)
}
