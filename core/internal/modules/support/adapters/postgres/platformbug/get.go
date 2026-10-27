package platformbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) GetPlatformBug(ctx context.Context, access ports.Access,
	id, tenantID domain.ID,
) (domain.BugReport, error) {
	if !store.ready() {
		return domain.BugReport{}, ports.ErrUnavailable
	}
	return store.driver.BugGet(ctx, access, id, tenantID)
}
