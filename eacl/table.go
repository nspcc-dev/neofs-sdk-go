package eacl

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Table represents NeoFS extended ACL (eACL): group of rules managing access to
// NeoFS resources in addition to the basic ACL.
//
// See also [acl.Basic].
type Table struct {
	version *version.Version
	cid     *cid.ID
	records []Record
}

// New constructs eACL from the given list of access rules. Being applied as
// part of access control to the NeoFS resources, the rules are matched
// according to the first hit: if a rule with a certain index is applicable,
// then any rule with a larger index is ignored. Thus, to increase the priority
// of a rule, place it before all adjacent ones (for example, at the beginning).
//
// The argument MUST be non-empty and MUST NOT be mutated within lifetime of the
// resulting Table. All records MUST be correctly constructed.
//
// See also [NewForContainer].
func New(records []Record) Table {
	if len(records) == 0 {
		panic("empty records")
	}

	for i := range records {
		if msg := records[i].validate(); msg != "" {
			panic(fmt.Sprintf("invalid record #%d: %s", i, msg))
		}
	}

	v := version.Current()
	return Table{
		version: &v,
		records: records,
	}
}

// NewForContainer constructs the eACL similar to [New] but also limits its
// scope by NeoFS container with the specified reference.
//
// See also [Table.LimitByContainer].
func NewForContainer(cnr cid.ID, records []Record) Table {
	res := New(records)
	res.LimitByContainer(cnr)
	return res
}

// CopyTo writes deep copy of the [Table] to dst.
func (t Table) CopyTo(dst *Table) {
	if t.version != nil {
		dst.version = new(version.Version)
		*dst.version = *t.version
	} else {
		dst.version = nil
	}

	if t.cid != nil {
		id := *t.cid
		dst.cid = &id
	} else {
		dst.cid = nil
	}

	if t.records != nil {
		dst.records = make([]Record, len(t.records))
		for i := range t.records {
			t.records[i].copyTo(&dst.records[i])
		}
	} else {
		dst.records = nil
	}
}

// LimitByContainer limits scope of the eACL to a given container.
// By default, the eACL is applicable to any container.
//
// See also [Table.Container], [NewForContainer].
func (t *Table) LimitByContainer(cnr cid.ID) {
	t.cid = &cnr
}

// Container returns identifier of the NeoFS container to which the eACL scope
// is limited. If container is not specified, second value is false meaning that
// eACL may be applied to any container.
//
// See also [Table.LimitByContainer].
func (t Table) Container() (cid.ID, bool) {
	if t.cid == nil {
		var zero cid.ID
		return zero, false
	}

	return *t.cid, true
}

// Records returns list of extended ACL rules.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Table.SetRecords].
func (t Table) Records() []Record {
	return t.records
}

// WriteToV2 writes Table to the [acl.Table] message of the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Table.ReadFromV2].
func (t Table) WriteToV2(m *acl.Table) {
	if t.version != nil {
		var ver refs.Version
		t.version.WriteToV2(&ver)

		m.SetVersion(&ver)
	} else {
		m.SetVersion(nil)
	}

	if t.cid != nil {
		var cnr refs.ContainerID
		t.cid.WriteToV2(&cnr)

		m.SetContainerID(&cnr)
	} else {
		m.SetContainerID(nil)
	}

	if t.records != nil {
		rs := make([]acl.Record, len(t.records))
		for i := range t.records {
			t.records[i].writeToV2(&rs[i])
		}

		m.SetRecords(rs)
	} else {
		m.SetRecords(nil)
	}
}

func (t *Table) readFromV2(m acl.Table, checkFieldPresence bool) error {
	var err error

	ver := m.GetVersion()
	if ver != nil {
		t.version = new(version.Version)

		err = t.version.ReadFromV2(*ver)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else {
		if checkFieldPresence {
			return errors.New("missing version")
		}

		t.version = nil
	}

	cnr := m.GetContainerID()
	if cnr != nil {
		t.cid = new(cid.ID)

		err = t.cid.ReadFromV2(*cnr)
		if err != nil {
			return fmt.Errorf("invalid container reference: %w", err)
		}
	} else {
		t.cid = nil
	}

	records := m.GetRecords()
	if len(records) == 0 {
		return errors.New("missing records")
	}

	t.records = make([]Record, len(records))
	for i := range records {
		err := t.records[i].readFromV2(records[i])
		if err != nil {
			return fmt.Errorf("invalid record #%d: %w", i, err)
		}
	}

	return nil
}

// ReadFromV2 reads Table from the [v2acl.Table] messages. Returns an error if
// any message is malformed according to the NeoFS API V2 protocol. Behavior is
// forward-compatible:
//   - unknown enum values are considered valid
//   - unknown format of binary public keys is considered valid
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Table.WriteToV2].
func (t *Table) ReadFromV2(m acl.Table) error {
	return t.readFromV2(m, true)
}

// Marshal encodes Table into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also [Table.Unmarshal].
func (t Table) Marshal() []byte {
	var m acl.Table
	t.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// SignedData returns signed data of the Table.
//
// See also [client.Client.ContainerSetEACL].
func (t Table) SignedData() []byte {
	return t.Marshal()
}

// Unmarshal decodes NeoFS API protocol binary format into the Table (Protocol
// Buffers with direct field order). Returns an error describing a format
// violation.
//
// See also [Table.Marshal].
func (t *Table) Unmarshal(data []byte) error {
	var m acl.Table
	if err := m.Unmarshal(data); err != nil {
		return err
	}

	return t.readFromV2(m, false)
}

// MarshalJSON encodes Table into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also [Table.UnmarshalJSON].
func (t Table) MarshalJSON() ([]byte, error) {
	var m acl.Table
	t.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Table
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also [Table.MarshalJSON].
func (t *Table) UnmarshalJSON(data []byte) error {
	var m acl.Table
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}

	return t.readFromV2(m, false)
}
