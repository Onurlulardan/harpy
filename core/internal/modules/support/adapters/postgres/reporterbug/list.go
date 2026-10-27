package reporterbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) ListMyBugs(ctx context.Context, access ports.Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !store.ready() {
		return domain.BugPage{}, ports.ErrUnavailable
	}
	return store.driver.BugsList(ctx, access, filter)
}
