package bpack

import "sync"

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			p := make([]byte, 4)
			return &p
		},
	}
)

func acquireBytes() *[]byte {
	return bytesPool.Get().(*[]byte)
}

func releaseBytes(p *[]byte) {
	if p != nil {
		bytesPool.Put(p)
	}
}
