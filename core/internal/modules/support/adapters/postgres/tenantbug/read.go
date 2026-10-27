package tenantbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) ReadTenantBugStats(ctx context.Context, access ports.Access,
	tenantID domain.ID,
) (domain.BugStats, error) {
	if !store.ready() {
		return domain.BugStats{}, ports.ErrUnavailable
	}
	return store.driver.BugStatsRead(ctx, access, tenantID)
}
