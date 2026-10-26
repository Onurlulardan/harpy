package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) BugsList(ctx context.Context, access application.Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	platform, valid := validListAccess(access)
	if !valid {
		return domain.BugPage{}, application.ErrForbidden
	}
	var page domain.BugPage
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, listBugsSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), platform,
			nullableUUID(filter.TenantID), nullableUUID(filter.ReporterID),
			nullableText(string(filter.Status)), nullableText(string(filter.Severity)),
			nullableText(string(filter.Category)), filter.Page.Number, filter.Page.Size).
			Scan(&document, &page.Total); err != nil {
			return err
		}
		items, err := decodeBugPage(document)
		page.Items, page.Page = items, filter.Page
		return err
	})
	return page, store.normalize(err)
}

func validListAccess(access application.Access) (bool, bool) {
	switch access.Scope() {
	case application.ScopeReporter:
		return false, access.Allows(application.ScopeReporter,
			application.PermissionBugReadOwn, application.OperationBugMine)
	case application.ScopeTenant:
		return false, access.Allows(application.ScopeTenant,
			application.PermissionBugRead, application.OperationBugTenantList)
	case application.ScopePlatform:
		return true, access.Allows(application.ScopePlatform,
			application.PermissionPlatformBugRead, application.OperationBugPlatformList)
	default:
		return false, false
	}
}
