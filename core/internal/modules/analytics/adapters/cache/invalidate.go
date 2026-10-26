package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (cache *Cache) Invalidate(ctx context.Context, queryID uuid.UUID) error {
	if cache == nil || ctx == nil || ctx.Err() != nil || queryID == uuid.Nil {
		return fmt.Errorf("analytics cache invalidation is invalid")
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	for key := range cache.queries[queryID] {
		delete(cache.entries, key)
	}
	delete(cache.queries, queryID)
	return nil
}
