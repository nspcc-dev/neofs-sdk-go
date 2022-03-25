package address

import (
	"errors"
	"strings"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Address represents v2-compatible object address.
//
// Address is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Address
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Address(refs.Address{}) // not recommended
type Address refs.Address

var errInvalidAddressString = errors.New("incorrect format of the string object address")

const (
	addressParts     = 2
	addressSeparator = "/"
)

// ReadFromV2 reads Address from the refs.Address message.
//
// See also WriteToV2.
func (a *Address) ReadFromV2(m refs.Address) {
	*a = Address(m)
}

// WriteToV2 writes Address to the refs.Address message.
// The message must not be nil.
//
// See also ReadFromV2.
func (a Address) WriteToV2(m *refs.Address) {
	*m = (refs.Address)(a)
}

// ContainerID returns container identifier.
//
// Zero Address has nil container ID.
//
// See also SetContainerID.
func (a Address) ContainerID() *cid.ID {
	var cID cid.ID
	v2 := (refs.Address)(a)

	cidV2 := v2.GetContainerID()
	if cidV2 == nil {
		return nil
	}

	cID.ReadFromV2(*cidV2)

	return &cID
}

// SetContainerID sets container identifier.
// Container ID must not be nil.
//
// See also ContainerID.
func (a *Address) SetContainerID(id *cid.ID) {
	var cidV2 refs.ContainerID
	id.WriteToV2(&cidV2)

	(*refs.Address)(a).SetContainerID(&cidV2)
}

// ObjectID returns object identifier.
//
// Zero Address has nil object ID.
//
// See also SetObjectID.
func (a Address) ObjectID() *oid.ID {
	v2 := (refs.Address)(a)

	oidV2 := v2.GetObjectID()
	if oidV2 == nil {
		return nil
	}

	var id oid.ID
	id.ReadFromV2(*oidV2)

	return &id
}

// SetObjectID sets object identifier.
// Object ID must not be nil.
//
// See also ObjectID.
func (a *Address) SetObjectID(id *oid.ID) {
	var idV2 refs.ObjectID
	id.WriteToV2(&idV2)

	(*refs.Address)(a).SetObjectID(&idV2)
}

// String implements fmt.Stringer interface method.
func (a Address) String() string {
	var (
		stringCID string
		stringOID string
	)

	if cID := a.ContainerID(); cID != nil {
		stringCID = cID.String()
	}

	if oID := a.ObjectID(); oID != nil {
		stringOID = oID.String()
	}

	return strings.Join([]string{
		stringCID,
		stringOID,
	}, addressSeparator)
}

// Parse is a reverse action to String().
func (a *Address) Parse(s string) error {
	var (
		err   error
		oid   oid.ID
		id    cid.ID
		parts = strings.Split(s, addressSeparator)
	)

	if len(parts) != addressParts {
		return errInvalidAddressString
	} else if err = id.Parse(parts[0]); err != nil {
		return err
	} else if err = oid.Parse(parts[1]); err != nil {
		return err
	}

	a.SetObjectID(&oid)
	a.SetContainerID(&id)

	return nil
}

// Marshal marshals Address into a protobuf binary form.
func (a Address) Marshal() ([]byte, error) {
	v2 := (refs.Address)(a)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Address.
func (a *Address) Unmarshal(data []byte) error {
	return (*refs.Address)(a).Unmarshal(data)
}

// MarshalJSON encodes Address to protobuf JSON format.
func (a Address) MarshalJSON() ([]byte, error) {
	v2 := (refs.Address)(a)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes Address from protobuf JSON format.
func (a *Address) UnmarshalJSON(data []byte) error {
	return (*refs.Address)(a).UnmarshalJSON(data)
}
