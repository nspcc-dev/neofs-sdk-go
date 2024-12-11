package eacl

import (
	"bytes"
	"fmt"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

var zeroVersion version.Version

// Table is a group of ContainerEACL records for single container.
//
// Table is compatible with v2 [protoacl.EACLTable] message.
//
// Table should be created using one of the constructors.
type Table struct {
	version version.Version
	cid     cid.ID
	records []Record
}

// ConstructTable constructs new Table with given records. Use
// [NewTableForContainer] to limit the NeoFS container. The rs must not be
// empty.
func ConstructTable(rs []Record) Table {
	return Table{version: version.Current(), records: rs}
}

// NewTableForContainer constructs new Table with given records which apply only
// to the specified NeoFS container. The rs must not be empty.
func NewTableForContainer(cnr cid.ID, rs []Record) Table {
	t := ConstructTable(rs)
	t.SetCID(cnr)
	return t
}

// Unmarshal creates new Table and makes [Table.Unmarshal].
func Unmarshal(b []byte) (Table, error) {
	var t Table
	return t, t.Unmarshal(b)
}

// UnmarshalJSON creates new Table and makes [Table.UnmarshalJSON].
func UnmarshalJSON(b []byte) (Table, error) {
	var t Table
	return t, t.UnmarshalJSON(b)
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

// SetRecords sets list of extended ACL rules.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (t *Table) SetRecords(rs []Record) {
	t.records = rs
}

// AddRecord adds single eACL rule.
// Deprecated: use [Table.SetRecords] instead.
func (t *Table) AddRecord(r *Record) {
	if r != nil {
		t.records = append(t.records, *r)
	}
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// t from it.
//
// See also [Table.ProtoMessage].
func (t *Table) FromProtoMessage(m *protoacl.EACLTable) error {
	// set container id
	if m.ContainerId != nil {
		if err := t.cid.FromProtoMessage(m.ContainerId); err != nil {
			return fmt.Errorf("invalid container ID: %w", err)
		}
	} else {
		t.cid = cid.ID{}
	}

	// set version
	if m.Version != nil {
		if err := t.version.FromProtoMessage(m.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	// set eacl records
	rs := m.Records
	t.records = make([]Record, len(rs))

	for i := range rs {
		if rs[i] == nil {
			return fmt.Errorf("nil record #%d", i)
		}
		if err := t.records[i].fromProtoMessage(rs[i]); err != nil {
			return fmt.Errorf("invalid record #%d: %w", i, err)
		}
	}

	return nil
}

// ProtoMessage converts t into message to transmit using the NeoFS API
// protocol.
//
// See also [Table.FromProtoMessage].
func (t Table) ProtoMessage() *protoacl.EACLTable {
	m := new(protoacl.EACLTable)
	if !t.cid.IsZero() {
		m.ContainerId = t.cid.ProtoMessage()
	}

	if t.records != nil {
		m.Records = make([]*protoacl.EACLRecord, len(t.records))
		for i := range t.records {
			m.Records[i] = t.records[i].toProtoMessage()
		}
	}

	m.Version = t.version.ProtoMessage()

	return m
}

// NewTable creates, initializes and returns blank Table instance.
//
// Defaults:
//   - version: version.Current();
//   - container ID: nil;
//   - records: nil.
//
// Deprecated: use [ConstructTable] instead.
func NewTable() *Table {
	t := ConstructTable(nil)
	return &t
}

// CreateTable creates, initializes with parameters and returns Table instance.
// Deprecated: use [NewTableForContainer] instead.
func CreateTable(cid cid.ID) *Table {
	t := NewTableForContainer(cid, nil)
	return &t
}

// Marshal marshals Table into a protobuf binary form.
func (t Table) Marshal() []byte {
	return neofsproto.Marshal(t)
}

// SignedData returns actual payload to sign.
//
// See also [client.Client.ContainerSetEACL].
func (t Table) SignedData() []byte {
	return t.Marshal()
}

// Unmarshal unmarshals protobuf binary representation of Table. Use [Unmarshal]
// to decode data into a new Table.
func (t *Table) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, t)
}

// MarshalJSON encodes Table to protobuf JSON format.
func (t Table) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(t)
}

// UnmarshalJSON decodes Table from protobuf JSON format. Use [UnmarshalJSON] to
// decode data into a new Table.
func (t *Table) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSON(data, t)
}

// EqualTables compares Table with each other.
// Deprecated: compare [Table.Marshal] instead.
func EqualTables(t1, t2 Table) bool { return bytes.Equal(t1.Marshal(), t2.Marshal()) }

// IsZero checks whether all fields of the table are zero/empty. The property
// can be used as a marker of unset eACL.
func (t Table) IsZero() bool {
	return t.cid.IsZero() && len(t.records) == 0 && t.version == zeroVersion
}
