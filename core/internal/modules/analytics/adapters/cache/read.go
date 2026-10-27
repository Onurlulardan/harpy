package cache

import (
	"bytes"
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (cache *Cache) Read(
	ctx context.Context, request queryapp.CacheRequest,
) ([]byte, bool, error) {
	key, valid := deriveKey(request)
	if cache == nil || ctx == nil || ctx.Err() != nil || !valid {
		return nil, false, nil
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	value, found := cache.entries[key]
	if !found {
		return nil, false, nil
	}
	if !cache.clock.Now().Before(value.expiresAt) {
		cache.remove(key, value)
		return nil, false, nil
	}
	return bytes.Clone(value.body), true, nil
}
