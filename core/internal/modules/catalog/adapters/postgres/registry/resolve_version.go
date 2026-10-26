package registry

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

//go:embed sql/platform/get_version.sql
var getVersionSQL string

//go:embed sql/platform/get_version_payloads.sql
var getVersionPayloadsSQL string

func (storage *store) ResolveVersion(
	ctx context.Context, packageID, version string,
) (packageregistry.SealedVersion, error) {
	if !validPackageID(packageID) {
		return packageregistry.SealedVersion{}, packageregistry.ErrInvalid
	}
	if parsed, err := packagecontract.ParseVersion(version); err != nil || parsed.String() != version {
		return packageregistry.SealedVersion{}, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	value, err := scanVersion(storage.queryRow(ctx, getVersionSQL, packageID, version))
	if err != nil {
		if storage.notFound(err) {
			return packageregistry.SealedVersion{}, packageregistry.ErrNotFound
		}
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	if !validVersion(value) || value.PackageID != packageID || value.Version != version {
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	return value, nil
}

func (storage *store) VersionPayloads(
	ctx context.Context, packageID, version string,
) (packageregistry.SealedVersion, error) {
	if !validPackageID(packageID) {
		return packageregistry.SealedVersion{}, packageregistry.ErrInvalid
	}
	if parsed, err := packagecontract.ParseVersion(version); err != nil ||
		parsed.String() != version {
		return packageregistry.SealedVersion{}, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	value, err := scanPublishedVersion(
		storage.queryRow(ctx, getVersionPayloadsSQL, packageID, version),
	)
	if err != nil {
		if storage.notFound(err) {
			return packageregistry.SealedVersion{}, packageregistry.ErrNotFound
		}
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	if !validRegistration(value) ||
		value.PackageID != packageID || value.Version != version {
		return packageregistry.SealedVersion{}, packageregistry.ErrUnavailable
	}
	return value, nil
}

var _ packageregistry.PayloadReader = (*store)(nil)
