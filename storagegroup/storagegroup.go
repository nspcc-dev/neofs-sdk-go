package storagegroup

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// StorageGroup represents v2-compatible storage group.
//
// StorageGroup is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/storagegroup.StorageGroup
// message. See ReadFromMessageV2 / WriteToMessageV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = StorageGroup(storagegroup.StorageGroup) // not recommended
type StorageGroup storagegroup.StorageGroup

// ReadFromV2 reads StorageGroup from the storagegroup.StorageGroup message.
//
// See also WriteToV2.
func (sg *StorageGroup) ReadFromV2(m storagegroup.StorageGroup) {
	*sg = StorageGroup(m)
}

// WriteToV2 writes StorageGroup to the storagegroup.StorageGroup message.
// The message must not be nil.
//
// See also ReadFromV2.
func (sg StorageGroup) WriteToV2(m *storagegroup.StorageGroup) {
	*m = (storagegroup.StorageGroup)(sg)
}

// ValidationDataSize returns total size of the payloads
// of objects in the storage group.
//
// Zero StorageGroup has 0 data size.
//
// See also SetValidationDataSize.
func (sg StorageGroup) ValidationDataSize() uint64 {
	v2 := (storagegroup.StorageGroup)(sg)
	return v2.GetValidationDataSize()
}

// SetValidationDataSize sets total size of the payloads
// of objects in the storage group.
//
// See also ValidationDataSize.
func (sg *StorageGroup) SetValidationDataSize(epoch uint64) {
	(*storagegroup.StorageGroup)(sg).SetValidationDataSize(epoch)
}

// ValidationDataHash returns homomorphic hash from the
// concatenation of the payloads of the storage group members.
//
// Zero StorageGroup has zero checksum.
//
// See also SetValidationDataHash.
func (sg StorageGroup) ValidationDataHash() *checksum.Checksum {
	v2 := (storagegroup.StorageGroup)(sg)
	if checksumV2 := v2.GetValidationHash(); checksumV2 != nil {
		var v checksum.Checksum
		v.ReadFromV2(*checksumV2)

		return &v
	}

	return nil
}

// SetValidationDataHash sets homomorphic hash from the
// concatenation of the payloads of the storage group members.
//
// See also ValidationDataHash.
func (sg *StorageGroup) SetValidationDataHash(hash *checksum.Checksum) {
	var v2 refs.Checksum
	hash.WriteToV2(&v2)

	(*storagegroup.StorageGroup)(sg).SetValidationHash(&v2)
}

// ExpirationEpoch returns last NeoFS epoch number
// of the storage group lifetime.
//
// Zero StorageGroup has 0 expiration epoch.
//
// See also SetExpirationEpoch.
func (sg StorageGroup) ExpirationEpoch() uint64 {
	v2 := (storagegroup.StorageGroup)(sg)
	return v2.GetExpirationEpoch()
}

// SetExpirationEpoch sets last NeoFS epoch number
// of the storage group lifetime.
//
// See also ExpirationEpoch.
func (sg *StorageGroup) SetExpirationEpoch(epoch uint64) {
	(*storagegroup.StorageGroup)(sg).SetExpirationEpoch(epoch)
}

// Members returns strictly ordered list of
// storage group member objects.
//
// Zero StorageGroup has nil members value.
//
// See also SetMembers.
func (sg StorageGroup) Members() []oid.ID {
	v2 := (storagegroup.StorageGroup)(sg)
	mV2 := v2.GetMembers()

	if mV2 == nil {
		return nil
	}

	m := make([]oid.ID, len(mV2))

	for i := range mV2 {
		m[i] = *oid.NewIDFromV2(&mV2[i])
	}

	return m
}

// SetMembers sets strictly ordered list of
// storage group member objects.
//
// See also Members.
func (sg *StorageGroup) SetMembers(members []oid.ID) {
	mV2 := (*storagegroup.StorageGroup)(sg).GetMembers()

	if members == nil {
		mV2 = nil
	} else {
		ln := len(members)

		if cap(mV2) >= ln {
			mV2 = mV2[:0]
		} else {
			mV2 = make([]refs.ObjectID, ln)
		}

		for i := 0; i < ln; i++ {
			mV2[i] = *members[i].ToV2()
		}
	}

	(*storagegroup.StorageGroup)(sg).SetMembers(mV2)
}

// Marshal marshals StorageGroup into a protobuf binary form.
//
// See also Unmarshal.
func (sg StorageGroup) Marshal() ([]byte, error) {
	v2 := (storagegroup.StorageGroup)(sg)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of StorageGroup.
//
// See also Marshal.
func (sg *StorageGroup) Unmarshal(data []byte) error {
	return (*storagegroup.StorageGroup)(sg).Unmarshal(data)
}

// MarshalJSON encodes StorageGroup to protobuf JSON format.
//
// See also UnmarshalJSON.
func (sg StorageGroup) MarshalJSON() ([]byte, error) {
	v2 := (storagegroup.StorageGroup)(sg)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes StorageGroup from protobuf JSON format.
//
// See also MarshalJSON.
func (sg *StorageGroup) UnmarshalJSON(data []byte) error {
	return (*storagegroup.StorageGroup)(sg).UnmarshalJSON(data)
}
