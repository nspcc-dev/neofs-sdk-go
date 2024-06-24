package object

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/lock"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"google.golang.org/protobuf/proto"
)

// Lock represents record with locked objects, i.e. objects protected from
// deletion. SplitChain is stored and transmitted as payload of system NeoFS
// objects.
type Lock struct {
	list []oid.ID
}

// readFromV2 reads Lock from the lock.Lock message. Returns an error if the
// message is malformed according to the NeoFS API V2 protocol. The message must
// not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also writeToV2.
func (x *Lock) readFromV2(m *lock.Lock) error {
	if len(m.Members) == 0 {
		return errors.New("missing members")
	}

	x.list = make([]oid.ID, len(m.Members))
	for i := range m.Members {
		err := x.list[i].ReadFromV2(m.Members[i])
		if err != nil {
			return fmt.Errorf("invalid member #%d: %w", i, err)
		}
	}

	return nil
}

// writeToV2 writes Lock to the lock.Lock message of the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also readFromV2.
func (x Lock) writeToV2(m *lock.Lock) {
	if len(x.list) > 0 {
		m.Members = make([]*refs.ObjectID, len(x.list))
		for i := range x.list {
			m.Members[i] = new(refs.ObjectID)
			x.list[i].WriteToV2(m.Members[i])
		}
	} else {
		x.list = nil
	}
}

// List returns list of locked objects.
//
// See also [Lock.SetList].
func (x Lock) List() []oid.ID {
	return x.list
}

// SetList sets list of locked objects.
//
// See also [Lock.List].
func (x *Lock) SetList(ids []oid.ID) {
	x.list = ids
}

// Marshal encodes Lock into a Protocol Buffers V3 binary format.
//
// See also [Lock.Unmarshal].
func (x Lock) Marshal() []byte {
	var m lock.Lock
	x.writeToV2(&m)

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Lock. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol.
//
// See also [Lock.Marshal].
func (x *Lock) Unmarshal(data []byte) error {
	var m lock.Lock
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m)
}
