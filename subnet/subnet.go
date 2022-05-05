package subnet

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/subnet"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// Info represents information about NeoFS subnet.
//
// The type is compatible with the corresponding message from NeoFS API V2 protocol.
//
// Zero value and nil pointer to it represents zero subnet w/o an owner.
type Info subnet.Info

// FromV2 initializes Info from subnet.Info message structure. Must not be called on nil.
func (x *Info) FromV2(msg subnet.Info) {
	*x = Info(msg)
}

// WriteToV2 writes Info to subnet.Info message structure. The message must not be nil.
func (x Info) WriteToV2(msg *subnet.Info) {
	*msg = subnet.Info(x)
}

// Marshal encodes Info into a binary format of NeoFS API V2 protocol (Protocol Buffers with direct field order).
func (x *Info) Marshal() ([]byte, error) {
	return (*subnet.Info)(x).StableMarshal(nil), nil
}

// Unmarshal decodes Info from NeoFS API V2 binary format (see Marshal). Must not be called on nil.
//
// Note: empty data corresponds to zero Info value or nil pointer to it.
func (x *Info) Unmarshal(data []byte) error {
	return (*subnet.Info)(x).Unmarshal(data)
}

// SetID sets the identifier of the subnet that Info describes.
func (x *Info) SetID(id subnetid.ID) {
	infov2 := (*subnet.Info)(x)

	idv2 := infov2.ID()
	if idv2 == nil {
		idv2 = new(refs.SubnetID)
		infov2.SetID(idv2)
	}

	id.WriteToV2(idv2)
}

// ReadID reads the identifier of the subnet that Info describes. Arg must not be nil.
func (x Info) ReadID(id *subnetid.ID) {
	infov2 := (subnet.Info)(x)

	idv2 := infov2.ID()
	if idv2 == nil {
		subnetid.MakeZero(id)
		return
	}

	id.FromV2(*idv2)
}

// SetOwner sets subnet owner ID.
func (x *Info) SetOwner(id owner.ID) {
	infov2 := (*subnet.Info)(x)

	idv2 := infov2.Owner()
	if idv2 == nil {
		idv2 = new(refs.OwnerID)
		infov2.SetOwner(idv2)
	}

	// FIXME: we need to implement and use owner.ID.WriteToV2() method
	*idv2 = *id.ToV2()
}

// ReadOwner reads the identifier of the subnet that Info describes.
// Must be called only if owner is set (see HasOwner). Arg must not be nil.
func (x Info) ReadOwner(id *owner.ID) {
	infov2 := (subnet.Info)(x)

	id2 := infov2.Owner()
	if id2 == nil {
		// TODO: implement owner.ID.Reset
		*id = owner.ID{}
		return
	}

	// TODO: we need to implement and use owner.ID.FromV2 method
	*id = *owner.NewIDFromV2(infov2.Owner())
}

// IsOwner checks subnet ownership.
func IsOwner(info Info, id owner.ID) bool {
	id2 := new(owner.ID)

	info.ReadOwner(id2)

	return id.Equal(id2)
}

// IDEquals checks if ID refers to subnet that Info describes.
func IDEquals(info Info, id subnetid.ID) bool {
	id2 := new(subnetid.ID)

	info.ReadID(id2)

	return id.Equals(id2)
}
