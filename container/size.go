package container

import (
	"errors"
	"fmt"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
)

// SizeEstimation groups information about estimation of the size of the data
// stored in the NeoFS container.
//
// SizeEstimation is mutually compatible with [container.UsedSpaceAnnouncement]
// message. See [Container.FromProtoMessage] / [Container.ProtoMessage] methods.
type SizeEstimation struct {
	epoch uint64
	cnr   cid.ID
	val   uint64
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [SizeEstimation.ProtoMessage].
func (x *SizeEstimation) FromProtoMessage(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) error {
	if m.ContainerId == nil {
		return errors.New("missing container")
	}

	err := x.cnr.FromProtoMessage(m.ContainerId)
	if err != nil {
		return fmt.Errorf("invalid container: %w", err)
	}

	x.epoch = m.Epoch
	x.val = m.UsedSpace

	return nil
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Container.FromProtoMessage].
func (x SizeEstimation) ProtoMessage() *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement {
	m := &protocontainer.AnnounceUsedSpaceRequest_Body_Announcement{
		Epoch:     x.epoch,
		UsedSpace: x.val,
	}
	if !x.cnr.IsZero() {
		m.ContainerId = x.cnr.ProtoMessage()
	}
	return m
}

// SetEpoch sets epoch when estimation of the container data size was calculated.
//
// See also Epoch.
func (x *SizeEstimation) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// Epoch return epoch set using SetEpoch.
//
// Zero SizeEstimation represents estimation in zero epoch.
func (x SizeEstimation) Epoch() uint64 {
	return x.epoch
}

// SetContainer specifies the container for which the amount of data is estimated.
// Required by the NeoFS API protocol.
//
// See also Container.
func (x *SizeEstimation) SetContainer(cnr cid.ID) {
	x.cnr = cnr
}

// Container returns container set using SetContainer.
//
// Zero SizeEstimation is not bound to any container (returns zero) which is
// incorrect according to NeoFS API protocol.
func (x SizeEstimation) Container() (res cid.ID) {
	return x.cnr
}

// SetValue sets estimated amount of data (in bytes) in the specified container.
//
// See also Value.
func (x *SizeEstimation) SetValue(value uint64) {
	x.val = value
}

// Value returns data size estimation set using SetValue.
//
// Zero SizeEstimation has zero value.
func (x SizeEstimation) Value() uint64 {
	return x.val
}
