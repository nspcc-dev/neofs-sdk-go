package object

import (
	"fmt"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protolink "github.com/nspcc-dev/neofs-sdk-go/proto/link"
)

// Link is a payload of helper objects that contain the full list of the split
// chain of the big NeoFS objects. It is compatible with NeoFS API V2 protocol.
//
// Link instance can be written to the [Object], see
// [Object.WriteLink]/[Object.ReadLink].
type Link struct {
	children []MeasuredObject
}

// WriteLink writes a link to the Object, and sets its type to [TypeLink].
//
// See also ReadLink.
func (o *Object) WriteLink(l Link) {
	o.SetType(TypeLink)
	o.SetPayload(l.Marshal())
}

// ReadLink reads a link from the [Object]. The link must not be nil.
// Returns an error describing incorrect format. Makes sense only
// if the object has [TypeLink] type.
//
// See also [Object.WriteLink].
func (o Object) ReadLink(l *Link) error {
	return l.Unmarshal(o.Payload())
}

// Marshal encodes the [Link] into a NeoFS protocol binary format.
//
// See also [Link.Unmarshal].
func (l *Link) Marshal() []byte {
	if l == nil || len(l.children) == 0 {
		return nil
	}
	m := &protolink.Link{
		Children: make([]*protolink.Link_MeasuredObject, len(l.children)),
	}
	for i := range l.children {
		m.Children[i] = &protolink.Link_MeasuredObject{
			Id:   l.children[i].id.ProtoMessage(),
			Size: l.children[i].sz,
		}
	}
	return neofsproto.MarshalMessage(m)
}

// Unmarshal decodes the [Link] from its NeoFS protocol binary representation.
//
// See also [Link.Marshal].
func (l *Link) Unmarshal(data []byte) error {
	m := new(protolink.Link)
	err := neofsproto.UnmarshalMessage(data, m)
	if err != nil {
		return err
	}

	if m.Children == nil {
		l.children = nil
		return nil
	}

	l.children = make([]MeasuredObject, len(m.Children))
	for i := range m.Children {
		if m.Children[i] == nil {
			return fmt.Errorf("nil child #%d", i)
		}
		if m.Children[i].Id == nil {
			return fmt.Errorf("invalid child #%d: nil ID", i)
		}
		if err = l.children[i].id.FromProtoMessage(m.Children[i].Id); err != nil {
			return fmt.Errorf("invalid child #%d: invalid ID: %w", i, err)
		}
		l.children[i].sz = m.Children[i].Size
	}

	return nil
}

// MeasuredObject groups object ID and its size length.
type MeasuredObject struct {
	id oid.ID
	sz uint32
}

// SetObjectID sets object identifier.
//
// See also [MeasuredObject.ObjectID].
func (m *MeasuredObject) SetObjectID(id oid.ID) {
	m.id = id
}

// ObjectID returns object identifier.
//
// See also [MeasuredObject.SetObjectID].
func (m *MeasuredObject) ObjectID() oid.ID {
	return m.id
}

// SetObjectSize sets size of the object.
//
// See also [MeasuredObject.ObjectSize].
func (m *MeasuredObject) SetObjectSize(s uint32) {
	m.sz = s
}

// ObjectSize returns size of the object.
//
// See also [MeasuredObject.SetObjectSize].
func (m *MeasuredObject) ObjectSize() uint32 {
	return m.sz
}

// Objects returns split chain's measured objects.
//
// See also [Link.SetObjects].
func (l *Link) Objects() []MeasuredObject {
	return l.children
}

// SetObjects sets split chain's measured objects.
//
// See also [Link.Objects].
func (l *Link) SetObjects(oo []MeasuredObject) {
	l.children = oo
}
