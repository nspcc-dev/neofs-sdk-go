package audit

import (
	"github.com/nspcc-dev/neofs-api-go/v2/audit"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Result represents report on the results of the data audit in NeoFS system.
//
// Result is mutually binary-compatible with github.com/nspcc-dev/neofs-api-go/v2/audit.DataAuditResult
// message. See Marshal / Unmarshal methods.
//
// Instances can be created using built-in var declaration.
type Result struct {
	versionEncoded bool

	v2 audit.DataAuditResult
}

// Marshal encodes Result into a canonical NeoFS binary format (Protocol Buffers
// with direct field order).
//
// Writes version.Current() protocol version into the resulting message if Result
// hasn't been already decoded from such a message using Unmarshal.
//
// See also Unmarshal.
func (r *Result) Marshal() []byte {
	if !r.versionEncoded {
		r.v2.SetVersion(version.Current().ToV2())
		r.versionEncoded = true
	}

	data, err := r.v2.StableMarshal(nil)
	if err != nil {
		panic(err)
	}

	return data
}

// Unmarshal decodes Result from its canonical NeoFS binary format (Protocol Buffers
// with direct field order). Returns an error describing a format violation.
//
// See also Marshal.
func (r *Result) Unmarshal(data []byte) error {
	err := r.v2.Unmarshal(data)
	if err == nil {
		r.versionEncoded = true
	}

	return err
}

// Epoch returns NeoFS epoch when the data associated with the Result was audited.
//
// Zero Result has zero epoch.
//
// See also ForEpoch.
func (r Result) Epoch() uint64 {
	return r.v2.GetAuditEpoch()
}

// ForEpoch specifies NeoFS epoch when the data associated with the Result was audited.
//
// See also Epoch.
func (r *Result) ForEpoch(epoch uint64) {
	r.v2.SetAuditEpoch(epoch)
}

// Container returns identifier of the container with which the data audit Result
// is associated.
//
// Returns nil if container is not specified. Zero Result has nil container.
// Return value MUST NOT be mutated: to do this, first make a copy.
//
// See also ForContainer.
func (r Result) Container() *cid.ID {
	return cid.NewFromV2(r.v2.GetContainerID())
}

// ForContainer returns identifier of the container with which the data audit Result
// is associated.
//
// See also Container.
func (r *Result) ForContainer(cnr cid.ID) {
	r.v2.SetContainerID(cnr.ToV2())
}

// AuditorKey returns public key of the auditing NeoFS Inner Ring node in
// a NeoFS binary key format.
//
// Zero Result has nil key. Return value MUST NOT be mutated: to do this,
// first make a copy.
//
// See also SetAuditorPublicKey.
func (r Result) AuditorKey() []byte {
	return r.v2.GetPublicKey()
}

// SetAuditorKey specifies public key of the auditing NeoFS Inner Ring node in
// a NeoFS binary key format.
//
// Argument MUST NOT be mutated at least until the end of using the Result.
//
// See also AuditorKey.
func (r *Result) SetAuditorKey(key []byte) {
	r.v2.SetPublicKey(key)
}

// Completed returns completion state of the data audit associated with the Result.
//
// Zero Result corresponds to incomplete data audit.
//
// See also Complete.
func (r Result) Completed() bool {
	return r.v2.GetComplete()
}

// Complete marks the data audit associated with the Result as completed.
//
// See also Completed.
func (r *Result) Complete() {
	r.v2.SetComplete(true)
}

// RequestsPoR returns number of requests made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// Zero Result has zero requests.
//
// See also SetRequestsPoR.
func (r Result) RequestsPoR() uint32 {
	return r.v2.GetRequests()
}

// SetRequestsPoR sets number of requests made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// See also RequestsPoR.
func (r *Result) SetRequestsPoR(v uint32) {
	r.v2.SetRequests(v)
}

// RetriesPoR returns number of retries made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// Zero Result has zero retries.
//
// See also SetRetriesPoR.
func (r Result) RetriesPoR() uint32 {
	return r.v2.GetRetries()
}

// SetRetriesPoR sets number of retries made by Proof-of-Retrievability
// audit check to get all headers of the objects inside storage groups.
//
// See also RetriesPoR.
func (r *Result) SetRetriesPoR(v uint32) {
	r.v2.SetRetries(v)
}

// IteratePassedStorageGroups iterates over all storage groups that passed
// Proof-of-Retrievability audit check and passes them into f. Breaks on f's
// false return, f MUST NOT be nil.
//
// Zero Result has no passed storage groups and doesn't call f.
//
// See also SubmitPassedStorageGroup.
func (r Result) IteratePassedStorageGroups(f func(oid.ID) bool) {
	r2 := r.v2.GetPassSG()

	for i := range r2 {
		if !f(*oid.NewIDFromV2(&r2[i])) {
			return
		}
	}
}

// SubmitPassedStorageGroup marks storage group as passed Proof-of-Retrievability
// audit check.
//
// See also IteratePassedStorageGroups.
func (r *Result) SubmitPassedStorageGroup(sg oid.ID) {
	r.v2.SetPassSG(append(r.v2.GetPassSG(), *sg.ToV2()))
}

// IterateFailedStorageGroups is similar to IteratePassedStorageGroups but for failed groups.
//
// See also SubmitFailedStorageGroup.
func (r Result) IterateFailedStorageGroups(f func(oid.ID) bool) {
	v := r.v2.GetFailSG()

	for i := range v {
		if !f(*oid.NewIDFromV2(&v[i])) {
			return
		}
	}
}

// SubmitFailedStorageGroup is similar to SubmitPassedStorageGroup but for failed groups.
//
// See also IterateFailedStorageGroups.
func (r *Result) SubmitFailedStorageGroup(sg oid.ID) {
	r.v2.SetFailSG(append(r.v2.GetFailSG(), *sg.ToV2()))
}

// Hits returns number of sampled objects under audit placed
// in an optimal way according to the container's placement policy
// when checking Proof-of-Placement.
//
// Zero result has zero hits.
//
// See also SetHits.
func (r Result) Hits() uint32 {
	return r.v2.GetHit()
}

// SetHits sets number of sampled objects under audit placed
// in an optimal way according to the containers placement policy
// when checking Proof-of-Placement.
//
// See also Hits.
func (r *Result) SetHits(hit uint32) {
	r.v2.SetHit(hit)
}

// Misses returns number of sampled objects under audit placed
// in suboptimal way according to the container's placement policy,
// but still at a satisfactory level when checking Proof-of-Placement.
//
// Zero Result has zero misses.
//
// See also SetMisses.
func (r Result) Misses() uint32 {
	return r.v2.GetMiss()
}

// SetMisses sets number of sampled objects under audit placed
// in suboptimal way according to the container's placement policy,
// but still at a satisfactory level when checking Proof-of-Placement.
//
// See also Misses.
func (r *Result) SetMisses(miss uint32) {
	r.v2.SetMiss(miss)
}

// Failures returns number of sampled objects under audit stored
// in a way not confirming placement policy or not found at all
// when checking Proof-of-Placement.
//
// Zero result has zero failures.
//
// See also SetFailures.
func (r Result) Failures() uint32 {
	return r.v2.GetFail()
}

// SetFailures sets number of sampled objects under audit stored
// in a way not confirming placement policy or not found at all
// when checking Proof-of-Placement.
//
// See also Failures.
func (r *Result) SetFailures(fail uint32) {
	r.v2.SetFail(fail)
}

// IteratePassedStorageNodes iterates over all storage nodes that passed at least one
// Proof-of-Data-Possession audit check and passes their public keys into f. Breaks on
// f's false return.
//
// f MUST NOT be nil and MUST NOT mutate parameter passed into it at least until
// the end of using the Result.
//
// Zero Result has no passed storage nodes and doesn't call f.
//
// See also SubmitPassedStorageNode.
func (r Result) IteratePassedStorageNodes(f func([]byte) bool) {
	v := r.v2.GetPassNodes()

	for i := range v {
		if !f(v[i]) {
			return
		}
	}
}

// SubmitPassedStorageNodes marks storage node list as passed Proof-of-Data-Possession
// audit check. The list contains public keys.
//
// Argument and its elements MUST NOT be mutated at least until the end of using the Result.
//
// See also IteratePassedStorageNodes.
func (r *Result) SubmitPassedStorageNodes(list [][]byte) {
	r.v2.SetPassNodes(list)
}

// IterateFailedStorageNodes is similar to IteratePassedStorageNodes but for failed nodes.
//
// See also SubmitPassedStorageNodes.
func (r Result) IterateFailedStorageNodes(f func([]byte) bool) {
	v := r.v2.GetFailNodes()

	for i := range v {
		if !f(v[i]) {
			return
		}
	}
}

// SubmitFailedStorageNodes is similar to SubmitPassedStorageNodes but for failed nodes.
//
// See also IterateFailedStorageNodes.
func (r *Result) SubmitFailedStorageNodes(list [][]byte) {
	r.v2.SetFailNodes(list)
}
