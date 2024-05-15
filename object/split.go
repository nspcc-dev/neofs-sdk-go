package object

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/link"
	"github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"google.golang.org/protobuf/proto"
)

// SplitChainElement describes an object in the chain of dividing a NeoFS object
// into several parts.
type SplitChainElement struct {
	id oid.ID
	sz uint32
}

// ID returns element ID.
//
// See also [SplitChainElement.SetID], [Object.ID].
func (x SplitChainElement) ID() oid.ID {
	return x.id
}

// SetID sets element ID.
//
// See also [SplitChainElement.ID].
func (x *SplitChainElement) SetID(id oid.ID) {
	x.id = id
}

// PayloadSize returns size of the element payload.
//
// See also [SplitChainElement.SetPayloadSize], [Header.PayloadSize].
func (x SplitChainElement) PayloadSize() uint32 {
	return x.sz
}

// SetPayloadSize sets size of the element payload.
//
// See also [SplitChainElement.PayloadSize].
func (x *SplitChainElement) SetPayloadSize(sz uint32) {
	x.sz = sz
}

// SplitChain describes split-chain of a NeoFS object divided into several
// parts. SplitChain is stored and transmitted as payload of system NeoFS
// objects.
type SplitChain struct {
	els []SplitChainElement
}

// readFromV2 reads SplitChain from the link.Link message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also writeToV2.
func (x *SplitChain) readFromV2(m *link.Link) error {
	if len(m.Children) == 0 {
		return fmt.Errorf("missing elements")
	}

	x.els = make([]SplitChainElement, len(m.Children))
	for i := range m.Children {
		if m.Children[i] == nil {
			return fmt.Errorf("element #%d is nil", i)
		}
		if m.Children[i].Id == nil {
			return fmt.Errorf("invalid element #%d: missing ID", i)
		}
		err := x.els[i].id.ReadFromV2(m.Children[i].Id)
		if err != nil {
			return fmt.Errorf("invalid element #%d: invalid ID: %w", i, err)
		}
		x.els[i].sz = m.Children[i].Size
	}

	return nil
}

// writeToV2 writes SplitChain to the link.Link message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also readFromV2.
func (x SplitChain) writeToV2(m *link.Link) {
	if x.els != nil {
		m.Children = make([]*link.Link_MeasuredObject, len(x.els))
		for i := range x.els {
			m.Children[i] = &link.Link_MeasuredObject{
				Id:   new(refs.ObjectID),
				Size: x.els[i].sz,
			}
			x.els[i].id.WriteToV2(m.Children[i].Id)
		}
	} else {
		m.Children = nil
	}
}

// Marshal encodes SplitChain into a Protocol Buffers V3 binary format.
//
// See also [SplitChain.Unmarshal].
func (x SplitChain) Marshal() []byte {
	var m link.Link
	x.writeToV2(&m)

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the SplitChain.
// Returns an error if the message is malformed according to the NeoFS API V2
// protocol.
//
// See also [SplitChain.Marshal].
func (x *SplitChain) Unmarshal(data []byte) error {
	var m link.Link
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m)
}

// Elements returns sorted list with split-chain elements.
func (x SplitChain) Elements() []SplitChainElement {
	return x.els
}

// SetElements sets sorted list of split-chain elements.
func (x *SplitChain) SetElements(els []SplitChainElement) {
	x.els = els
}

// SplitInfo represents a collection of references related to particular
// [SplitChain].
type SplitInfo struct {
	first, last, link oid.ID
	// deprecated
	splitID []byte

	firstSet, lastSet, linkSet bool
}

// ReadFromV2 reads SplitInfo from the [object.SplitInfo] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [SplitInfo.WriteToV2].
func (s *SplitInfo) ReadFromV2(m *object.SplitInfo) error {
	if ln := len(m.SplitId); ln > 0 && ln != 16 {
		return fmt.Errorf("invalid split ID length %d", ln)
	}

	s.lastSet = m.LastPart != nil
	s.linkSet = m.Link != nil
	if !s.lastSet && !s.linkSet {
		return errors.New("both linking and last split-chain elements are missing")
	}

	if s.lastSet {
		err := s.last.ReadFromV2(m.LastPart)
		if err != nil {
			return fmt.Errorf("invalid last split-chain element: %w", err)
		}
	}
	if s.linkSet {
		err := s.link.ReadFromV2(m.Link)
		if err != nil {
			return fmt.Errorf("invalid linking split-chain element: %w", err)
		}
	}
	if s.firstSet = m.FirstPart != nil; s.firstSet {
		err := s.first.ReadFromV2(m.FirstPart)
		if err != nil {
			return fmt.Errorf("invalid first split-chain element: %w", err)
		}
	}

	s.splitID = m.SplitId

	return nil
}

// WriteToV2 writes SplitInfo to the [object.SplitInfo] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [SplitInfo.ReadFromV2].
func (s SplitInfo) WriteToV2(m *object.SplitInfo) {
	if s.lastSet {
		m.LastPart = new(refs.ObjectID)
		s.last.WriteToV2(m.LastPart)
	}
	if s.linkSet {
		m.Link = new(refs.ObjectID)
		s.link.WriteToV2(m.Link)
	}
	if s.firstSet {
		m.FirstPart = new(refs.ObjectID)
		s.first.WriteToV2(m.FirstPart)
	}
	m.SplitId = s.splitID
}

// LastPart returns identifier of the last split-chain element. Zero return
// indicates unset relation.
//
// See also [SplitInfo.SetLastPart], [SplitChain.Elements].
func (s SplitInfo) LastPart() oid.ID {
	if s.lastSet {
		return s.last
	}
	return oid.ID{}
}

// SetLastPart sets identifier of the last split-chain element.
//
// See also [SplitInfo.LastPart].
func (s *SplitInfo) SetLastPart(v oid.ID) {
	s.last, s.lastSet = v, true
}

// Linker returns identifier of the object carrying full [SplitChain] in its
// payload. Zero return indicates unset relation.
//
// See also [SplitInfo.SetLinker].
func (s SplitInfo) Linker() oid.ID {
	if s.linkSet {
		return s.link
	}
	return oid.ID{}
}

// SetLinker sets identifier of the object carrying full information about the
// split-chain in its payload.
//
// See also [SplitInfo.Linker], [SplitChain].
func (s *SplitInfo) SetLinker(v oid.ID) {
	s.link, s.linkSet = v, true
}

// FirstPart returns identifier of the first split-chain element. Zero return
// indicates unset relation.
//
// See also [SplitInfo.SetFirstPart], [Header.FirstSplitObject],
// [SplitChain.Elements].
func (s SplitInfo) FirstPart() oid.ID {
	if s.firstSet {
		return s.first
	}
	return oid.ID{}
}

// SetFirstPart sets identifier of the first split-chain element.
//
// See also [SplitInfo.FirstPart].
func (s *SplitInfo) SetFirstPart(v oid.ID) {
	s.first, s.firstSet = v, true
}

// Marshal encodes SplitInfo into a Protocol Buffers V3 binary format.
//
// See also [SplitInfo.Unmarshal].
func (s SplitInfo) Marshal() []byte {
	var m object.SplitInfo
	s.WriteToV2(&m)

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the SplitInfo. Returns
// an error if the message is malformed according to the NeoFS API V2 protocol.
//
// See also [SplitInfo.Marshal].
func (s *SplitInfo) Unmarshal(data []byte) error {
	var m object.SplitInfo
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return s.ReadFromV2(&m)
}

// SplitInfoError is an error wrapping SplitInfo which is returned to indicate
// split object: an object presented as several smaller objects.
type SplitInfoError SplitInfo

// Error implements built-in error interface.
func (s SplitInfoError) Error() string {
	return "object is split"
}
