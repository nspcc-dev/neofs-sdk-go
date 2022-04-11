package address

import (
	"errors"
	"strings"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Address represents v2-compatible object address.
type Address refs.Address

var errInvalidAddressString = errors.New("incorrect format of the string object address")

const (
	addressParts     = 2
	addressSeparator = "/"
)

// NewAddressFromV2 converts v2 Address message to Address.
//
// Nil refs.Address converts to nil.
func NewAddressFromV2(aV2 *refs.Address) *Address {
	return (*Address)(aV2)
}

// NewAddress creates and initializes blank Address.
//
// Works similar as NewAddressFromV2(new(Address)).
//
// Defaults:
// 	- cid: nil;
//	- oid: nil.
func NewAddress() *Address {
	return NewAddressFromV2(new(refs.Address))
}

// ToV2 converts Address to v2 Address message.
//
// Nil Address converts to nil.
func (a *Address) ToV2() *refs.Address {
	return (*refs.Address)(a)
}

// ContainerID returns container identifier.
func (a *Address) ContainerID() (v cid.ID) {
	var cID cid.ID
	v2 := (*refs.Address)(a)

	cidV2 := v2.GetContainerID()
	if cidV2 != nil {
		_ = cID.ReadFromV2(*cidV2)
	}

	return cID
}

// SetContainerID sets container identifier.
func (a *Address) SetContainerID(id cid.ID) {
	var cidV2 refs.ContainerID
	id.WriteToV2(&cidV2)

	(*refs.Address)(a).SetContainerID(&cidV2)
}

// ObjectID returns object identifier.
func (a *Address) ObjectID() (v oid.ID) {
	var id oid.ID
	v2 := (*refs.Address)(a)

	oidV2 := v2.GetObjectID()
	if oidV2 != nil {
		_ = id.ReadFromV2(*oidV2)
	}

	return id
}

// SetObjectID sets object identifier.
func (a *Address) SetObjectID(id oid.ID) {
	var oidV2 refs.ObjectID
	id.WriteToV2(&oidV2)

	(*refs.Address)(a).SetObjectID(&oidV2)
}

// Parse converts base58 string representation into Address.
func (a *Address) Parse(s string) error {
	var (
		err   error
		oid   oid.ID
		id    cid.ID
		parts = strings.Split(s, addressSeparator)
	)

	if len(parts) != addressParts {
		return errInvalidAddressString
	} else if err = id.DecodeString(parts[0]); err != nil {
		return err
	} else if err = oid.DecodeString(parts[1]); err != nil {
		return err
	}

	a.SetObjectID(oid)
	a.SetContainerID(id)

	return nil
}

// String returns string representation of Object.Address.
func (a *Address) String() string {
	return strings.Join([]string{
		a.ContainerID().String(),
		a.ObjectID().String(),
	}, addressSeparator)
}

// Marshal marshals Address into a protobuf binary form.
func (a *Address) Marshal() ([]byte, error) {
	return (*refs.Address)(a).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Address.
func (a *Address) Unmarshal(data []byte) error {
	return (*refs.Address)(a).Unmarshal(data)
}

// MarshalJSON encodes Address to protobuf JSON format.
func (a *Address) MarshalJSON() ([]byte, error) {
	return (*refs.Address)(a).MarshalJSON()
}

// UnmarshalJSON decodes Address from protobuf JSON format.
func (a *Address) UnmarshalJSON(data []byte) error {
	return (*refs.Address)(a).UnmarshalJSON(data)
}
