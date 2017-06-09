package gls

import (
	"sync/atomic"
)

var (
	keyCounter uint64
)

// ContextKey is a throwaway value you can use as a key to a ContextManager
type ContextKey struct{ id uint64 }

// GenSym will return a brand new, never-before-used ContextKey
func GenSym() ContextKey {
	return ContextKey{id: atomic.AddUint64(&keyCounter, 1)}
}
