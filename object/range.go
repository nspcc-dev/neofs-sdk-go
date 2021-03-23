package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
)

// Range represents v2-compatible object payload range.
//
// Range is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.Range
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Range(object.Range{}) // not recommended
type Range object.Range

// ReadFromV2 reads Range from the object.Range message.
//
// See also WriteToV2.
func (r *Range) ReadFromV2(m object.Range) {
	*r = Range(m)
}

// WriteToV2 writes Range to the object.Range message.
// The message must not be nil.
//
// See also ReadFromV2.
func (r Range) WriteToV2(m *object.Range) {
	*m = (object.Range)(r)
}

// Length returns payload range size.
//
// Zero Range has 0 length.
//
// See also SetLength.
func (r Range) Length() uint64 {
	v2 := (object.Range)(r)
	return v2.GetLength()
}

// SetLength sets payload range size.
//
// See also Length.
func (r *Range) SetLength(v uint64) {
	(*object.Range)(r).SetLength(v)
}

// Offset sets payload range offset from start.
//
// Zero Range has 0 offset.
//
// See also SetOffset.
func (r Range) Offset() uint64 {
	v2 := (object.Range)(r)
	return v2.GetOffset()
}

// SetOffset gets payload range offset from start.
//
// See also Offset.
func (r *Range) SetOffset(v uint64) {
	(*object.Range)(r).SetOffset(v)
}
