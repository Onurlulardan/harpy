package query

import (
	"context"

	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PublishedSnapshot struct {
	Snapshot     Snapshot
	Dependencies []DependencySnapshot
}

func (application *Application) ReadPublished(
	ctx context.Context, scope platformtx.TenantScope, queryID uuid.UUID,
) (PublishedSnapshot, error) {
	if !validUUIDv7(queryID) {
		return PublishedSnapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot,
		OperationGetPublished, PermissionRead)
	if err != nil {
		return PublishedSnapshot{}, err
	}
	return application.readPublishedBound(ctx, store, queryID)
}

func (application *Application) readPublishedBound(
	ctx context.Context, store Store, queryID uuid.UUID,
) (PublishedSnapshot, error) {
	snapshot, err := store.LatestPublished(ctx, queryID)
	if err != nil || snapshot.QueryID != queryID || validateSnapshot(snapshot) != nil ||
		(snapshot.VersionState != analyticsquery.VersionPublished &&
			snapshot.VersionState != analyticsquery.VersionNeedsRevalidation) {
		return PublishedSnapshot{}, publishedReadError(err)
	}
	dependencies, err := store.Dependencies(ctx, queryID, snapshot.VersionID)
	if err != nil || len(dependencies) == 0 || len(dependencies) > 32 {
		return PublishedSnapshot{}, ErrUnavailable
	}
	return PublishedSnapshot{Snapshot: snapshot, Dependencies: dependencies}, nil
}

func publishedReadError(err error) error {
	if err != nil {
		return err
	}
	return ErrUnavailable
}
