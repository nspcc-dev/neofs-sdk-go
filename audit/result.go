package audit

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/audit"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/proto"
)

// Result represents report on the results of the data audit in NeoFS system.
//
// Instances can be created using built-in var declaration.
type Result struct {
	decoded bool

	versionSet bool
	version    version.Version

	auditEpoch    uint64
	auditorPubKey []byte

	cnrSet bool
	cnr    cid.ID

	completed bool

	requestsPoR, retriesPoR uint32

	hits, misses, fails uint32

	passSG, failSG []oid.ID

	passNodes, failNodes [][]byte
}

// Marshal encodes Result into a Protocol Buffers V3 binary format.
//
// Writes current protocol version into the resulting message if Result hasn't
// been already decoded from such a message.
//
// See also [Result.Unmarshal].
func (r Result) Marshal() []byte {
	m := &audit.DataAuditResult{
		AuditEpoch: r.auditEpoch,
		PublicKey:  r.auditorPubKey,
		Complete:   r.completed,
		Requests:   r.requestsPoR,
		Retries:    r.retriesPoR,
		Hit:        r.hits,
		Miss:       r.misses,
		Fail:       r.fails,
		PassNodes:  r.passNodes,
		FailNodes:  r.failNodes,
	}
	if r.versionSet {
		m.Version = new(refs.Version)
		r.version.WriteToV2(m.Version)
	} else if !r.decoded {
		m.Version = new(refs.Version)
		version.Current().WriteToV2(m.Version)
	}
	if r.cnrSet {
		m.ContainerId = new(refs.ContainerID)
		r.cnr.WriteToV2(m.ContainerId)
	}
	if r.passSG != nil {
		m.PassSg = make([]*refs.ObjectID, len(r.passSG))
		for i := range r.passSG {
			m.PassSg[i] = new(refs.ObjectID)
			r.passSG[i].WriteToV2(m.PassSg[i])
		}
	}
	if r.failSG != nil {
		m.FailSg = make([]*refs.ObjectID, len(r.failSG))
		for i := range r.failSG {
			m.FailSg[i] = new(refs.ObjectID)
			r.failSG[i].WriteToV2(m.FailSg[i])
		}
	}

	b, err := proto.Marshal(m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

var errCIDNotSet = errors.New("container ID is not set")

// Unmarshal decodes Protocol Buffers V3 binary data into the Result. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Result.Marshal].
func (r *Result) Unmarshal(data []byte) error {
	var m audit.DataAuditResult
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	// format checks
	r.cnrSet = m.ContainerId != nil
	if !r.cnrSet {
		return errCIDNotSet
	}

	err = r.cnr.ReadFromV2(m.ContainerId)
	if err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	r.versionSet = m.Version != nil
	if r.versionSet {
		err = r.version.ReadFromV2(m.Version)
		if err != nil {
			return fmt.Errorf("invalid protocol version: %w", err)
		}
	}

	if m.PassSg != nil {
		r.passSG = make([]oid.ID, len(m.PassSg))
		for i := range m.PassSg {
			err = r.passSG[i].ReadFromV2(m.PassSg[i])
			if err != nil {
				return fmt.Errorf("invalid passed storage group ID #%d: %w", i, err)
			}
		}
	} else {
		r.passSG = nil
	}

	if m.FailSg != nil {
		r.failSG = make([]oid.ID, len(m.FailSg))
		for i := range m.FailSg {
			err = r.failSG[i].ReadFromV2(m.FailSg[i])
			if err != nil {
				return fmt.Errorf("invalid failed storage group ID #%d: %w", i, err)
			}
		}
	} else {
		r.failSG = nil
	}

	r.auditEpoch = m.AuditEpoch
	r.auditorPubKey = m.PublicKey
	r.completed = m.Complete
	r.requestsPoR = m.Requests
	r.retriesPoR = m.Retries
	r.hits = m.Hit
	r.misses = m.Miss
	r.fails = m.Fail
	r.passNodes = m.PassNodes
	r.failNodes = m.FailNodes
	r.decoded = true

	return nil
}

// Epoch returns NeoFS epoch when the data associated with the Result was audited.
//
// Zero Result has zero epoch.
//
// See also [Result.ForEpoch].
func (r Result) Epoch() uint64 {
	return r.auditEpoch
}

// ForEpoch specifies NeoFS epoch when the data associated with the Result was audited.
//
// See also [Result.Epoch].
func (r *Result) ForEpoch(epoch uint64) {
	r.auditEpoch = epoch
}

// Container returns identifier of the container with which the data audit Result
// is associated and a bool that indicates container ID field presence in the Result.
//
// Zero Result does not have container ID.
//
// See also [Result.ForContainer].
func (r Result) Container() (cid.ID, bool) {
	return r.cnr, r.cnrSet
}

// ForContainer sets identifier of the container with which the data audit Result
// is associated.
//
// See also [Result.Container].
func (r *Result) ForContainer(cnr cid.ID) {
	r.cnr, r.cnrSet = cnr, true
}

// AuditorKey returns public key of the auditing NeoFS Inner Ring node in
// a NeoFS binary key format.
//
// Zero Result has nil key.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// See also [Result.SetAuditorKey].
func (r Result) AuditorKey() []byte {
	return r.auditorPubKey
}

// SetAuditorKey specifies public key of the auditing NeoFS Inner Ring node in
// a NeoFS binary key format.
//
// Argument MUST NOT be mutated at least until the end of using the Result.
//
// Parameter key is a serialized compressed public key. See [elliptic.MarshalCompressed].
//
// See also [Result.AuditorKey].
func (r *Result) SetAuditorKey(key []byte) {
	r.auditorPubKey = key
}

// Completed returns completion state of the data audit associated with the Result.
//
// Zero Result corresponds to incomplete data audit.
//
// See also [Result.SetCompleted].
func (r Result) Completed() bool {
	return r.completed
}

// SetCompleted sets data audit completion flag.
//
// See also [Result.SetCompleted].
func (r *Result) SetCompleted(completed bool) {
	r.completed = completed
}

// RequestsPoR returns number of requests made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// Zero Result has zero requests.
//
// See also [Result.SetRequestsPoR].
func (r Result) RequestsPoR() uint32 {
	return r.requestsPoR
}

// SetRequestsPoR sets number of requests made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// See also [Result.RequestsPoR].
func (r *Result) SetRequestsPoR(v uint32) {
	r.requestsPoR = v
}

// RetriesPoR returns number of retries made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// Zero Result has zero retries.
//
// See also [Result.SetRetriesPoR].
func (r Result) RetriesPoR() uint32 {
	return r.retriesPoR
}

// SetRetriesPoR sets number of retries made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// See also [Result.RetriesPoR].
func (r *Result) SetRetriesPoR(v uint32) {
	r.retriesPoR = v
}

// PassedStorageGroups returns storage groups that passed
// Proof-of-Retrievability audit check. Breaks on f's false return, f MUST NOT
// be nil.
//
// Zero Result has no passed storage groups and doesn't call f.
//
// Return value MUST NOT be mutated at least until the end of using the Result.
//
// See also [Result.SetPassedStorageGroups].
func (r Result) PassedStorageGroups() []oid.ID {
	return r.passSG
}

// SetPassedStorageGroups sets storage groups that passed Proof-of-Retrievability
// audit check.
//
// Argument MUST NOT be mutated at least until the end of using the Result.
//
// See also [Result.PassedStorageGroups].
func (r *Result) SetPassedStorageGroups(ids []oid.ID) {
	r.passSG = ids
}

// FailedStorageGroups is similar to [Result.PassedStorageGroups] but for failed groups.
//
// See also [Result.SetFailedStorageGroups].
func (r Result) FailedStorageGroups() []oid.ID {
	return r.failSG
}

// SetFailedStorageGroups is similar to [Result.PassedStorageGroups] but for failed groups.
//
// See also [Result.FailedStorageGroups].
func (r *Result) SetFailedStorageGroups(ids []oid.ID) {
	r.failSG = ids
}

// Hits returns number of sampled objects under audit placed
// in an optimal way according to the container's placement policy
// when checking Proof-of-Placement.
//
// Zero result has zero hits.
//
// See also [Result.SetHits].
func (r Result) Hits() uint32 {
	return r.hits
}

// SetHits sets number of sampled objects under audit placed
// in an optimal way according to the containers placement policy
// when checking Proof-of-Placement.
//
// See also [Result.Hits].
func (r *Result) SetHits(hits uint32) {
	r.hits = hits
}

// Misses returns number of sampled objects under audit placed
// in suboptimal way according to the container's placement policy,
// but still at a satisfactory level when checking Proof-of-Placement.
//
// Zero Result has zero misses.
//
// See also [Result.SetMisses].
func (r Result) Misses() uint32 {
	return r.misses
}

// SetMisses sets number of sampled objects under audit placed
// in suboptimal way according to the container's placement policy,
// but still at a satisfactory level when checking Proof-of-Placement.
//
// See also [Result.Misses].
func (r *Result) SetMisses(misses uint32) {
	r.misses = misses
}

// Failures returns number of sampled objects under audit stored
// in a way not confirming placement policy or not found at all
// when checking Proof-of-Placement.
//
// Zero result has zero failures.
//
// See also [Result.SetFailures].
func (r Result) Failures() uint32 {
	return r.fails
}

// SetFailures sets number of sampled objects under audit stored
// in a way not confirming placement policy or not found at all
// when checking Proof-of-Placement.
//
// See also [Result.Failures].
func (r *Result) SetFailures(fails uint32) {
	r.fails = fails
}

// PassedStorageNodes returns public keys of storage nodes that passed at least
// one Proof-of-Data-Possession audit check. Breaks on f's false return.
//
// f MUST NOT be nil and MUST NOT mutate parameter passed into it at least until
// the end of using the Result.
//
// Zero Result has no passed storage nodes and doesn't call f.
//
// See also [Result.SetPassedStorageNodes]].
func (r Result) PassedStorageNodes() [][]byte {
	return r.passNodes
}

// SetPassedStorageNodes sets public keys of storage nodes that passed at least
// one Proof-of-Data-Possession audit check.
//
// Argument and its elements MUST NOT be mutated at least until the end of using the Result.
//
// See also [Result.PassedStorageNodes].
func (r *Result) SetPassedStorageNodes(list [][]byte) {
	r.passNodes = list
}

// FailedStorageNodes is similar to [Result.PassedStorageNodes] but for
// failures.
//
// See also [Result.SetFailedStorageNodes].
func (r Result) FailedStorageNodes() [][]byte {
	return r.failNodes
}

// SetFailedStorageNodes is similar to [Result.SetPassedStorageNodes] but for
// failures.
//
// See also [Result.SetFailedStorageNodes].
func (r *Result) SetFailedStorageNodes(list [][]byte) {
	r.failNodes = list
}
