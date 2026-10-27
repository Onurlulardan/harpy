package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) BugUpdate(ctx context.Context, access application.Access,
	update domain.BugUpdate,
) (domain.BugReport, bool, error) {
	platform, valid := updateBugAccess(access)
	if !valid || update.ActorID != domain.ID(access.ActorID()) {
		return domain.BugReport{}, false, application.ErrForbidden
	}
	var report domain.BugReport
	var replayed bool
	err := store.within(ctx, CommandTransaction, func(database Database) error {
		var document []byte
		err := database.QueryRow(ctx, updateBugSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), platform,
			uuid.UUID(update.TenantID), uuid.UUID(update.ID), update.Status != nil,
			bugStatusValue(update.Status), update.Priority != nil, bugPriorityValue(update.Priority),
			update.Resolution != nil, stringValue(update.Resolution),
			update.AdminNotes != nil, stringValue(update.AdminNotes),
			update.ExpectedRevision, update.IdempotencyKey, update.RequestSHA256[:],
			update.ObservedAt).Scan(&document, &replayed)
		if err != nil {
			return err
		}
		report, err = decodeBug(document)
		return err
	})
	return report, replayed, store.normalize(err)
}

func updateBugAccess(access application.Access) (bool, bool) {
	if access.Allows(application.ScopeTenant, application.PermissionBugManage,
		application.OperationBugTenantUpdate) {
		return false, true
	}
	return true, access.Allows(application.ScopePlatform,
		application.PermissionPlatformBugManage, application.OperationBugPlatformUpdate)
}

func bugStatusValue(value *domain.BugStatus) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

func bugPriorityValue(value *domain.BugPriority) any {
	if value == nil {
		return nil
	}
	return string(*value)
}
