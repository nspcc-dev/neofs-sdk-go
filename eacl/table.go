package eacl

import (
	"bytes"
	"fmt"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Table is a group of ContainerEACL records for single container.
//
// Table is compatible with v2 acl.EACLTable message.
type Table struct {
	version version.Version
	cid     cid.ID
	records []Record
}

// CopyTo writes deep copy of the [Table] to dst.
func (t Table) CopyTo(dst *Table) {
	ver := t.version
	dst.version = ver
	dst.cid = t.cid

	dst.records = make([]Record, len(t.records))
	for i := range t.records {
		t.records[i].CopyTo(&dst.records[i])
	}
}

// CID returns identifier of the container that should use given access control rules.
// Deprecated: use [Table.GetCID] instead.
func (t Table) CID() (cid.ID, bool) { return t.cid, !t.cid.IsZero() }

// GetCID returns identifier of the NeoFS container to which the eACL scope is
// limited. Zero return means the eACL may be applied to any container.
func (t Table) GetCID() cid.ID { return t.cid }

// SetCID limits scope of the eACL to a referenced container. By default, if ID
// is zero, the eACL is applicable to any container.
func (t *Table) SetCID(cid cid.ID) {
	t.cid = cid
}

// Version returns version of eACL format.
func (t Table) Version() version.Version {
	return t.version
}

// SetVersion sets version of eACL format.
func (t *Table) SetVersion(version version.Version) {
	t.version = version
}

// Records returns list of extended ACL rules.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (t Table) Records() []Record {
	return t.records
}

// AddRecord adds single eACL rule.
func (t *Table) AddRecord(r *Record) {
	if r != nil {
		t.records = append(t.records, *r)
	}
}

// ReadFromV2 reads Table from the [v2acl.Table] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Table.ToV2].
func (t *Table) ReadFromV2(m v2acl.Table) error {
	// set container id
	if id := m.GetContainerID(); id != nil {
		if err := t.cid.ReadFromV2(*id); err != nil {
			return fmt.Errorf("invalid container ID: %w", err)
		}
	} else {
		t.cid = cid.ID{}
	}

	// set version
	if v := m.GetVersion(); v != nil {
		ver := version.Version{}
		ver.SetMajor(v.GetMajor())
		ver.SetMinor(v.GetMinor())

		t.SetVersion(ver)
	}

	// set eacl records
	v2records := m.GetRecords()
	t.records = make([]Record, len(v2records))

	for i := range v2records {
		if err := t.records[i].fromProtoMessage(&v2records[i]); err != nil {
			return fmt.Errorf("invalid record #%d: %w", i, err)
		}
	}

	return nil
}

// ToV2 converts Table to v2 acl.EACLTable message.
//
// Nil Table converts to nil.
//
// See also [Table.ReadFromV2].
func (t *Table) ToV2() *v2acl.Table {
	if t == nil {
		return nil
	}

	v2 := new(v2acl.Table)
	var cidV2 refs.ContainerID

	if !t.cid.IsZero() {
		t.cid.WriteToV2(&cidV2)
		v2.SetContainerID(&cidV2)
	}

	if t.records != nil {
		records := make([]v2acl.Record, len(t.records))
		for i := range t.records {
			records[i] = *t.records[i].toProtoMessage()
		}

		v2.SetRecords(records)
	}

	var verV2 refs.Version
	t.version.WriteToV2(&verV2)
	v2.SetVersion(&verV2)

	return v2
}

// NewTable creates, initializes and returns blank Table instance.
//
// Defaults:
//   - version: version.Current();
//   - container ID: nil;
//   - records: nil;
//   - session token: nil;
//   - signature: nil.
func NewTable() *Table {
	t := new(Table)
	t.SetVersion(version.Current())

	return t
}

// CreateTable creates, initializes with parameters and returns Table instance.
func CreateTable(cid cid.ID) *Table {
	t := NewTable()
	t.SetCID(cid)

	return t
}

// NewTableFromV2 converts v2 acl.EACLTable message to Table.
//
// Deprecated: BUG: container ID length is not checked. Use [Table.ReadFromV2]
// instead.
func NewTableFromV2(table *v2acl.Table) *Table {
	t := new(Table)

	if table == nil {
		return t
	}

	// set version
	if v := table.GetVersion(); v != nil {
		ver := version.Version{}
		ver.SetMajor(v.GetMajor())
		ver.SetMinor(v.GetMinor())

		t.SetVersion(ver)
	}

	// set container id
	if id := table.GetContainerID(); id != nil {
		copy(t.cid[:], id.GetValue())
	}

	// set eacl records
	v2records := table.GetRecords()
	t.records = make([]Record, len(v2records))

	for i := range v2records {
		_ = t.records[i].fromProtoMessage(&v2records[i])
	}

	return t
}

// Marshal marshals Table into a protobuf binary form.
func (t *Table) Marshal() []byte {
	return t.ToV2().StableMarshal(nil)
}

// SignedData returns actual payload to sign.
//
// See also [client.Client.ContainerSetEACL].
func (t Table) SignedData() []byte {
	return t.Marshal()
}

// Unmarshal unmarshals protobuf binary representation of Table.
func (t *Table) Unmarshal(data []byte) error {
	var m v2acl.Table
	if err := m.Unmarshal(data); err != nil {
		return err
	}
	return t.ReadFromV2(m)
}

// MarshalJSON encodes Table to protobuf JSON format.
func (t *Table) MarshalJSON() ([]byte, error) {
	return t.ToV2().MarshalJSON()
}

// UnmarshalJSON decodes Table from protobuf JSON format.
func (t *Table) UnmarshalJSON(data []byte) error {
	var m v2acl.Table
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	return t.ReadFromV2(m)
}

// EqualTables compares Table with each other.
// Deprecated: compare [Table.Marshal] instead.
func EqualTables(t1, t2 Table) bool { return bytes.Equal(t1.Marshal(), t2.Marshal()) }
