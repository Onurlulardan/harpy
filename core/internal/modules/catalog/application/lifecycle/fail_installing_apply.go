package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func (application *Application) failInstallingOnDurableStop(
	ctx context.Context,
	store Store,
	identity ScopeIdentity,
	operation catalogdomain.Operation,
) error {
	installation, err := store.GetInstallation(ctx, operation.Plan.PackageID)
	if err != nil {
		return err
	}
	if !shouldFailInstalling(operation, installation.State) {
		return nil
	}
	failed, err := installation.Fail()
	if err != nil {
		return ErrConflict
	}
	now, timeErr := application.now()
	if timeErr != nil {
		return timeErr
	}
	digest, digestErr := requestDigest(struct{ OperationID string }{operation.ID.String()})
	if digestErr != nil {
		return digestErr
	}
	_, err = UpdateInstallation(ctx, store, InstallationMutation{
		Installation: failed, ExpectedRevision: installation.RowVersion,
		ActorID: identity.ActorID,
		IdempotencyKey: operation.ID.String() + ":durable-stop",
		RequestSHA256: digest, OccurredAt: now,
	})
	return err
}
