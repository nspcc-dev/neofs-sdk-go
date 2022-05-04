package storagegroup

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// StorageGroup represents storage group of the NeoFS objects.
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
// concatenation of the payloads of the storage group members
// and bool that indicates checksum presence in the storage
// group.
//
// Zero StorageGroup does not have validation data checksum.
//
// See also SetValidationDataHash.
func (sg StorageGroup) ValidationDataHash() (v checksum.Checksum, isSet bool) {
	v2 := (storagegroup.StorageGroup)(sg)
	if checksumV2 := v2.GetValidationHash(); checksumV2 != nil {
		v.ReadFromV2(*checksumV2)
		isSet = true
	}

	return
}

// SetValidationDataHash sets homomorphic hash from the
// concatenation of the payloads of the storage group members.
//
// See also ValidationDataHash.
func (sg *StorageGroup) SetValidationDataHash(hash checksum.Checksum) {
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
		_ = m[i].ReadFromV2(mV2[i])
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

		var oidV2 refs.ObjectID

		for i := 0; i < ln; i++ {
			members[i].WriteToV2(&oidV2)
			mV2[i] = oidV2
		}
	}

	(*storagegroup.StorageGroup)(sg).SetMembers(mV2)
}

// Marshal marshals StorageGroup into a protobuf binary form.
//
// See also Unmarshal.
func (sg StorageGroup) Marshal() ([]byte, error) {
	v2 := (storagegroup.StorageGroup)(sg)
	return v2.StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of StorageGroup.
//
// See also Marshal.
func (sg *StorageGroup) Unmarshal(data []byte) error {
	v2 := (*storagegroup.StorageGroup)(sg)
	err := v2.Unmarshal(data)
	if err != nil {
		return err
	}

	return formatCheck(v2)
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
	v2 := (*storagegroup.StorageGroup)(sg)
	err := v2.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return formatCheck(v2)
}

func formatCheck(v2 *storagegroup.StorageGroup) error {
	var oID oid.ID

	for _, m := range v2.GetMembers() {
		err := oID.ReadFromV2(m)
		if err != nil {
			return err
		}
	}

	return nil
}
