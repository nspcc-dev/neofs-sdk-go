package eacl

import (
	"crypto/ecdsa"
	"fmt"
	"slices"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
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

	dst.targets = make([]Target, len(r.targets))
	for i, t := range r.targets {
		var newTarget Target
		t.CopyTo(&newTarget)

		dst.targets[i] = newTarget
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

// AddRecordTarget adds single Target to the Record.
// Deprecated: use [Record.SetTargets] instead.
func AddRecordTarget(r *Record, t *Target) {
	r.SetTargets(append(r.Targets(), *t)...)
}

// AddFormedTarget forms Target with specified Role and list of
// ECDSA public keys and adds it to the Record.
// Deprecated: use [Record.SetTargets] with [TargetByRole] or
// [TargetByPublicKeys] instead. Note that role and public keys are mutually
// exclusive.
func AddFormedTarget(r *Record, role Role, keys ...ecdsa.PublicKey) {
	t := NewTarget()
	t.SetRole(role)

	SetTargetECDSAKeys(t, ecdsaKeysToPtrs(keys)...)
	AddRecordTarget(r, t)
}

type stringEncoder interface {
	EncodeToString() string
}

// AddFilter adds generic filter.
//
// If matcher is [MatchNotPresent], the value must be empty. If matcher is
// numeric (e.g. [MatchNumGT]), value must be a base-10 integer.
// Deprecated: use [ConstructRecord] with [ConstructFilter] instead.
func (r *Record) AddFilter(from FilterHeaderType, matcher Match, name, value string) {
	r.SetFilters(append(r.Filters(), ConstructFilter(from, name, matcher, value)))
}

// AddObjectAttributeFilter adds filter by object attribute.
//
// If m is [MatchNotPresent], the value must be empty. If matcher is numeric
// (e.g. [MatchNumGT]), value must be a base-10 integer.
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] instead.
func (r *Record) AddObjectAttributeFilter(m Match, key, value string) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(key, m, value)))
}

// AddObjectVersionFilter adds filter by object version.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] instead.
func (r *Record) AddObjectVersionFilter(m Match, v *version.Version) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectVersion, m, version.EncodeToString(*v))))
}

// AddObjectIDFilter adds filter by object ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] or
// [NewFilterObjectWithID] instead.
func (r *Record) AddObjectIDFilter(m Match, id oid.ID) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectID, m, id.EncodeToString())))
}

// AddObjectContainerIDFilter adds filter by object container ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] or
// [NewFilterObjectsFromContainer] instead.
func (r *Record) AddObjectContainerIDFilter(m Match, id cid.ID) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectContainerID, m, id.EncodeToString())))
}

// AddObjectOwnerIDFilter adds filter by object owner ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] or
// [NewFilterObjectOwnerEquals] instead.
func (r *Record) AddObjectOwnerIDFilter(m Match, id *user.ID) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectOwnerID, m, id.EncodeToString())))
}

// AddObjectCreationEpoch adds filter by object creation epoch.
//
// The m must not be [MatchNotPresent].
// Deprecated: use [ConstructRecord] with [NewFilterObjectCreationEpochIs] instead.
func (r *Record) AddObjectCreationEpoch(m Match, epoch uint64) {
	r.SetFilters(append(r.Filters(), NewFilterObjectCreationEpochIs(m, epoch)))
}

// AddObjectPayloadLengthFilter adds filter by object payload length.
//
// The m must not be [MatchNotPresent].
// Deprecated: use [ConstructRecord] with [NewFilterObjectPayloadSizeIs] instead.
func (r *Record) AddObjectPayloadLengthFilter(m Match, size uint64) {
	r.SetFilters(append(r.Filters(), NewFilterObjectPayloadSizeIs(m, size)))
}

// AddObjectPayloadHashFilter adds filter by object payload hash value.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] instead.
func (r *Record) AddObjectPayloadHashFilter(m Match, h checksum.Checksum) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectPayloadChecksum, m, h.String())))
}

// AddObjectTypeFilter adds filter by object type.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] instead.
func (r *Record) AddObjectTypeFilter(m Match, t object.Type) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectType, m, object.TypeToString(t))))
}

// AddObjectHomomorphicHashFilter adds filter by object payload homomorphic hash value.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
// Deprecated: use [ConstructRecord] with [NewObjectPropertyFilter] instead.
func (r *Record) AddObjectHomomorphicHashFilter(m Match, h checksum.Checksum) {
	r.SetFilters(append(r.Filters(), NewObjectPropertyFilter(FilterObjectPayloadHomomorphicChecksum, m, h.String())))
}

// ToV2 converts Record to v2 acl.EACLRecord message.
//
// Nil Record converts to nil.
// Deprecated: do not use it.
func (r *Record) ToV2() *v2acl.Record {
	if r != nil {
		return r.toProtoMessage()
	}
	return nil
}

func (r Record) toProtoMessage() *v2acl.Record {
	v2 := new(v2acl.Record)

	if r.targets != nil {
		targets := make([]v2acl.Target, len(r.targets))
		for i := range r.targets {
			targets[i] = *r.targets[i].toProtoMessage()
		}

		v2.SetTargets(targets)
	}

	if r.filters != nil {
		filters := make([]v2acl.HeaderFilter, len(r.filters))
		for i := range r.filters {
			filters[i] = *r.filters[i].toProtoMessage()
		}

		v2.SetFilters(filters)
	}

	v2.SetAction(v2acl.Action(r.action))
	v2.SetOperation(v2acl.Operation(r.operation))

	return v2
}

func (r *Record) fromProtoMessage(m *v2acl.Record) error {
	mt := m.GetTargets()
	r.targets = make([]Target, len(mt))
	for i := range mt {
		if err := r.targets[i].fromProtoMessage(&mt[i]); err != nil {
			return fmt.Errorf("invalid subject descriptor #%d: %w", i, err)
		}
	}

	mf := m.GetFilters()
	r.filters = make([]Filter, len(mf))
	for i := range mf {
		if err := r.filters[i].fromProtoMessage(&mf[i]); err != nil {
			return fmt.Errorf("invalid filter #%d: %w", i, err)
		}
	}

	r.action = Action(m.GetAction())
	r.operation = Operation(m.GetOperation())

	return nil
}

// NewRecord creates and returns blank Record instance.
//
// Defaults:
//   - action: ActionUnspecified;
//   - operation: OperationUnspecified;
//   - targets: nil,
//   - filters: nil.
//
// Deprecated: use [ConstructRecord] instead.
func NewRecord() *Record {
	return new(Record)
}

// CreateRecord creates, initializes with parameters and returns Record instance.
// Deprecated: use [ConstructRecord] instead.
func CreateRecord(action Action, operation Operation) *Record {
	r := NewRecord()
	r.action = action
	r.operation = operation
	r.targets = []Target{}
	r.filters = []Filter{}

	return r
}

// NewRecordFromV2 converts v2 acl.EACLRecord message to Record.
// Deprecated: do not use it.
func NewRecordFromV2(record *v2acl.Record) *Record {
	r := NewRecord()

	if record == nil {
		return r
	}

	_ = r.fromProtoMessage(record)
	return r
}

// Marshal marshals Record into a protobuf binary form.
func (r Record) Marshal() []byte {
	return r.toProtoMessage().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Record.
func (r *Record) Unmarshal(data []byte) error {
	m := new(v2acl.Record)
	if err := m.Unmarshal(data); err != nil {
		return err
	}
	return r.fromProtoMessage(m)
}

// MarshalJSON encodes Record to protobuf JSON format.
func (r Record) MarshalJSON() ([]byte, error) {
	return r.toProtoMessage().MarshalJSON()
}

// UnmarshalJSON decodes Record from protobuf JSON format.
func (r *Record) UnmarshalJSON(data []byte) error {
	m := new(v2acl.Record)
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	return r.fromProtoMessage(m)
}
