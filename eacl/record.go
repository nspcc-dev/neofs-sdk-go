package eacl

import (
	"crypto/ecdsa"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Record of the ContainerEACL rule, that defines ContainerEACL action, targets for this action,
// object service operation and filters for request headers.
//
// Record is compatible with v2 acl.EACLRecord message.
type Record struct {
	action    Action
	operation Operation
	filters   []Filter
	targets   []Target
}

// CopyTo writes deep copy of the [Record] to dst.
func (r Record) CopyTo(dst *Record) {
	dst.action = r.action
	dst.operation = r.operation

	dst.filters = make([]Filter, len(r.filters))
	copy(dst.filters, r.filters)

	dst.targets = make([]Target, len(r.targets))
	for i, t := range r.targets {
		var newTarget Target
		t.CopyTo(&newTarget)

		dst.targets[i] = newTarget
	}
}

// Targets returns list of target subjects to apply ACL rule to.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (r Record) Targets() []Target {
	return r.targets
}

// SetTargets sets list of target subjects to apply ACL rule to.
func (r *Record) SetTargets(targets ...Target) {
	r.targets = targets
}

// Filters returns list of filters to match and see if rule is applicable.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (r Record) Filters() []Filter {
	return r.filters
}

// Operation returns NeoFS request verb to match.
func (r Record) Operation() Operation {
	return r.operation
}

// SetOperation sets NeoFS request verb to match.
func (r *Record) SetOperation(operation Operation) {
	r.operation = operation
}

// Action returns rule execution result.
func (r Record) Action() Action {
	return r.action
}

// SetAction sets rule execution result.
func (r *Record) SetAction(action Action) {
	r.action = action
}

// AddRecordTarget adds single Target to the Record.
func AddRecordTarget(r *Record, t *Target) {
	r.SetTargets(append(r.Targets(), *t)...)
}

// AddFormedTarget forms Target with specified Role and list of
// ECDSA public keys and adds it to the Record.
func AddFormedTarget(r *Record, role Role, keys ...ecdsa.PublicKey) {
	t := NewTarget()
	t.SetRole(role)

	SetTargetECDSAKeys(t, ecdsaKeysToPtrs(keys)...)
	AddRecordTarget(r, t)
}

type stringEncoder interface {
	EncodeToString() string
}

func (r *Record) addFilter(from FilterHeaderType, m Match, key string, val stringEncoder) {
	filter := Filter{
		from:    from,
		key:     key,
		matcher: m,
		value:   val,
	}

	r.filters = append(r.filters, filter)
}

func (r *Record) addObjectFilter(m Match, key string, val stringEncoder) {
	r.addFilter(HeaderFromObject, m, key, val)
}

// AddFilter adds generic filter.
//
// If matcher is [MatchNotPresent], the value must be empty. If matcher is
// numeric (e.g. [MatchNumGT]), value must be a base-10 integer.
func (r *Record) AddFilter(from FilterHeaderType, matcher Match, name, value string) {
	r.addFilter(from, matcher, name, staticStringer(value))
}

// AddObjectAttributeFilter adds filter by object attribute.
//
// If m is [MatchNotPresent], the value must be empty. If matcher is numeric
// (e.g. [MatchNumGT]), value must be a base-10 integer.
func (r *Record) AddObjectAttributeFilter(m Match, key, value string) {
	r.addObjectFilter(m, key, staticStringer(value))
}

// AddObjectVersionFilter adds filter by object version.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectVersionFilter(m Match, v *version.Version) {
	r.addObjectFilter(m, FilterObjectVersion, staticStringer(version.EncodeToString(*v)))
}

// AddObjectIDFilter adds filter by object ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectIDFilter(m Match, id oid.ID) {
	r.addObjectFilter(m, FilterObjectID, id)
}

// AddObjectContainerIDFilter adds filter by object container ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectContainerIDFilter(m Match, id cid.ID) {
	r.addObjectFilter(m, FilterObjectContainerID, id)
}

// AddObjectOwnerIDFilter adds filter by object owner ID.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectOwnerIDFilter(m Match, id *user.ID) {
	r.addObjectFilter(m, FilterObjectOwnerID, id)
}

// AddObjectCreationEpoch adds filter by object creation epoch.
//
// The m must not be [MatchNotPresent].
func (r *Record) AddObjectCreationEpoch(m Match, epoch uint64) {
	r.addObjectFilter(m, FilterObjectCreationEpoch, u64Stringer(epoch))
}

// AddObjectPayloadLengthFilter adds filter by object payload length.
//
// The m must not be [MatchNotPresent].
func (r *Record) AddObjectPayloadLengthFilter(m Match, size uint64) {
	r.addObjectFilter(m, FilterObjectPayloadSize, u64Stringer(size))
}

// AddObjectPayloadHashFilter adds filter by object payload hash value.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectPayloadHashFilter(m Match, h checksum.Checksum) {
	r.addObjectFilter(m, FilterObjectPayloadChecksum, staticStringer(h.String()))
}

// AddObjectTypeFilter adds filter by object type.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectTypeFilter(m Match, t object.Type) {
	r.addObjectFilter(m, FilterObjectType, staticStringer(t.EncodeToString()))
}

// AddObjectHomomorphicHashFilter adds filter by object payload homomorphic hash value.
//
// The m must not be [MatchNotPresent] or numeric (e.g. [MatchNumGT]).
func (r *Record) AddObjectHomomorphicHashFilter(m Match, h checksum.Checksum) {
	r.addObjectFilter(m, FilterObjectPayloadHomomorphicChecksum, staticStringer(h.String()))
}

// ToV2 converts Record to v2 acl.EACLRecord message.
//
// Nil Record converts to nil.
func (r *Record) ToV2() *v2acl.Record {
	if r == nil {
		return nil
	}

	v2 := new(v2acl.Record)

	if r.targets != nil {
		targets := make([]v2acl.Target, len(r.targets))
		for i := range r.targets {
			targets[i] = *r.targets[i].ToV2()
		}

		v2.SetTargets(targets)
	}

	if r.filters != nil {
		filters := make([]v2acl.HeaderFilter, len(r.filters))
		for i := range r.filters {
			filters[i] = *r.filters[i].ToV2()
		}

		v2.SetFilters(filters)
	}

	v2.SetAction(r.action.ToV2())
	v2.SetOperation(r.operation.ToV2())

	return v2
}

// NewRecord creates and returns blank Record instance.
//
// Defaults:
//   - action: ActionUnknown;
//   - operation: OperationUnknown;
//   - targets: nil,
//   - filters: nil.
func NewRecord() *Record {
	return new(Record)
}

// CreateRecord creates, initializes with parameters and returns Record instance.
func CreateRecord(action Action, operation Operation) *Record {
	r := NewRecord()
	r.action = action
	r.operation = operation
	r.targets = []Target{}
	r.filters = []Filter{}

	return r
}

// NewRecordFromV2 converts v2 acl.EACLRecord message to Record.
func NewRecordFromV2(record *v2acl.Record) *Record {
	r := NewRecord()

	if record == nil {
		return r
	}

	r.action = ActionFromV2(record.GetAction())
	r.operation = OperationFromV2(record.GetOperation())

	v2targets := record.GetTargets()
	v2filters := record.GetFilters()

	r.targets = make([]Target, len(v2targets))
	for i := range v2targets {
		r.targets[i] = *NewTargetFromV2(&v2targets[i])
	}

	r.filters = make([]Filter, len(v2filters))
	for i := range v2filters {
		r.filters[i] = *NewFilterFromV2(&v2filters[i])
	}

	return r
}

// Marshal marshals Record into a protobuf binary form.
func (r *Record) Marshal() ([]byte, error) {
	return r.ToV2().StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of Record.
func (r *Record) Unmarshal(data []byte) error {
	fV2 := new(v2acl.Record)
	if err := fV2.Unmarshal(data); err != nil {
		return err
	}

	*r = *NewRecordFromV2(fV2)

	return nil
}

// MarshalJSON encodes Record to protobuf JSON format.
func (r *Record) MarshalJSON() ([]byte, error) {
	return r.ToV2().MarshalJSON()
}

// UnmarshalJSON decodes Record from protobuf JSON format.
func (r *Record) UnmarshalJSON(data []byte) error {
	tV2 := new(v2acl.Record)
	if err := tV2.UnmarshalJSON(data); err != nil {
		return err
	}

	*r = *NewRecordFromV2(tV2)

	return nil
}

// equalRecords compares Record with each other.
func equalRecords(r1, r2 Record) bool {
	if r1.Operation() != r2.Operation() ||
		r1.Action() != r2.Action() {
		return false
	}

	fs1, fs2 := r1.Filters(), r2.Filters()
	ts1, ts2 := r1.Targets(), r2.Targets()

	if len(fs1) != len(fs2) ||
		len(ts1) != len(ts2) {
		return false
	}

	for i := 0; i < len(fs1); i++ {
		if !equalFilters(fs1[i], fs2[i]) {
			return false
		}
	}

	for i := 0; i < len(ts1); i++ {
		if !equalTargets(ts1[i], ts2[i]) {
			return false
		}
	}

	return true
}
