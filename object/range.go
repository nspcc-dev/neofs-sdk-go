package object

// Range represents v2 [object.Range] object payload range.
type Range struct{ off, ln uint64 }

// NewRange creates and initializes blank [Range].
//
// Defaults:
//   - offset: 0;
//   - length: 0.
func NewRange() *Range {
	return new(Range)
}

// GetLength returns payload range size.
//
// See also [Range.SetLength].
func (r *Range) GetLength() uint64 {
	return r.ln
}

// SetLength sets payload range size.
//
// See also [Range.GetLength].
func (r *Range) SetLength(v uint64) {
	r.ln = v
}

// GetOffset sets payload range offset from start.
//
// See also [Range.SetOffset].
func (r *Range) GetOffset() uint64 {
	return r.off
}

// SetOffset gets payload range offset from start.
//
// See also [Range.GetOffset].
func (r *Range) SetOffset(v uint64) {
	r.off = v
}
