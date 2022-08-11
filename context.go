// Package gls implements goroutine-local storage.
package gls

import (
	"sync"
)

const (
	initialMaxGoroutineCount = 1024
	extendUnit               = 128
)

var (
	mgrRegistry    = make(map[*ContextManager]bool)
	mgrRegistryMtx sync.RWMutex
)

// Values is simply a map of key types to value types. Used by SetValues to
// set multiple values at once.
type Values map[interface{}]interface{}

// ContextManager is the main entrypoint for interacting with
// Goroutine-local-storage. You can have multiple independent ContextManagers
// at any given time. ContextManagers are usually declared globally for a given
// class of context variables. You should use NewContextManager for
// construction.
type ContextManager struct {
	extendLock               sync.RWMutex
	values                   []Values
	currentMaxGoroutineCount int
}

// NewContextManager returns a brand new ContextManager. It also registers the
// new ContextManager in the ContextManager registry which is used by the Go
// method. ContextManagers are typically defined globally at package scope.
func NewContextManager() *ContextManager {
	mgr := &ContextManager{values: make([]Values, initialMaxGoroutineCount)}
	mgr.currentMaxGoroutineCount = len(mgr.values)
	mgrRegistryMtx.Lock()
	defer mgrRegistryMtx.Unlock()
	mgrRegistry[mgr] = true
	return mgr
}

// Unregister removes a ContextManager from the global registry, used by the
// Go method. Only intended for use when you're completely done with a
// ContextManager. Use of Unregister at all is rare.
func (m *ContextManager) Unregister() {
	mgrRegistryMtx.Lock()
	defer mgrRegistryMtx.Unlock()
	delete(mgrRegistry, m)
}

// SetValues takes a collection of values and a function to call for those
// values to be set in. Anything further down the stack will have the set
// values available through GetValue. SetValues will add new values or replace
// existing values of the same key and will not mutate or change values for
// previous stack frames.
// SetValues is slow (makes a copy of all current and new values for the new
// gls-context) in order to reduce the amount of lookups GetValue requires.
func (m *ContextManager) SetValues(new_values Values, context_call func()) {
	if len(new_values) == 0 {
		context_call()
		return
	}

	mutated_keys := make([]interface{}, 0, len(new_values))
	mutated_vals := make(Values, len(new_values))

	EnsureGoroutineId(func(gid uint32) {
		var found bool
		m.extendIfNeeded(gid)

		state := m.values[gid]
		if state != nil {
			found = true
		} else {
			state = make(Values, len(new_values))
			m.values[gid] = state
		}

		for key, new_val := range new_values {
			mutated_keys = append(mutated_keys, key)
			if old_val, ok := state[key]; ok {
				mutated_vals[key] = old_val
			}
			state[key] = new_val
		}

		defer func() {
			if !found {
				m.values[gid] = nil
				return
			}

			for _, key := range mutated_keys {
				if val, ok := mutated_vals[key]; ok {
					state[key] = val
				} else {
					delete(state, key)
				}
			}
		}()

		context_call()
	})
}

// GetValue will return a previously set value, provided that the value was set
// by SetValues somewhere higher up the stack. If the value is not found, ok
// will be false.
func (m *ContextManager) GetValue(key interface{}) (
	value interface{}, ok bool) {
	gid, ok := GetGoroutineId()
	if !ok {
		return nil, false
	}

	state := m.values[gid]

	if state == nil {
		return nil, false
	}
	value, ok = state[key]
	return value, ok
}

func (m *ContextManager) getValues() Values {
	gid, ok := GetGoroutineId()
	if !ok {
		return nil
	}
	state := m.values[gid]
	return state
}

// Go preserves ContextManager values and Goroutine-local-storage across new
// goroutine invocations. The Go method makes a copy of all existing values on
// all registered context managers and makes sure they are still set after
// kicking off the provided function in a new goroutine. If you don't use this
// Go method instead of the standard 'go' keyword, you will lose values in
// ContextManagers, as goroutines have brand new stacks.
func Go(cb func()) {
	mgrRegistryMtx.RLock()
	defer mgrRegistryMtx.RUnlock()

	for mgr := range mgrRegistry {
		values := mgr.getValues()
		if len(values) > 0 {
			cb = func(mgr *ContextManager, cb func()) func() {
				return func() { mgr.SetValues(values, cb) }
			}(mgr, cb)
		}
	}

	go cb()
}

func (m *ContextManager) extend(gid uint32) {
	m.extendLock.Lock()
	defer m.extendLock.Unlock()
	if gid >= uint32(m.currentMaxGoroutineCount) {
		unit := ((gid-uint32(m.currentMaxGoroutineCount))/extendUnit + 1) * extendUnit
		m.values = append(m.values, make([]Values, unit)...)
		m.currentMaxGoroutineCount += int(unit)
	}
}

func (m *ContextManager) extendIfNeeded(gid uint32) {
	m.extendLock.RLock()
	if gid >= uint32(m.currentMaxGoroutineCount) {
		m.extendLock.RUnlock()
		m.extend(gid)
	} else {
		m.extendLock.RUnlock()
	}
}
