package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) commandInstallation(ctx context.Context,
	scope platformtx.TenantScope, store Store, identity ScopeIdentity,
	command PackageCommand, action string,
	transition func(catalogdomain.Installation) (catalogdomain.Installation, error),
) (catalogdomain.Installation, error) {
	if !platformtx.ContextOwnsScope(ctx, scope) || store == nil || transition == nil {
		return catalogdomain.Installation{}, ErrUnavailable
	}
	packageID, err := packagecontract.ParsePackageID(command.PackageID.String())
	if err != nil {
		return catalogdomain.Installation{}, ErrInvalid
	}
	installation, err := store.GetInstallation(ctx, packageID)
	if err != nil {
		return catalogdomain.Installation{}, ErrNotFound
	}
	expectedRevision := installation.RowVersion
	installation, err = transition(installation)
	if err != nil {
		return catalogdomain.Installation{}, ErrConflict
	}
	now, err := application.now()
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	digest, err := requestDigest(struct {
		PackageID string `json:"package_id"`
		Action    string `json:"action"`
	}{PackageID: command.PackageID.String(), Action: action})
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	return UpdateInstallation(ctx, store, InstallationMutation{
		Installation: installation, ExpectedRevision: expectedRevision,
		ActorID: identity.ActorID, IdempotencyKey: command.IdempotencyKey,
		RequestSHA256: digest, OccurredAt: now,
	})
}
