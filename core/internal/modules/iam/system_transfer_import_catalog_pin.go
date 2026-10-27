package iam

import (
	"context"
	"crypto/subtle"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	querycatalog "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferImportCatalogReader interface {
	ActiveSchema(
		context.Context,
		profileapp.ResourceKind,
	) (profileapp.Schema, error)
}

func verifySystemTransferImportCatalogPin(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.Attempt,
) error {
	reader, err := iampostgres.BindProfileReader(scope)
	if err != nil {
		return systemtransfer.ErrWorkerUnavailable
	}
	return verifySystemTransferImportCatalogPinWithReader(ctx, reader, attempt)
}

func verifySystemTransferImportCatalogPinWithReader(
	ctx context.Context,
	reader systemTransferImportCatalogReader,
	attempt systemtransfer.Attempt,
) error {
	if ctx == nil || ctx.Err() != nil || nilIdentityDependency(reader) ||
		!attempt.Resource.Valid() || attempt.CatalogVersionID.Version() != 7 ||
		attempt.CatalogSHA256 == [32]byte{} ||
		attempt.IdentityManifestSHA256 == [32]byte{} {
		return systemtransfer.ErrImportCatalogPinStale
	}
	kind, err := systemTransferImportResourceKind(attempt.Resource)
	if err != nil {
		return systemtransfer.ErrImportCatalogPinStale
	}
	schema, err := reader.ActiveSchema(ctx, kind)
	if err != nil {
		return systemtransfer.ErrWorkerUnavailable
	}
	if schema.ID() != attempt.CatalogVersionID ||
		schema.Resource() != kind ||
		!systemTransferImportCatalogMatches(schema, attempt) ||
		!systemTransferImportManifestMatches(attempt) {
		return systemtransfer.ErrImportCatalogPinStale
	}
	return nil
}

func systemTransferImportCatalogMatches(
	schema profileapp.Schema,
	attempt systemtransfer.Attempt,
) bool {
	custom, err := profileapp.NewQueryCatalog(schema)
	if err != nil {
		return false
	}
	published, err := querycatalog.Build(custom)
	if err != nil {
		return false
	}
	pin, err := systemTransferPublishedPin(published)
	return err == nil && pin.versionID == attempt.CatalogVersionID &&
		subtle.ConstantTimeCompare(pin.sha256[:], attempt.CatalogSHA256[:]) == 1
}

func systemTransferImportManifestMatches(attempt systemtransfer.Attempt) bool {
	manifest, err := systemtransfer.ManifestFor(attempt.Resource)
	return err == nil && subtle.ConstantTimeCompare(
		manifest.Digest[:], attempt.IdentityManifestSHA256[:],
	) == 1
}

func systemTransferImportResourceKind(
	resource publicresource.Resource,
) (profileapp.ResourceKind, error) {
	switch resource {
	case publicresource.Users:
		return profileapp.ResourceUser, nil
	case publicresource.Roles:
		return profileapp.ResourceRole, nil
	case publicresource.Organizations:
		return profileapp.ResourceOrganization, nil
	default:
		return "", systemtransfer.ErrImportCatalogPinStale
	}
}
