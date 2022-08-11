package gls

// though this could probably be better at keeping ids smaller, the goal of
// this class is to keep a registry of the smallest unique integer ids
// per-process possible

import (
	"sync"
	"sync/atomic"
)

type idPool struct {
	sync.Pool
	curID uint32
}

func (p *idPool) newID() uint32 {
	curID := atomic.AddUint32(&p.curID, 1)
	return curID - 1
}

func (p *idPool) Acquire() (id uint32) {
	return p.Get().(uint32)
}

func (p *idPool) Release(id uint32) {
	p.Put(id)
}
