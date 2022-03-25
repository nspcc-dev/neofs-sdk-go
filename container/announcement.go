package container

import (
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// UsedSpaceAnnouncement is an announcement message used by storage nodes to
// estimate actual container sizes.
//
// UsedSpaceAnnouncement is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/container.UsedSpaceAnnouncement
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = UsedSpaceAnnouncement(accounting.UsedSpaceAnnouncement{}) // not recommended
type UsedSpaceAnnouncement container.UsedSpaceAnnouncement

// ReadFromV2 reads UsedSpaceAnnouncement from the container.UsedSpaceAnnouncement message.
//
// See also WriteToV2.
func (a *UsedSpaceAnnouncement) ReadFromV2(m container.UsedSpaceAnnouncement) {
	*a = UsedSpaceAnnouncement(m)
}

// WriteToV2 writes UsedSpaceAnnouncement to the container.UsedSpaceAnnouncement message.
// The message must not be nil.
//
// See also ReadFromV2.
func (a UsedSpaceAnnouncement) WriteToV2(m *container.UsedSpaceAnnouncement) {
	*m = (container.UsedSpaceAnnouncement)(a)
}

// Epoch returns NeoFS epoch when UsedSpaceAnnouncement was made.
//
// Zero UsedSpaceAnnouncement has 0 epoch.
//
// See also SetEpoch.
func (a UsedSpaceAnnouncement) Epoch() uint64 {
	v2 := (container.UsedSpaceAnnouncement)(a)
	return v2.GetEpoch()
}

// SetEpoch sets announcement epoch value.
//
// See also Epoch.
func (a *UsedSpaceAnnouncement) SetEpoch(epoch uint64) {
	(*container.UsedSpaceAnnouncement)(a).SetEpoch(epoch)
}

// ContainerID returns container ID of the announcement.
//
// Zero UsedSpaceAnnouncement has nil container ID.
//
// See also SetContainerID.
func (a UsedSpaceAnnouncement) ContainerID() *cid.ID {
	v2 := (container.UsedSpaceAnnouncement)(a)
	var cID cid.ID

	cidV2 := v2.GetContainerID()
	if cidV2 == nil {
		return nil
	}

	cID.ReadFromV2(*cidV2)

	return &cID
}

// SetContainerID sets announcement container value.
// Container ID must not be nil.
//
// See also ContainerID.
func (a *UsedSpaceAnnouncement) SetContainerID(cid *cid.ID) {
	var cidV2 refs.ContainerID
	cid.WriteToV2(&cidV2)

	(*container.UsedSpaceAnnouncement)(a).SetContainerID(&cidV2)
}

// UsedSpace returns announced used space in the container.
//
// Zero UsedSpaceAnnouncement has 0 used space size.
//
// See also SetUsedSpace.
func (a UsedSpaceAnnouncement) UsedSpace() uint64 {
	v2 := (container.UsedSpaceAnnouncement)(a)
	return v2.GetUsedSpace()
}

// SetUsedSpace sets used space value by specified container.
//
// See also UsedSpace.
func (a *UsedSpaceAnnouncement) SetUsedSpace(value uint64) {
	(*container.UsedSpaceAnnouncement)(a).SetUsedSpace(value)
}

// Marshal marshals UsedSpaceAnnouncement into a protobuf binary form.
func (a UsedSpaceAnnouncement) Marshal() ([]byte, error) {
	v2 := (container.UsedSpaceAnnouncement)(a)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of UsedSpaceAnnouncement.
func (a *UsedSpaceAnnouncement) Unmarshal(data []byte) error {
	return (*container.UsedSpaceAnnouncement)(a).Unmarshal(data)
}
