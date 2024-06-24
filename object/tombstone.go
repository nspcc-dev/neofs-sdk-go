package object

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/tombstone"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"google.golang.org/protobuf/proto"
)

// Tombstone contains information about removed objects. Tombstone is stored and
// transmitted as payload of system NeoFS objects.
type Tombstone struct {
	members []oid.ID
	exp     uint64 // deprecated
	splitID []byte // deprecated
}

// readFromV2 reads Tombstone from the [tombstone.Tombstone] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also writeToV2.
func (t *Tombstone) readFromV2(m *tombstone.Tombstone) error {
	if len(m.Members) == 0 {
		return errors.New("missing members")
	}

	if ln := len(m.SplitId); ln > 0 && ln != 16 {
		return fmt.Errorf("invalid split ID length %d", ln)
	}

	t.members = make([]oid.ID, len(m.Members))
	for i := range m.Members {
		if m.Members[i] == nil {
			return fmt.Errorf("member #%d is nil", i)
		}
		err := t.members[i].ReadFromV2(m.Members[i])
		if err != nil {
			return fmt.Errorf("invalid member #%d: %w", i, err)
		}
	}

	t.exp = m.ExpirationEpoch
	t.splitID = m.SplitId

	return nil
}

// writeToV2 writes Tombstone to the [tombstone.Tombstone] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also readFromV2.
func (t Tombstone) writeToV2(m *tombstone.Tombstone) {
	if t.members != nil {
		m.Members = make([]*refs.ObjectID, len(t.members))
		for i := range t.members {
			m.Members[i] = new(refs.ObjectID)
			t.members[i].WriteToV2(m.Members[i])
		}
	} else {
		m.Members = nil
	}

	m.ExpirationEpoch = t.exp
	m.SplitId = t.splitID
}

// Members returns list of objects to be deleted.
//
// See also [Tombstone.SetMembers].
func (t Tombstone) Members() []oid.ID {
	return t.members
}

// SetMembers sets list of objects to be deleted.
//
// See also [Tombstone.Members].
func (t *Tombstone) SetMembers(v []oid.ID) {
	t.members = v
}

// Marshal encodes Tombstone into a Protocol Buffers V3 binary format.
//
// See also [Tombstone.Unmarshal].
func (t Tombstone) Marshal() []byte {
	var m tombstone.Tombstone
	t.writeToV2(&m)

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Tombstone. Returns
// an error if the message is malformed according to the NeoFS API V2 protocol.
//
// See also [Tombstone.Marshal].
func (t *Tombstone) Unmarshal(data []byte) error {
	var m tombstone.Tombstone
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return t.readFromV2(&m)
}
