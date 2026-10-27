package reporterbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) CreateBug(ctx context.Context,
	command ports.BugSubmission,
) (domain.BugReport, bool, error) {
	if !store.ready() {
		return domain.BugReport{}, false, ports.ErrUnavailable
	}
	return store.driver.BugCreate(ctx, command)
}
