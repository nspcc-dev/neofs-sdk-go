package storagegroup

import (
	"errors"
	"fmt"
	"strconv"

	objectV2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	objectSDK "github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// StorageGroup represents storage group of the NeoFS objects.
//
// StorageGroup is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/storagegroup.StorageGroup
// message. See ReadFromMessageV2 / WriteToMessageV2 methods.
//
// Instances should be created using one of the constructors.
type StorageGroup struct {
	sz      uint64
	exp     uint64
	members []oid.ID
	csSet   bool
	cs      checksum.Checksum
}

// New constructs new StorageGroup from given members and total size/checksum
// calculated for them. Member set must not be empty.
func New(size uint64, cs checksum.Checksum, members []oid.ID) StorageGroup {
	var res StorageGroup
	res.SetValidationDataSize(size)
	res.SetValidationDataHash(cs)
	res.SetMembers(members)
	return res
}

// Unmarshal creates new StorageGroup and makes [StorageGroup.Unmarshal].
func Unmarshal(b []byte) (StorageGroup, error) {
	var res StorageGroup
	return res, res.Unmarshal(b)
}

// UnmarshalJSON creates new StorageGroup and makes
// [StorageGroup.UnmarshalJSON].
func UnmarshalJSON(b []byte) (StorageGroup, error) {
	var res StorageGroup
	return res, res.UnmarshalJSON(b)
}

// reads StorageGroup from the storagegroup.StorageGroup message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field.
func (sg *StorageGroup) readFromV2(m storagegroup.StorageGroup, checkFieldPresence bool) error {
	var err error

	h := m.GetValidationHash()
	if sg.csSet = h != nil; sg.csSet {
		err = sg.cs.ReadFromV2(*h)
		if err != nil {
			return fmt.Errorf("invalid hash: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing hash")
	}

	members := m.GetMembers()
	if len(members) > 0 {
		sg.members = make([]oid.ID, len(members))
		for i := range members {
			err = sg.members[i].ReadFromV2(members[i])
			if err != nil {
				return fmt.Errorf("invalid member #%d: %w", i, err)
			}

			for j := range i {
				if sg.members[j] == sg.members[i] {
					return fmt.Errorf("duplicated member %s", sg.members[i])
				}
			}
		}
	} else if checkFieldPresence {
		return errors.New("missing members")
	}

	sg.sz = m.GetValidationDataSize()
	//nolint:staticcheck
	sg.exp = m.GetExpirationEpoch()

	return nil
}

// ReadFromV2 reads StorageGroup from the storagegroup.StorageGroup message.
// Checks if the message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (sg *StorageGroup) ReadFromV2(m storagegroup.StorageGroup) error {
	return sg.readFromV2(m, true)
}

// WriteToV2 writes StorageGroup to the storagegroup.StorageGroup message.
// The message must not be nil.
//
// See also ReadFromV2.
func (sg StorageGroup) WriteToV2(m *storagegroup.StorageGroup) {
	if sg.csSet {
		var cs refs.Checksum
		sg.cs.WriteToV2(&cs)
		m.SetValidationHash(&cs)
	}
	var members []refs.ObjectID
	if len(sg.members) > 0 {
		members = make([]refs.ObjectID, len(sg.members))
		for i := range sg.members {
			sg.members[i].WriteToV2(&members[i])
		}
	}
	m.SetMembers(members)
	m.SetValidationDataSize(sg.sz)
	//nolint:staticcheck
	m.SetExpirationEpoch(sg.exp)
}

// ValidationDataSize returns total size of the payloads
// of objects in the storage group.
//
// Zero StorageGroup has 0 data size.
//
// See also SetValidationDataSize.
func (sg StorageGroup) ValidationDataSize() uint64 {
	return sg.sz
}

// SetValidationDataSize sets total size of the payloads
// of objects in the storage group.
//
// See also ValidationDataSize.
func (sg *StorageGroup) SetValidationDataSize(sz uint64) {
	sg.sz = sz
}

// ValidationDataHash returns homomorphic hash from the
// concatenation of the payloads of the storage group members
// and bool that indicates checksum presence in the storage
// group.
//
// Zero StorageGroup does not have validation data checksum.
//
// See also SetValidationDataHash.
func (sg StorageGroup) ValidationDataHash() (checksum.Checksum, bool) {
	return sg.cs, sg.csSet
}

// SetValidationDataHash sets homomorphic hash from the
// concatenation of the payloads of the storage group members.
//
// See also ValidationDataHash.
func (sg *StorageGroup) SetValidationDataHash(hash checksum.Checksum) {
	sg.cs, sg.csSet = hash, true
}

// ExpirationEpoch returns last NeoFS epoch number
// of the storage group lifetime.
//
// Zero StorageGroup has 0 expiration epoch.
//
// See also SetExpirationEpoch.
// Deprecated: use expiration attribute in header of the object carrying
// StorageGroup.
func (sg StorageGroup) ExpirationEpoch() uint64 {
	return sg.exp
}

// SetExpirationEpoch sets last NeoFS epoch number
// of the storage group lifetime.
//
// See also ExpirationEpoch.
// Deprecated: use expiration attribute in header of the object carrying
// StorageGroup.
func (sg *StorageGroup) SetExpirationEpoch(epoch uint64) {
	sg.exp = epoch
}

// Members returns strictly ordered list of
// storage group member objects.
//
// Zero StorageGroup has nil members value.
//
// See also SetMembers.
func (sg StorageGroup) Members() []oid.ID {
	return sg.members
}

// SetMembers sets strictly ordered list of
// storage group member objects.
//
// See also Members.
func (sg *StorageGroup) SetMembers(members []oid.ID) {
	sg.members = members
}

// Marshal marshals StorageGroup into a protobuf binary form.
//
// See also Unmarshal.
func (sg StorageGroup) Marshal() []byte {
	var m storagegroup.StorageGroup
	sg.WriteToV2(&m)
	return m.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of StorageGroup.
//
// See also Marshal.
func (sg *StorageGroup) Unmarshal(data []byte) error {
	var m storagegroup.StorageGroup
	if err := m.Unmarshal(data); err != nil {
		return err
	}

	return sg.readFromV2(m, false)
}

// MarshalJSON encodes StorageGroup to protobuf JSON format.
//
// See also UnmarshalJSON.
func (sg StorageGroup) MarshalJSON() ([]byte, error) {
	var m storagegroup.StorageGroup
	sg.WriteToV2(&m)
	return m.MarshalJSON()
}

// UnmarshalJSON decodes StorageGroup from protobuf JSON format.
//
// See also MarshalJSON.
func (sg *StorageGroup) UnmarshalJSON(data []byte) error {
	var m storagegroup.StorageGroup
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}

	return sg.readFromV2(m, false)
}

// ReadFromObject assemble StorageGroup from a regular
// Object structure. Object must contain unambiguous information
// about its expiration epoch, otherwise behaviour is undefined.
//
// Returns any error appeared during storage group parsing; returns
// error if object is not of TypeStorageGroup type.
func ReadFromObject(sg *StorageGroup, o objectSDK.Object) error {
	if typ := o.Type(); typ != objectSDK.TypeStorageGroup {
		return fmt.Errorf("object is not of StorageGroup type: %v", typ)
	}

	err := sg.Unmarshal(o.Payload())
	if err != nil {
		return fmt.Errorf("could not unmarshal storage group from object payload: %w", err)
	}

	var expObj uint64

	for _, attr := range o.Attributes() {
		if attr.Key() == objectV2.SysAttributeExpEpoch {
			expObj, err = strconv.ParseUint(attr.Value(), 10, 64)
			if err != nil {
				return fmt.Errorf("could not get expiration from object: %w", err)
			}

			break
		}
	}

	// Supporting deprecated functionality.
	// See https://github.com/nspcc-dev/neofs-api/pull/205.
	if expSG := sg.ExpirationEpoch(); expSG > 0 && expObj != expSG {
		return fmt.Errorf(
			"expiration does not match: from object: %d, from payload: %d",
			expObj, expSG)
	}

	return nil
}

// WriteToObject writes StorageGroup to a regular
// Object structure. Object must not contain ambiguous
// information about its expiration epoch or must not
// have it at all.
//
// Written information:
//   - expiration epoch;
//   - object type (TypeStorageGroup);
//   - raw payload.
func WriteToObject(sg StorageGroup, o *objectSDK.Object) {
	o.SetPayload(sg.Marshal())
	o.SetType(objectSDK.TypeStorageGroup)

	attrs := o.Attributes()
	var expAttrFound bool

	for i := range attrs {
		if attrs[i].Key() == objectV2.SysAttributeExpEpoch {
			expAttrFound = true
			attrs[i].SetValue(strconv.FormatUint(sg.ExpirationEpoch(), 10))

			break
		}
	}

	if !expAttrFound {
		var attr objectSDK.Attribute

		attr.SetKey(objectV2.SysAttributeExpEpoch)
		attr.SetValue(strconv.FormatUint(sg.ExpirationEpoch(), 10))

		attrs = append(attrs, attr)
	}

	o.SetAttributes(attrs...)
}
