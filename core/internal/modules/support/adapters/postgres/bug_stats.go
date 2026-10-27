package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) BugStatsRead(ctx context.Context, access application.Access,
	tenantID domain.ID,
) (domain.BugStats, error) {
	platform := access.Scope() == application.ScopePlatform
	if !validStatsAccess(access) {
		return domain.BugStats{}, application.ErrForbidden
	}
	var stats domain.BugStats
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, bugStatsSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), platform,
			nullableUUID(tenantID)).Scan(&document); err != nil {
			return err
		}
		var err error
		stats, err = decodeBugStats(document)
		return err
	})
	return stats, store.normalize(err)
}

func validStatsAccess(access application.Access) bool {
	return access.Allows(application.ScopeTenant, application.PermissionBugRead,
		application.OperationBugTenantStats) || access.Allows(application.ScopePlatform,
		application.PermissionPlatformBugRead, application.OperationBugPlatformStats)
}
