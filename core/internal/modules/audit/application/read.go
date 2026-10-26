package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) List(
	ctx context.Context, scope platformtx.TenantScope, filter domain.Filter, operation string,
) (domain.Page, error) {
	if operation != OperationList && operation != OperationEntityHistory &&
		operation != OperationFormRecordHistory || filter.Validate() != nil {
		return domain.Page{}, ErrInvalid
	}
	permission := PermissionRead
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operation, &permission)
	if err != nil {
		return domain.Page{}, err
	}
	page, err := store.List(ctx, filter)
	if err != nil {
		return domain.Page{}, err
	}
	if page.Page != filter.Page || page.Limit != filter.Limit || page.Total < 0 ||
		page.TotalPages < 0 || len(page.Entries) > filter.Limit {
		return domain.Page{}, ErrUnavailable
	}
	for _, entry := range page.Entries {
		if entry.Validate() != nil {
			return domain.Page{}, ErrUnavailable
		}
	}
	return page, nil
}

func (application *Application) Stats(
	ctx context.Context, scope platformtx.TenantScope,
) (domain.Stats, error) {
	permission := PermissionRead
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationStats, &permission)
	if err != nil {
		return domain.Stats{}, err
	}
	now, err := application.now()
	if err != nil {
		return domain.Stats{}, err
	}
	stats, err := store.Stats(ctx, now)
	if err != nil {
		return domain.Stats{}, err
	}
	if !validStats(stats) {
		return domain.Stats{}, ErrUnavailable
	}
	return stats, nil
}

func (application *Application) Get(
	ctx context.Context, scope platformtx.TenantScope, id uuid.UUID,
) (domain.Entry, error) {
	entryID := domain.ID(id)
	if !entryID.UUIDv7() {
		return domain.Entry{}, ErrInvalid
	}
	permission := PermissionRead
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationGet, &permission)
	if err != nil {
		return domain.Entry{}, err
	}
	entry, err := store.Get(ctx, entryID)
	if err != nil {
		return domain.Entry{}, err
	}
	if entry.ID != entryID || entry.Validate() != nil {
		return domain.Entry{}, ErrUnavailable
	}
	return entry, nil
}

func validStats(stats domain.Stats) bool {
	if stats.Total < 0 || stats.Last24Hours < 0 || stats.LastHour < 0 ||
		stats.LastHour > stats.Last24Hours || stats.Last24Hours > stats.Total ||
		stats.ByAction == nil || stats.ByEntity == nil || len(stats.Actors) > 10 ||
		len(stats.Entities) > 10 {
		return false
	}
	for _, value := range stats.ByAction {
		if value < 0 || value > stats.Total {
			return false
		}
	}
	for _, value := range stats.ByEntity {
		if value < 0 || value > stats.Total {
			return false
		}
	}
	for _, actor := range stats.Actors {
		if !actor.ActorID.Valid() || actor.Count < 1 {
			return false
		}
	}
	for _, entity := range stats.Entities {
		if entity.Count < 1 || entity.EntityType == "" || entity.EntityID == "" {
			return false
		}
	}
	return true
}
