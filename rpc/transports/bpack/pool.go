package bpack

import "sync"

type PackPool struct {
	pool sync.Pool
}

func (pp *PackPool) Acquire() *Pack {
	v := pp.pool.Get()
	if v == nil {
		p, _ := New()
		return p
	}
	p := v.(*Pack)
	return p
}

func (pp *PackPool) Release(p *Pack) {
	if p == nil {
		return
	}
	p.dict.Reset()
	pp.pool.Put(p)
}

var (
	DefaultPool = PackPool{}
)

func Acquire() *Pack {
	return DefaultPool.Acquire()
}

func Release(p *Pack) {
	DefaultPool.Release(p)
}
