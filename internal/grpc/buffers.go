package grpc

import (
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/mem"
)

// TODO: we also have proto/protobuf.MemBuffer type, any chance to share?

// MemBuffer wraps [sync.Pool] item of *[]byte type to provide [mem.Buffer].
type MemBuffer struct {
	// Sub-slice of original item.
	mem.SliceBuffer
	item *[]byte
	pool *sync.Pool
	refs atomic.Int32
}

// Ref implements [mem.Buffer].
//
// Ref panics if buffer has already been freed.
func (x *MemBuffer) Ref() {
	if x.refs.Add(1) <= 1 {
		panic("ref of freed buffer")
	}
}

// Free implements [mem.Buffer].
//
// Free panics if buffer has already been freed.
func (x *MemBuffer) Free() {
	switch refs := x.refs.Add(-1); {
	case refs > 0:
	case refs == 0:
		x.pool.Put(x.item)
	default:
		panic("free of freed buffer")
	}
}

// NewMemBuffer initializes MemBuffer with reference counter set to 1. Item is
// put into pool once reference counter reaches zero.
func NewMemBuffer(item *[]byte, pool *sync.Pool) *MemBuffer {
	res := &MemBuffer{item: item, pool: pool}
	res.refs.Store(1)
	return res
}
