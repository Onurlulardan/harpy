package contact

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) ReadContactStats(ctx context.Context, access ports.Access,
	observedAt time.Time,
) (domain.ContactStats, error) {
	if !store.ready() {
		return domain.ContactStats{}, ports.ErrUnavailable
	}
	return store.driver.ContactStatsRead(ctx, access, observedAt)
}
