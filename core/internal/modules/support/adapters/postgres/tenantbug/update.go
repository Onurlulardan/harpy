package tenantbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) UpdateTenantBug(ctx context.Context, access ports.Access,
	update domain.BugUpdate,
) (domain.BugReport, bool, error) {
	if !store.ready() {
		return domain.BugReport{}, false, ports.ErrUnavailable
	}
	return store.driver.BugUpdate(ctx, access, update)
}
