package catalog

import (
	"context"
	"fmt"
	"sort"
	"time"

	registryadapter "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/registry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	releasepublication "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/releasepublication"
	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type catalogReleaseFinalizer struct {
	release  attestation.VerifiedRelease
	keys     releasepublication.KeyResolver
	versions []packageregistry.SealedVersion
}

type ReleasePackage = packageregistry.SealedVersion
type ReleasePackageKeyResolver = releasepublication.KeyResolver

func NewReleaseFinalizerFromPackageSet(
	release attestation.VerifiedRelease,
	keys ReleasePackageKeyResolver,
	set attestation.VerifiedCatalogPackageSet,
) (platformmigrator.ReleaseFinalizer, error) {
	value := set.Value()
	if !set.IsVerified() || value.ReleaseID == "" ||
		value.ReleaseID != release.Value().ReleaseID {
		return nil, fmt.Errorf("catalog package set binding failed")
	}
	versions, err := releasepublication.PackageVersionsFromReleaseSet(value)
	if err != nil {
		return nil, err
	}
	return NewReleaseFinalizer(release, keys, versions)
}

func NewReleaseFinalizer(
	release attestation.VerifiedRelease,
	keys ReleasePackageKeyResolver,
	versions []ReleasePackage,
) (platformmigrator.ReleaseFinalizer, error) {
	if release.Value().ReleaseID == "" || release.IssuedAt().IsZero() ||
		keys == nil || len(versions) == 0 {
		return nil, fmt.Errorf("catalog release publication evidence is incomplete")
	}
	copied := make([]packageregistry.SealedVersion, len(versions))
	for index := range versions {
		copied[index] = versions[index].Clone()
	}
	sort.Slice(copied, func(left, right int) bool {
		if copied[left].PackageID == copied[right].PackageID {
			return copied[left].Version < copied[right].Version
		}
		return copied[left].PackageID < copied[right].PackageID
	})
	return &catalogReleaseFinalizer{
		release: release, keys: keys, versions: copied,
	}, nil
}

func (finalizer *catalogReleaseFinalizer) FinalizeRelease(
	ctx context.Context,
	queryRow platformmigrator.FinalizationQueryRow,
) error {
	if finalizer == nil || queryRow == nil {
		return packageregistry.ErrUnavailable
	}
	registrar, err := registryadapter.NewDeployRegistrar(
		func(
			callContext context.Context,
			statement string,
			arguments ...any,
		) registryadapter.DeployRow {
			return queryRow(callContext, statement, arguments...)
		},
		platformmigrator.FinalizationRowNotFound,
	)
	if err != nil {
		return err
	}
	application, err := releasepublication.NewApplication(registrar, finalizer.keys)
	if err != nil {
		return err
	}
	release := finalizer.release.Value()
	registeredAt := finalizer.release.IssuedAt().Truncate(time.Microsecond)
	for _, source := range finalizer.versions {
		version := source.Clone()
		version.ReleaseID = release.ReleaseID
		version.RegisteredAt = registeredAt
		if err := application.Register(ctx, version); err != nil {
			return err
		}
	}
	return nil
}
