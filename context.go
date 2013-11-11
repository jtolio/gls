package context

import (
	"runtime"
	"sync"
)

const (
	maxCallers = 64
)

var (
	stackTagPool   = &IdPool{}
	mgrRegistry    = []*ContextManager{}
	mgrRegistryMtx sync.RWMutex
)

type Values map[interface{}]interface{}

func currentStack(skip int) []uintptr {
	stack := make([]uintptr, maxCallers)
	return stack[:runtime.Callers(2+skip, stack)]
}

type ContextManager struct {
	mtx    sync.RWMutex
	values map[uint]Values
}

func NewContextManager() *ContextManager {
	mgr := &ContextManager{values: make(map[uint]Values)}
	mgrRegistryMtx.Lock()
	defer mgrRegistryMtx.Unlock()
	mgrRegistry = append(mgrRegistry, mgr)
	return mgr
}

func (m *ContextManager) AddValues(new_values Values, context_call func()) {
	if len(new_values) == 0 {
		context_call()
		return
	}

	tags := readStackTags(currentStack(1))

	m.mtx.Lock()
	values := new_values
	for _, tag := range tags {
		if existing_values, ok := m.values[tag]; ok {
			// oh, we found existing values, let's make a copy
			values = make(Values, len(existing_values)+len(new_values))
			for key, val := range existing_values {
				values[key] = val
			}
			for key, val := range new_values {
				values[key] = val
			}
			break
		}
	}
	new_tag := stackTagPool.Acquire()
	m.values[new_tag] = values
	m.mtx.Unlock()
	defer func() {
		m.mtx.Lock()
		delete(m.values, new_tag)
		m.mtx.Unlock()
		stackTagPool.Release(new_tag)
	}()

	addStackTag(new_tag, context_call)
}

func (m *ContextManager) GetValue(key interface{}) (
	value interface{}, ok bool) {

	tags := readStackTags(currentStack(1))
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, tag := range tags {
		if values, ok := m.values[tag]; ok {
			value, ok := values[key]
			return value, ok
		}
	}
	return "", false
}

func (m *ContextManager) getValues() Values {
	tags := readStackTags(currentStack(2))
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, tag := range tags {
		if values, ok := m.values[tag]; ok {
			return values
		}
	}
	return nil
}

func Go(cb func()) {
	mgrRegistryMtx.RLock()
	defer mgrRegistryMtx.RUnlock()

	for _, mgr := range mgrRegistry {
		values := mgr.getValues()
		if len(values) > 0 {
			mgr_copy := mgr
			cb_copy := cb
			cb = func() { mgr_copy.AddValues(values, cb_copy) }
		}
	}

	go cb()
}
