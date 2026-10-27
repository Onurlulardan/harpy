package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) BugGet(ctx context.Context, access application.Access,
	id, tenantID domain.ID,
) (domain.BugReport, error) {
	reporterID, valid := getBugAccess(access)
	if !valid {
		return domain.BugReport{}, application.ErrForbidden
	}
	var report domain.BugReport
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, getBugSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), uuid.UUID(tenantID),
			nullableUUID(reporterID), uuid.UUID(id)).Scan(&document); err != nil {
			return err
		}
		var err error
		report, err = decodeBug(document)
		return err
	})
	return report, store.normalize(err)
}

func getBugAccess(access application.Access) (domain.ID, bool) {
	if access.Allows(application.ScopeReporter, application.PermissionBugReadOwn,
		application.OperationBugRead) {
		return access.ActorID(), true
	}
	if access.Allows(application.ScopeTenant, application.PermissionBugRead,
		application.OperationBugTenantRead) || access.Allows(application.ScopePlatform,
		application.PermissionPlatformBugRead, application.OperationBugPlatformRead) {
		return domain.NilID, true
	}
	return domain.NilID, false
}
