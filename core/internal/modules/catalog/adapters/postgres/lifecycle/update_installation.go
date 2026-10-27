package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/update_installation.sql
var updateInstallationSQL string

func (storage *store) UpdateInstallation(
	ctx context.Context, mutation catalogapp.InstallationMutation,
) (catalogdomain.Installation, error) {
	value := mutation.Installation
	if !validInstallationMutation(storage, mutation) {
		return catalogdomain.Installation{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return catalogdomain.Installation{}, unavailable()
	}
	var activeVersion any
	if value.ActiveVersion.Valid() {
		activeVersion = value.ActiveVersion.String()
	}
	installationID, _ := uuid.Parse(value.ID.String())
	stored, err := scanInstallation(storage.queryRow(ctx, updateInstallationSQL,
		installationID, value.PackageID.String(), activeVersion, value.State,
		value.ReferenceEpoch, value.RowVersion, mutation.OccurredAt, mutation.ExpectedRevision))
	if err != nil || stored.ID != value.ID || stored.PackageID.String() != value.PackageID.String() ||
		stored.RowVersion != value.RowVersion {
		return catalogdomain.Installation{}, unavailable()
	}
	if mutation.OperationID.Valid() {
		if err = storage.applyOperationPlan(ctx, mutation.OperationID.String(), stored,
			mutation.OccurredAt); err != nil {
			return catalogdomain.Installation{}, err
		}
	}
	if len(mutation.OwnerReceipts) > 0 {
		if err = storage.applyCommandReceipts(ctx, mutation, stored); err != nil {
			return catalogdomain.Installation{}, err
		}
	}
	return stored, nil
}

func validInstallationMutation(storage *store, value catalogapp.InstallationMutation) bool {
	return storage != nil && value.Installation.ID.Valid() && value.Installation.PackageID.Valid() &&
		value.ExpectedRevision > 0 && value.Installation.RowVersion == value.ExpectedRevision+1 &&
		value.ActorID == storage.actorID && idempotencyPattern.MatchString(value.IdempotencyKey) &&
		digestPattern.MatchString(value.RequestSHA256) && canonicalTime(value.OccurredAt)
}
