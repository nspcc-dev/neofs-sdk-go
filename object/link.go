package object

import (
	"fmt"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Link is a payload of helper objects that contain the full list of the split
// chain of the big NeoFS objects. It is compatible with NeoFS API V2 protocol.
//
// Link instance can be written to the [Object], see
// [Object.WriteLink]/[Object.ReadLink].
type Link v2object.Link

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
	return (*v2object.Link)(l).StableMarshal(nil)
}

// Unmarshal decodes the [Link] from its NeoFS protocol binary representation.
//
// See also [Link.Marshal].
func (l *Link) Unmarshal(data []byte) error {
	m := (*v2object.Link)(l)
	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	var id oid.ID
	var i int
	m.IterateChildren(func(mo v2object.MeasuredObject) {
		if err == nil {
			if err = id.ReadFromV2(mo.ID); err != nil {
				err = fmt.Errorf("invalid member #%d: %w", i, err)
			}
		}
		i++
	})
	return err
}

// MeasuredObject groups object ID and its size length. It is compatible with
// NeoFS API V2 protocol.
type MeasuredObject v2object.MeasuredObject

// SetObjectID sets object identifier.
//
// See also [MeasuredObject.ObjectID].
func (m *MeasuredObject) SetObjectID(id oid.ID) {
	var idV2 refs.ObjectID
	id.WriteToV2(&idV2)

	m.ID = idV2
}

// ObjectID returns object identifier.
//
// See also [MeasuredObject.SetObjectID].
func (m *MeasuredObject) ObjectID() oid.ID {
	var id oid.ID
	if m.ID.GetValue() != nil {
		if err := id.ReadFromV2(m.ID); err != nil {
			panic(fmt.Errorf("invalid ID: %w", err))
		}
	}

	return id
}

// SetObjectSize sets size of the object.
//
// See also [MeasuredObject.ObjectSize].
func (m *MeasuredObject) SetObjectSize(s uint32) {
	m.Size = s
}

// ObjectSize returns size of the object.
//
// See also [MeasuredObject.SetObjectSize].
func (m *MeasuredObject) ObjectSize() uint32 {
	return m.Size
}

// Objects returns split chain's measured objects.
//
// See also [Link.SetObjects].
func (l *Link) Objects() []MeasuredObject {
	res := make([]MeasuredObject, (*v2object.Link)(l).NumberOfChildren())
	var i int
	var id oid.ID
	(*v2object.Link)(l).IterateChildren(func(object v2object.MeasuredObject) {
		if err := id.ReadFromV2(object.ID); err != nil {
			panic(fmt.Errorf("invalid member #%d: %w", i, err))
		}
		res[i] = MeasuredObject(object)
		i++
	})

	return res
}

// SetObjects sets split chain's measured objects.
//
// See also [Link.Objects].
func (l *Link) SetObjects(oo []MeasuredObject) {
	v2OO := make([]v2object.MeasuredObject, len(oo))
	for i, o := range oo {
		v2OO[i] = v2object.MeasuredObject(o)
	}

	(*v2object.Link)(l).SetChildren(v2OO)
}
