package maintenance

import (
	"errors"
	"sort"
	"sync"
)

var ErrRuntimeInvalid = errors.New("IAM maintenance runtime is invalid")

type Runtime struct {
	mutex  sync.RWMutex
	keys   []DirectoryKey
	closed bool
}

func NewRuntime(input []DirectoryKey) (*Runtime, error) {
	if len(input) < 1 || len(input) > 65 {
		return nil, ErrRuntimeInvalid
	}
	keys := append([]DirectoryKey(nil), input...)
	sort.Slice(keys, func(left, right int) bool { return keys[left].ID < keys[right].ID })
	active := 0
	for index, key := range keys {
		if !key.valid() || index > 0 && keys[index-1].ID == key.ID {
			wipeKeys(keys)
			return nil, ErrRuntimeInvalid
		}
		if key.Active {
			active++
		}
	}
	if active != 1 {
		wipeKeys(keys)
		return nil, ErrRuntimeInvalid
	}
	return &Runtime{keys: keys}, nil
}

func (runtime *Runtime) snapshot() ([]DirectoryKey, error) {
	if runtime == nil {
		return nil, ErrRuntimeInvalid
	}
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	if runtime.closed || len(runtime.keys) == 0 {
		return nil, ErrRuntimeInvalid
	}
	return append([]DirectoryKey(nil), runtime.keys...), nil
}

func (runtime *Runtime) usable() bool {
	if runtime == nil {
		return false
	}
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	return !runtime.closed && len(runtime.keys) > 0
}

func (runtime *Runtime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	if !runtime.closed {
		wipeKeys(runtime.keys)
		runtime.keys = nil
		runtime.closed = true
	}
	return nil
}

func wipeKeys(keys []DirectoryKey) {
	for index := range keys {
		clear(keys[index].Material[:])
		keys[index].ID = ""
		keys[index].Active = false
	}
}
