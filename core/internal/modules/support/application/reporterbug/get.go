package reporterbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) GetMyBug(ctx context.Context, access ports.Access,
	id domain.ID,
) (domain.BugReport, error) {
	if !service.ready() {
		return domain.BugReport{}, ports.ErrUnavailable
	}
	return service.kernel.ReporterBugGet(ctx, access, id)
}
