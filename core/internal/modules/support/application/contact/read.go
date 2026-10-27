package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) ReadContactStats(ctx context.Context,
	access ports.Access,
) (domain.ContactStats, error) {
	if !service.ready() {
		return domain.ContactStats{}, ports.ErrUnavailable
	}
	return service.kernel.ContactStatsRead(ctx, access)
}
