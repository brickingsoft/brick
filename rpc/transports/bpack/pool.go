package bpack

import "sync"

type PackPool struct {
	pool sync.Pool
}

func (pp *PackPool) Acquire() *Packer {
	v := pp.pool.Get()
	if v == nil {
		p, _ := New()
		return p
	}
	p := v.(*Packer)
	return p
}

func (pp *PackPool) Release(p *Packer) {
	if p == nil {
		return
	}
	p.dict.Reset()
	pp.pool.Put(p)
}

var (
	DefaultPool = PackPool{}
)

func Acquire() *Packer {
	return DefaultPool.Acquire()
}

func Release(p *Packer) {
	DefaultPool.Release(p)
}
