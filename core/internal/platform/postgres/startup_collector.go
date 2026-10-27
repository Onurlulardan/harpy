package postgres

import (
	"fmt"
	"sync"
)

type startupCollector struct {
	mutex  sync.Mutex
	result *StartupResult
}

func (collector *startupCollector) accept(candidate StartupResult) error {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	if collector.result == nil {
		copy := candidate
		collector.result = &copy
		return nil
	}
	if *collector.result != candidate {
		return fmt.Errorf("%w: inconsistent database identity", ErrStartupPreflight)
	}
	return nil
}

func (collector *startupCollector) value() (StartupResult, error) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	if collector.result == nil {
		return StartupResult{}, fmt.Errorf("%w: no connection evidence", ErrStartupPreflight)
	}
	return *collector.result, nil
}
