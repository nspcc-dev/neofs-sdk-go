package storagegroup

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/storagegroup"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"google.golang.org/protobuf/proto"
)

// StorageGroup represents storage group of the NeoFS objects. StorageGroup is
// stored and transmitted as payload of system NeoFS objects.
//
// StorageGroup is mutually compatible with [storagegroup.StorageGroup] message.
// See [StorageGroup.ReadFromV2] / [StorageGroup.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type StorageGroup struct {
	sz  uint64
	exp uint64
	ids []oid.ID

	csSet bool
	cs    checksum.Checksum
}

// ValidationDataSize returns total size of the payloads
// of objects in the storage group.
//
// Zero StorageGroup has 0 data size.
//
// See also [StorageGroup.SetValidationDataSize].
func (sg StorageGroup) ValidationDataSize() uint64 {
	return sg.sz
}

// SetValidationDataSize sets total size of the payloads
// of objects in the storage group.
//
// See also [StorageGroup.ValidationDataSize].
func (sg *StorageGroup) SetValidationDataSize(sz uint64) {
	sg.sz = sz
}

// ValidationDataHash returns homomorphic hash from the
// concatenation of the payloads of the storage group members
// and bool that indicates checksum presence in the storage
// group.
//
// Zero StorageGroup does not have validation data checksum (zero type).
//
// See also [StorageGroup.SetValidationDataHash].
func (sg StorageGroup) ValidationDataHash() checksum.Checksum {
	if sg.csSet {
		return sg.cs
	}
	return checksum.Checksum{}
}

// SetValidationDataHash sets homomorphic hash from the
// concatenation of the payloads of the storage group members.
//
// See also [StorageGroup.ValidationDataHash].
func (sg *StorageGroup) SetValidationDataHash(hash checksum.Checksum) {
	sg.cs, sg.csSet = hash, true
}

// Members returns strictly ordered list of storage group member objects.
//
// Zero StorageGroup has no members.
//
// See also [StorageGroup.SetMembers].
func (sg StorageGroup) Members() []oid.ID {
	return sg.ids
}

// SetMembers sets strictly ordered list of
// storage group member objects.
//
// See also [StorageGroup.Members].
func (sg *StorageGroup) SetMembers(members []oid.ID) {
	sg.ids = members
}

// Marshal encodes StorageGroup into a Protocol Buffers V3 binary format.
//
// See also [StorageGroup.Unmarshal].
func (sg StorageGroup) Marshal() []byte {
	m := storagegroup.StorageGroup{
		ValidationDataSize: sg.sz,
		ExpirationEpoch:    sg.exp,
	}

	if sg.csSet {
		m.ValidationHash = new(refs.Checksum)
		sg.cs.WriteToV2(m.ValidationHash)
	}

	if sg.ids != nil {
		m.Members = make([]*refs.ObjectID, len(sg.ids))
		for i := range sg.ids {
			m.Members[i] = new(refs.ObjectID)
			sg.ids[i].WriteToV2(m.Members[i])
		}
	}

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}

	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the StorageGroup.
// Returns an error describing a format violation of the specified fields.
// Unmarshal does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [StorageGroup.Marshal].
func (sg *StorageGroup) Unmarshal(data []byte) error {
	var m storagegroup.StorageGroup
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	if sg.csSet = m.ValidationHash != nil; sg.csSet {
		err = sg.cs.ReadFromV2(m.ValidationHash)
		if err != nil {
			return fmt.Errorf("invalid hash: %w", err)
		}
	}

	if len(m.Members) > 0 {
		sg.ids = make([]oid.ID, len(m.Members))
		for i := range m.Members {
			if m.Members[i] == nil {
				return fmt.Errorf("member #%d is nil", m.Members[i])
			}
			err = sg.ids[i].ReadFromV2(m.Members[i])
			if err != nil {
				return fmt.Errorf("invalid member #%d: %w", i, err)
			}

			for j := 0; j < i; j++ {
				if sg.ids[i] == sg.ids[j] {
					return fmt.Errorf("duplicated member %s", sg.ids[i])
				}
			}
		}
	} else {
		sg.ids = nil
	}

	sg.sz = m.ValidationDataSize
	sg.exp = m.ExpirationEpoch

	return nil
}

// // ReadFromObject reads StorageGroup from the NeoFS object. Object must contain
// // unambiguous information about its expiration epoch, otherwise behaviour is
// // undefined.
// //
// // Returns any error appeared during storage group parsing; returns error if
// // object is not of [object.TypeStorageGroup] type.
// func ReadFromObject(sg *StorageGroup, o object.Object) error {
// 	if typ := o.Type(); typ != object.TypeStorageGroup {
// 		return fmt.Errorf("object is not of StorageGroup type: %v", typ)
// 	}
//
// 	err := sg.Unmarshal(o.Payload())
// 	if err != nil {
// 		return fmt.Errorf("could not unmarshal object: %w", err)
// 	}
//
// 	var expObj uint64
//
// 	for _, attr := range o.Attributes() {
// 		if attr.Key() == object.AttributeExpirationEpoch {
// 			expObj, err = strconv.ParseUint(attr.Value(), 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("could not get expiration from object: %w", err)
// 			}
//
// 			break
// 		}
// 	}
//
// 	// Supporting deprecated functionality.
// 	// See https://github.com/nspcc-dev/neofs-api/pull/205.
// 	if expObj != sg.exp {
// 		return fmt.Errorf(
// 			"expiration does not match: from object: %d, from payload: %d",
// 			expObj, sg.exp)
// 	}
//
// 	return nil
// }
//
// // WriteToObject writes StorageGroup to the NeoFS object. Object must not
// // contain ambiguous information about its expiration epoch or must not have it
// // at all.
// //
// // Written information:
// //   - expiration epoch;
// //   - object type ([object.TypeStorageGroup]);
// //   - raw payload.
// func WriteToObject(sg StorageGroup, o *object.Object) {
// 	o.SetPayload(sg.Marshal())
// 	o.SetType(object.TypeStorageGroup)
//
// 	// TODO: simplify object attribute setting like for container
// 	attrs := o.Attributes()
// 	var expAttrFound bool
//
// 	for i := range attrs {
// 		if attrs[i].Key() == object.AttributeExpirationEpoch {
// 			expAttrFound = true
// 			attrs[i].SetValue(strconv.FormatUint(sg.exp, 10))
//
// 			break
// 		}
// 	}
//
// 	if !expAttrFound {
// 		var attr object.Attribute
//
// 		attr.SetKey(object.AttributeExpirationEpoch)
// 		attr.SetValue(strconv.FormatUint(sg.exp, 10))
//
// 		attrs = append(attrs, attr)
// 	}
//
// 	o.SetAttributes(attrs...)
// }
