package profilepostgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

func (storage *store) PreparePatch(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	recordID identitydomain.IdentityID,
	patch []byte,
) (identitydomain.ProfileDocument, error) {
	prepared, err := storage.preparePatchWrite(
		ctx, scope, resource, recordID, patch,
	)
	if err != nil {
		return identitydomain.ProfileDocument{}, err
	}
	return newProfileDocument(resource, prepared.target)
}

func (storage *store) preparePatchWrite(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	recordID identitydomain.IdentityID,
	patch []byte,
) (preparedProfileWrite, error) {
	if !scopeUsable(ctx, scope, commandMode) || !resource.Valid() ||
		recordID == (identitydomain.IdentityID{}) {
		return preparedProfileWrite{}, identitydomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return preparedProfileWrite{}, identitydomain.ErrStorage
	}
	contract, err := storage.loadWriteContract(ctx, bound, resource)
	if err != nil {
		return preparedProfileWrite{}, err
	}
	current, err := storage.loadWriteResource(
		ctx, bound, contract, resource, databaseID(recordID),
	)
	if err != nil {
		return preparedProfileWrite{}, err
	}
	if contract.migrating() {
		return storage.prepareMigratingPatch(
			ctx, bound, contract, current, resource, recordID, patch,
		)
	}
	source, err := sourceProfileWrite(contract, current)
	if err != nil {
		return preparedProfileWrite{}, identitydomain.ErrStorage
	}
	patched, err := profileapp.Patch(
		contract.source, source, patch, profileNormalizer{},
	)
	if err != nil {
		return preparedProfileWrite{}, identitydomain.ErrInvalidCommand
	}
	return preparedProfileWrite{
		expectedSourceSchemaID: current.actualSchemaID,
		sourceSchema:           contract.source, target: patched,
	}, nil
}

func (storage *store) prepareMigratingPatch(
	ctx context.Context,
	bound *boundStore,
	contract profileWriteContract,
	current profileWriteResource,
	resource identitydomain.ProfileResource,
	recordID identitydomain.IdentityID,
	patch []byte,
) (preparedProfileWrite, error) {
	planned, err := migrationTargetForWrite(contract, current)
	if err != nil {
		return preparedProfileWrite{}, identitydomain.ErrInvalidCommand
	}
	target, err := profileapp.UpgradeMigratedDocument(
		contract.source, contract.target, contract.mappings,
		planned, patch, profileNormalizer{},
	)
	if err != nil {
		return preparedProfileWrite{}, identitydomain.ErrInvalidCommand
	}
	if current.actualSchemaID != contract.target.ID() &&
		reserveProfilePreconversion(
			ctx, bound, contract, resource, databaseID(recordID),
			current, planned, target,
		) != nil {
		return preparedProfileWrite{}, identitydomain.ErrStorage
	}
	return preparedProfileWrite{
		expectedSourceSchemaID: current.actualSchemaID,
		sourceSchema:           contract.source, target: target,
	}, nil
}

func sourceProfileWrite(
	contract profileWriteContract,
	current profileWriteResource,
) (profileapp.Document, error) {
	if !contract.migrating() || current.actualSchemaID != contract.target.ID() {
		if current.actualSchemaID == contract.source.ID() {
			return current.document, nil
		}
		wire := current.document.CanonicalJSON()
		defer clear(wire)
		digest := current.document.SHA256()
		return profileapp.ProjectLineageDocument(
			current.actualSchema, contract.source, wire, digest[:], profileNormalizer{},
		)
	}
	wire := current.document.CanonicalJSON()
	defer clear(wire)
	digest := current.document.SHA256()
	return profileapp.ProjectMigrationSource(
		contract.source, contract.target, contract.mappings,
		wire, digest[:], profileNormalizer{},
	)
}
