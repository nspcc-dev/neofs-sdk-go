package container

import (
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// UsedSpaceAnnouncement is an announcement message used by storage nodes to
// estimate actual container sizes.
type UsedSpaceAnnouncement container.UsedSpaceAnnouncement

// NewAnnouncement initialize empty UsedSpaceAnnouncement message.
//
// Defaults:
//  - epoch: 0;
//  - usedSpace: 0;
//  - cid: nil.
func NewAnnouncement() *UsedSpaceAnnouncement {
	return NewAnnouncementFromV2(new(container.UsedSpaceAnnouncement))
}

// NewAnnouncementFromV2 wraps protocol dependent version of
// UsedSpaceAnnouncement message.
//
// Nil container.UsedSpaceAnnouncement converts to nil.
func NewAnnouncementFromV2(v *container.UsedSpaceAnnouncement) *UsedSpaceAnnouncement {
	return (*UsedSpaceAnnouncement)(v)
}

// Epoch of the announcement.
func (a *UsedSpaceAnnouncement) Epoch() uint64 {
	return (*container.UsedSpaceAnnouncement)(a).GetEpoch()
}

// SetEpoch sets announcement epoch value.
func (a *UsedSpaceAnnouncement) SetEpoch(epoch uint64) {
	(*container.UsedSpaceAnnouncement)(a).SetEpoch(epoch)
}

// ContainerID of the announcement.
func (a *UsedSpaceAnnouncement) ContainerID() (cID cid.ID) {
	v2 := (*container.UsedSpaceAnnouncement)(a)

	cidV2 := v2.GetContainerID()
	if cidV2 == nil {
		return
	}

	_ = cID.ReadFromV2(*cidV2)

	return
}

// SetContainerID sets announcement container value.
func (a *UsedSpaceAnnouncement) SetContainerID(cnr cid.ID) {
	var cidV2 refs.ContainerID
	cnr.WriteToV2(&cidV2)

	(*container.UsedSpaceAnnouncement)(a).SetContainerID(&cidV2)
}

// UsedSpace in container.
func (a *UsedSpaceAnnouncement) UsedSpace() uint64 {
	return (*container.UsedSpaceAnnouncement)(a).GetUsedSpace()
}

// SetUsedSpace sets used space value by specified container.
func (a *UsedSpaceAnnouncement) SetUsedSpace(value uint64) {
	(*container.UsedSpaceAnnouncement)(a).SetUsedSpace(value)
}

// ToV2 returns protocol dependent version of UsedSpaceAnnouncement message.
//
// Nil UsedSpaceAnnouncement converts to nil.
func (a *UsedSpaceAnnouncement) ToV2() *container.UsedSpaceAnnouncement {
	return (*container.UsedSpaceAnnouncement)(a)
}

// Marshal marshals UsedSpaceAnnouncement into a protobuf binary form.
func (a *UsedSpaceAnnouncement) Marshal() ([]byte, error) {
	return a.ToV2().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of UsedSpaceAnnouncement.
func (a *UsedSpaceAnnouncement) Unmarshal(data []byte) error {
	return a.ToV2().Unmarshal(data)
}
