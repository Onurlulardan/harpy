package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (cache *Cache) Write(
	ctx context.Context, request queryapp.CacheRequest, body []byte, expiresAt time.Time,
) error {
	key, valid := deriveKey(request)
	if cache == nil || ctx == nil || ctx.Err() != nil || !valid || len(body) == 0 ||
		len(body) > 1<<20 || !json.Valid(body) || expiresAt.IsZero() {
		return fmt.Errorf("analytics cache write is invalid")
	}
	now := cache.clock.Now()
	if !now.Before(expiresAt) || expiresAt.Sub(now) > 300*time.Second {
		return fmt.Errorf("analytics cache expiry is invalid")
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.evict(now, key)
	if previous, exists := cache.entries[key]; exists {
		cache.remove(key, previous)
	}
	cache.entries[key] = entry{queryID: request.QueryID,
		body: bytes.Clone(body), expiresAt: expiresAt}
	keys := cache.queries[request.QueryID]
	if keys == nil {
		keys = make(map[[32]byte]struct{})
		cache.queries[request.QueryID] = keys
	}
	keys[key] = struct{}{}
	return nil
}
