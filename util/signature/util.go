package signature

import (
	"errors"
	"sync"
)

var bytesPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 5<<20)
		return &b
	},
}

func dataForSignature(src DataSource) ([]byte, error) {
	if src == nil {
		return nil, errors.New("nil source")
	}

	buf := *bytesPool.Get().(*[]byte)

	if size := src.SignedDataSize(); size < 0 {
		return nil, errors.New("negative length")
	} else if size <= cap(buf) {
		buf = buf[:size]
	} else {
		buf = make([]byte, size)
	}

	return src.ReadSignedData(buf)
}
