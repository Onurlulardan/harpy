package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/find_idempotent_operation.sql
var findIdempotentOperationSQL string

//go:embed sql/tenant/start_installation.sql
var startInstallationSQL string

//go:embed sql/tenant/verify_installation.sql
var verifyInstallationSQL string

//go:embed sql/tenant/create_operation.sql
var createOperationSQL string

func (storage *store) CreateOperation(
	ctx context.Context, mutation catalogapp.StartMutation,
) (catalogapp.OperationHandle, error) {
	if !validStartMutation(storage, mutation) {
		return catalogapp.OperationHandle{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return catalogapp.OperationHandle{}, unavailable()
	}
	if replay, found, err := storage.findReplay(ctx, mutation); err != nil || found {
		return replay, err
	}
	installationID, err := storage.prepareInstallation(ctx, mutation)
	if err != nil {
		return catalogapp.OperationHandle{}, err
	}
	operationID, _ := uuid.Parse(mutation.Operation.ID.String())
	planDocument, err := encodePlan(mutation.Operation.Plan)
	if err != nil {
		return catalogapp.OperationHandle{}, err
	}
	value, returnedInstallation, err := scanOperation(storage.queryRow(ctx, createOperationSQL,
		operationID, installationID, mutation.Operation.Plan.PackageID.String(), mutation.Operation.Kind,
		mutation.Operation.State, planDocument, mutation.Operation.Plan.SHA256,
		mutation.PreviousOperation, mutation.ActorID, mutation.IdempotencyKey,
		mutation.RequestSHA256, mutation.Operation.StartedAt, mutation.Operation.Revision))
	if err != nil || returnedInstallation != installationID || value.ID != mutation.Operation.ID ||
		value.Plan.SHA256 != mutation.Operation.Plan.SHA256 {
		return catalogapp.OperationHandle{}, unavailable()
	}
	if err = storage.authorityEvidence(ctx, mutation); err != nil {
		return catalogapp.OperationHandle{}, err
	}
	return catalogapp.OperationHandle{Operation: value}, nil
}

func (storage *store) findReplay(
	ctx context.Context, mutation catalogapp.StartMutation,
) (catalogapp.OperationHandle, bool, error) {
	var operationID uuid.UUID
	var requestSHA256 string
	err := storage.queryRow(ctx, findIdempotentOperationSQL,
		mutation.ActorID, mutation.IdempotencyKey).Scan(&operationID, &requestSHA256)
	if err != nil {
		if storage.notFound(err) {
			return catalogapp.OperationHandle{}, false, nil
		}
		return catalogapp.OperationHandle{}, false, unavailable()
	}
	if requestSHA256 != mutation.RequestSHA256 {
		return catalogapp.OperationHandle{}, true, catalogapp.ErrConflict
	}
	value, err := storage.loadOperation(ctx, operationID)
	if err != nil || value.Plan.PackageID.String() != mutation.Operation.Plan.PackageID.String() {
		return catalogapp.OperationHandle{}, true, unavailable()
	}
	if mutation.PurgeAuthority != nil {
		stored, authorityErr := storage.AuthorityForPurge(ctx, operationID)
		if authorityErr != nil {
			return catalogapp.OperationHandle{}, true, authorityErr
		}
		if stored != *mutation.PurgeAuthority {
			return catalogapp.OperationHandle{}, true, catalogapp.ErrConflict
		}
	}
	return catalogapp.OperationHandle{Operation: value, Replayed: true}, true, nil
}

func (storage *store) prepareInstallation(
	ctx context.Context, mutation catalogapp.StartMutation,
) (uuid.UUID, error) {
	installationID, _ := uuid.Parse(mutation.Installation.ID.String())
	if !mutation.MutateInstallation {
		var returned uuid.UUID
		err := storage.queryRow(ctx, verifyInstallationSQL, installationID,
			mutation.Installation.PackageID.String(), mutation.Installation.RowVersion).Scan(&returned)
		if err != nil || returned != installationID {
			return uuid.Nil, unavailable()
		}
		return returned, nil
	}
	var activeVersion any
	if mutation.Installation.ActiveVersion.Valid() {
		activeVersion = mutation.Installation.ActiveVersion.String()
	}
	var previousState any
	if mutation.PreviousState != nil {
		previousState = string(*mutation.PreviousState)
	}
	var returned uuid.UUID
	err := storage.queryRow(ctx, startInstallationSQL, installationID,
		mutation.Installation.PackageID.String(), activeVersion, mutation.Installation.State,
		mutation.Installation.ReferenceEpoch, mutation.Installation.RowVersion, mutation.OccurredAt,
		previousState, mutation.Installation.RowVersion-1).Scan(&returned)
	if err != nil || returned != installationID {
		return uuid.Nil, unavailable()
	}
	return returned, nil
}

func validStartMutation(storage *store, value catalogapp.StartMutation) bool {
	valid := storage != nil && value.Installation.ID.Valid() && value.Installation.PackageID.Valid() &&
		value.Operation.ID.Valid() && value.Operation.Plan.Valid() && value.Operation.State == "PLANNED" &&
		value.Operation.Plan.PackageID.String() == value.Installation.PackageID.String() &&
		value.ActorID == storage.actorID && idempotencyPattern.MatchString(value.IdempotencyKey) &&
		digestPattern.MatchString(value.RequestSHA256) && canonicalTime(value.OccurredAt) &&
		value.Operation.StartedAt.Equal(value.OccurredAt)
	if !valid {
		return false
	}
	if value.Operation.Kind != "PURGE" {
		return value.PurgeAuthority == nil
	}
	if value.PurgeAuthority == nil || !value.PurgeAuthority.Valid() {
		return false
	}
	authority := *value.PurgeAuthority
	return authority.TenantID == storage.tenantID.String() &&
		authority.PackageID == value.Installation.PackageID.String() &&
		authority.PackageSlug == value.Operation.Plan.Slug &&
		authority.OperationID == value.Operation.ID.String() &&
		authority.ActorID == value.ActorID.String() &&
		authority.IdempotencyKey == value.IdempotencyKey &&
		authority.ReferenceEpoch == value.Installation.ReferenceEpoch &&
		!authority.RequestedAt.After(value.Operation.StartedAt)
}
