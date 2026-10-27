package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) BugCreate(ctx context.Context,
	command application.BugSubmission,
) (domain.BugReport, bool, error) {
	access := command.Access
	if !access.Allows(application.ScopeReporter, application.PermissionBugCreate,
		application.OperationBugCreate) {
		return domain.BugReport{}, false, application.ErrForbidden
	}
	var report domain.BugReport
	var replayed bool
	err := store.within(ctx, CommandTransaction, func(database Database) error {
		var document []byte
		err := database.QueryRow(ctx, createBugSQL, uuid.UUID(command.ID), uuid.UUID(access.TenantID()),
			access.TenantSlug(), access.TenantName(), uuid.UUID(access.ActorID()),
			access.ReporterName(), access.ReporterEmail(), access.AuthorizationVersion(),
			access.AuthorityClock(), command.Input.Title, command.Input.Description,
			string(command.Input.Category), string(command.Input.Severity),
			command.Input.PageURL, command.Input.UserAgent, string(command.Input.BrowserInfo),
			string(command.Input.ConsoleErrors), string(command.Input.NetworkInfo),
			command.Input.StepsToReproduce, command.IdempotencyKey,
			command.RequestSHA256[:], command.ObservedAt).Scan(&document, &replayed)
		if err != nil {
			return err
		}
		report, err = decodeBug(document)
		return err
	})
	return report, replayed, store.normalize(err)
}
