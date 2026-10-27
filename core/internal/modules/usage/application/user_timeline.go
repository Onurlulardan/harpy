package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) UserTimeline(ctx context.Context, access Access, subject string,
	page, limit int,
) (domain.TimelinePage, error) {
	if err := service.ready(ctx); err != nil {
		return domain.TimelinePage{}, err
	}
	if err := authorize(access, OperationTenantTimeline, OperationPlatformTimeline); err != nil {
		return domain.TimelinePage{}, err
	}
	query, err := domain.NewTimelineQuery(access.Scope(), subject, page, limit)
	if err != nil {
		return domain.TimelinePage{}, err
	}
	value, err := service.store.UserTimeline(ctx, access, query)
	return decorateTimeline(value), err
}
