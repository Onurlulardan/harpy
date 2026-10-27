package cache

import (
	"bytes"
	"time"
)

func (cache *Cache) evict(now time.Time, incoming [32]byte) {
	for key, value := range cache.entries {
		if !now.Before(value.expiresAt) {
			cache.remove(key, value)
		}
	}
	if _, replacing := cache.entries[incoming]; replacing || len(cache.entries) < cache.limit {
		return
	}
	var selected [32]byte
	var selectedEntry entry
	found := false
	for key, value := range cache.entries {
		if !found || value.expiresAt.Before(selectedEntry.expiresAt) ||
			(value.expiresAt.Equal(selectedEntry.expiresAt) && bytes.Compare(key[:], selected[:]) < 0) {
			selected, selectedEntry, found = key, value, true
		}
	}
	if found {
		cache.remove(selected, selectedEntry)
	}
}

func (cache *Cache) remove(key [32]byte, value entry) {
	delete(cache.entries, key)
	keys := cache.queries[value.queryID]
	delete(keys, key)
	if len(keys) == 0 {
		delete(cache.queries, value.queryID)
	}
}
