package container

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// SizeEstimation groups information about estimation of the size of the data
// stored in the NeoFS container.
//
// SizeEstimation is mutually compatible with
// [container.AnnounceUsedSpaceRequest_Body_Announcement] message. See
// [SizeEstimation.ReadFromV2] / [SizeEstimation.WriteToV2] methods.
type SizeEstimation struct {
	epoch uint64
	val   uint64
	cnr   cid.ID
}

// ReadFromV2 reads SizeEstimation from the
// [container.AnnounceUsedSpaceRequest_Body_Announcement] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [SizeEstimation.WriteToV2].
func (x *SizeEstimation) ReadFromV2(m *container.AnnounceUsedSpaceRequest_Body_Announcement) error {
	if m.ContainerId == nil {
		return errors.New("missing container")
	}
	err := x.cnr.ReadFromV2(m.ContainerId)
	if err != nil {
		return fmt.Errorf("invalid container: %w", err)
	}

	x.epoch = m.Epoch
	x.val = m.UsedSpace

	return nil
}

// WriteToV2 writes ID to the
// [container.AnnounceUsedSpaceRequest_Body_Announcement] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [ID.ReadFromV2].
func (x SizeEstimation) WriteToV2(m *container.AnnounceUsedSpaceRequest_Body_Announcement) {
	m.ContainerId = new(refs.ContainerID)
	x.cnr.WriteToV2(m.ContainerId)
	m.Epoch = x.epoch
	m.UsedSpace = x.val
}

// SetEpoch sets epoch when estimation of the container data size was calculated.
//
// See also [SizeEstimation.Epoch].
func (x *SizeEstimation) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// Epoch returns epoch set using [SizeEstimation.SetEpoch].
//
// Zero SizeEstimation represents estimation in zero epoch.
func (x SizeEstimation) Epoch() uint64 {
	return x.epoch
}

// SetContainer specifies the container for which the amount of data is
// estimated. Required by the NeoFS API protocol.
//
// See also [SizeEstimation.Container].
func (x *SizeEstimation) SetContainer(cnr cid.ID) {
	x.cnr = cnr
}

// Container returns container set using [SizeEstimation.SetContainer].
//
// Zero SizeEstimation is not bound to any container (returns zero) which is
// incorrect according to NeoFS API protocol.
func (x SizeEstimation) Container() cid.ID {
	return x.cnr
}

// SetValue sets estimated amount of data (in bytes) in the specified container.
//
// See also [SizeEstimation.Value].
func (x *SizeEstimation) SetValue(value uint64) {
	x.val = value
}

// Value returns data size estimation set using [SizeEstimation.SetValue].
//
// Zero SizeEstimation has zero value.
func (x SizeEstimation) Value() uint64 {
	return x.val
}
