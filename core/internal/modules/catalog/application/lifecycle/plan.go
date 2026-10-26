package lifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) buildPlan(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	manifest packagecontract.Manifest,
) (catalogdomain.Plan, error) {
	return application.buildPlanWithTargets(
		ctx, scope, store, manifest, false,
	)
}

func (application *Application) buildProvisioningPlan(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	manifest packagecontract.Manifest,
) (catalogdomain.Plan, error) {
	return application.buildPlanWithTargets(
		ctx, scope, store, manifest, true,
	)
}

func (application *Application) buildPlanWithTargets(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	manifest packagecontract.Manifest,
	provision bool,
) (catalogdomain.Plan, error) {
	resolved, err := application.resolveDependencies(ctx, scope, manifest)
	if err != nil {
		return catalogdomain.Plan{}, err
	}
	dependencies := make([]catalogdomain.DependencyPin, 0, len(resolved))
	for _, item := range resolved {
		installed, err := store.GetInstallation(ctx, item.PackageID)
		if err != nil || installed.ActiveVersion.Compare(item.Version) != 0 ||
			(installed.State != catalogdomain.InstallationActive &&
				installed.State != catalogdomain.InstallationDisabled) {
			return catalogdomain.Plan{}, ErrDependencyBlocked
		}
		dependencies = append(dependencies, catalogdomain.DependencyPin{
			PackageID: item.PackageID, Version: item.Version,
			Required: item.Kind == packagecontract.DependencyRequired,
		})
	}
	components := make([]catalogdomain.ComponentPin, 0, len(manifest.Components))
	for _, component := range manifest.Components {
		owner, err := application.owners.Resolve(component.Owner)
		if err != nil || owner == nil {
			return catalogdomain.Plan{}, ErrOwnerNotV2
		}
		identity := componentowner.Component{ID: component.ID.String(), Owner: component.Owner,
			Contract: component.Contract, ContentSHA256: component.ContentHash}
		generation := int64(0)
		if provision {
			provisioner, available := owner.(componentowner.Provisioner)
			if !available {
				return catalogdomain.Plan{}, ErrOwnerNotV2
			}
			target, targetErr := provisioner.Target(ctx, scope, identity)
			if targetErr != nil || !componentTargetMatches(target, identity) {
				return catalogdomain.Plan{}, ErrOwnerNotV2
			}
			generation = target.Generation
		} else {
			readiness, readinessErr := owner.Readiness(ctx, scope, identity)
			if readinessErr != nil || !readiness.WriterV2 ||
				!componentTargetMatches(readiness.Component, identity) {
				return catalogdomain.Plan{}, ErrOwnerNotV2
			}
			generation = readiness.Generation
		}
		components = append(components, catalogdomain.ComponentPin{
			ComponentID: component.ID, Owner: component.Owner,
			OwnerGeneration: generation, Contract: component.Contract,
			ContentSHA256: component.ContentHash,
		})
	}
	menu := make([]catalogdomain.MenuPin, 0, len(manifest.Menu))
	for _, item := range manifest.Menu {
		menu = append(menu, catalogdomain.MenuPin{ItemID: item.ID, ParentID: item.ParentID,
			Kind: item.Kind, TargetID: item.TargetID, Label: item.Label, Order: item.Order, URL: item.URL})
	}
	manifestDigest := sha256.Sum256(manifest.Canonical())
	return catalogdomain.NewSealedPlan(
		manifest.PackageID, manifest.Version, manifest.Slug, hex.EncodeToString(manifestDigest[:]),
		dependencies, components, menu,
	)
}

func componentTargetMatches(
	target, identity componentowner.Component,
) bool {
	return target.Generation > 0 &&
		target.ID == identity.ID && target.Owner == identity.Owner &&
		target.Contract == identity.Contract &&
		target.ContentSHA256 == identity.ContentSHA256
}
