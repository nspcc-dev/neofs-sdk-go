package eacl

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Table represents NeoFS extended ACL (eACL): group of rules managing access to
// NeoFS resources in addition to the basic ACL.
//
// Table is mutually compatible with [acl.EACLTable] message. See
// [Table.ReadFromV2] / [Table.WriteToV2] methods.
type Table struct {
	decoded bool

	versionSet bool
	version    version.Version

	cnrSet bool
	cnr    cid.ID

	records []Record
}

// CopyTo writes deep copy of the [Table] to dst.
func (t Table) CopyTo(dst *Table) {
	dst.decoded = t.decoded
	dst.version = t.version
	dst.cnr, dst.cnrSet = t.cnr, t.cnrSet

	if t.records != nil {
		dst.records = make([]Record, len(t.records))
		for i := range t.records {
			t.records[i].CopyTo(&dst.records[i])
		}
	} else {
		dst.records = nil
	}
}

// LimitedContainer returns identifier of the NeoFS container to which the eACL
// scope is limited. Zero return means the eACL may be applied to any container.
//
// See also [Table.LimitToContainer].
func (t Table) LimitedContainer() cid.ID {
	if t.cnrSet {
		return t.cnr
	}
	return cid.ID{}
}

// LimitToContainer limits scope of the eACL to a referenced container. By
// default, the eACL is applicable to any container.
//
// See also [Table.LimitedContainer].
func (t *Table) LimitToContainer(cnr cid.ID) {
	t.cnr = cnr
	t.cnrSet = true
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

// SetRecords list of extended ACL rules.
//
// See also [Table.Records].
func (t *Table) SetRecords(rs []Record) {
	t.records = rs
}

// WriteToV2 writes Table to the [acl.EACLTable] message of the NeoFS API
// protocol.
//
// WriteToV2 writes current protocol version into the resulting message if
// Result hasn't been already decoded from such a message.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Table.ReadFromV2].
func (t Table) WriteToV2(m *acl.EACLTable) {
	if t.versionSet {
		m.Version = new(refs.Version)
		t.version.WriteToV2(m.Version)
	} else if !t.decoded {
		m.Version = new(refs.Version)
		version.Current.WriteToV2(m.Version)
	} else {
		m.Version = nil
	}
	if t.cnrSet {
		m.ContainerId = new(refs.ContainerID)
		t.cnr.WriteToV2(m.ContainerId)
	}
	if t.records != nil {
		m.Records = make([]*acl.EACLRecord, len(t.records))
		for i := range t.records {
			m.Records[i] = recordToAPI(t.records[i])
		}
	} else {
		m.Records = nil
	}
}

func (t *Table) readFromV2(m *acl.EACLTable, checkFieldPresence bool) error {
	var err error
	t.cnrSet = m.ContainerId != nil
	if t.cnrSet {
		err = t.cnr.ReadFromV2(m.ContainerId)
		if err != nil {
			return fmt.Errorf("invalid container: %w", err)
		}
	}

	t.versionSet = m.Version != nil
	if t.versionSet {
		err = t.version.ReadFromV2(m.Version)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	if len(m.Records) > 0 {
		t.records = make([]Record, len(m.Records))
		for i := range m.Records {
			if m.Records[i] != nil {
				err = t.records[i].readFromV2(m.Records[i], checkFieldPresence)
				if err != nil {
					return fmt.Errorf("invalid record #%d: %w", i, err)
				}
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing records")
	} else {
		t.records = nil
	}

	t.decoded = true

	return nil
}

// ReadFromV2 reads Table from the [acl.EACLTable] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Table.WriteToV2].
func (t *Table) ReadFromV2(m *acl.EACLTable) error {
	return t.readFromV2(m, true)
}

// Marshal encodes Table into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// Marshal writes current protocol version into the resulting message if Result
// hasn't been already decoded from such a message.
//
// See also [Table.Unmarshal].
func (t Table) Marshal() []byte {
	var m acl.EACLTable
	t.WriteToV2(&m)

	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// SignedData returns signed data of the Table.
//
// See also [client.Client.ContainerSetEACL].
func (t Table) SignedData() []byte {
	return t.Marshal()
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Table. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Table.Marshal].
func (t *Table) Unmarshal(data []byte) error {
	var m acl.EACLTable
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return t.readFromV2(&m, false)
}

// MarshalJSON encodes Table into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// MarshalJSON writes current protocol version into the resulting message if
// Result hasn't been already decoded from such a message.
//
// See also [Table.UnmarshalJSON].
func (t *Table) MarshalJSON() ([]byte, error) {
	var m acl.EACLTable
	t.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Table (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Table.MarshalJSON].
func (t *Table) UnmarshalJSON(data []byte) error {
	var m acl.EACLTable
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}
	return t.readFromV2(&m, false)
}
