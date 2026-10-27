package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (store *Store) Overview(ctx context.Context, access application.Access,
	window domain.TimeRange, now time.Time,
) (domain.Overview, error) {
	if !allows(access, application.OperationTenantOverview, application.OperationPlatformOverview) {
		return domain.Overview{}, domain.ErrForbidden
	}
	document, err := store.read(ctx, overviewSQL,
		withArguments(access, window.Start, window.End, now)...)
	if err != nil {
		return domain.Overview{}, err
	}
	return decodeOverview(document, access.Scope())
}

func (store *Store) ActiveUsers(ctx context.Context, access application.Access,
	now time.Time,
) (domain.ActiveUsers, error) {
	if !allows(access, application.OperationTenantActiveUsers,
		application.OperationPlatformActiveUsers) {
		return domain.ActiveUsers{}, domain.ErrForbidden
	}
	document, err := store.read(ctx, activeUsersSQL, withArguments(access, now)...)
	if err != nil {
		return domain.ActiveUsers{}, err
	}
	return decodeActiveUsers(document)
}

func (store *Store) TopUsers(ctx context.Context, access application.Access,
	query domain.RankedQuery,
) ([]domain.TopUser, error) {
	if !allows(access, application.OperationTenantTopUsers, application.OperationPlatformTopUsers) {
		return nil, domain.ErrForbidden
	}
	document, err := store.read(ctx, topUsersSQL,
		withArguments(access, query.Range.Start, query.Range.End, query.Limit)...)
	if err != nil {
		return nil, err
	}
	return decodeTopUsers(document, access.Scope())
}

func (store *Store) UserSummary(ctx context.Context, access application.Access,
	subject string, window domain.TimeRange,
) (domain.UserSummary, error) {
	if !allows(access, application.OperationTenantUserSummary,
		application.OperationPlatformUserSummary) {
		return domain.UserSummary{}, domain.ErrForbidden
	}
	if _, err := domain.NewTimelineQuery(access.Scope(), subject, 1, 1); err != nil {
		return domain.UserSummary{}, err
	}
	document, err := store.read(ctx, userSummarySQL,
		withArguments(access, subject, window.Start, window.End)...)
	if err != nil {
		return domain.UserSummary{}, err
	}
	return decodeUserSummary(document, access.Scope())
}
