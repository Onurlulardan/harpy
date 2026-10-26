package adoption

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const operationVerify = "catalog.adoption.verify"

func (application *Application) Validate(
	ctx context.Context, scope platformtx.TenantScope, request VerifyRequest,
) (Run, error) {
	if request.RunID.Version() != 7 || request.CutoverGeneration < 1 ||
		!digestPattern.MatchString(request.SourceManifestSHA256) ||
		!digestPattern.MatchString(request.TransformSHA256) {
		return Run{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand,
		operationVerify, "catalog.install")
	if err != nil {
		return Run{}, err
	}
	inventory, err := store.ValidateInventory(ctx)
	if err != nil || inventory.SourceCount < 0 || inventory.TargetCount < 0 ||
		!digestPattern.MatchString(inventory.SourceRootSHA256) ||
		!digestPattern.MatchString(inventory.TargetRootSHA256) {
		return Run{}, ErrUnavailable
	}
	registryBlockers, err := application.validateRegistry(ctx, scope, inventory.RegistryReferences)
	if err != nil {
		return Run{}, err
	}
	inventory.Blockers = append(inventory.Blockers, registryBlockers...)
	now, err := application.now()
	if err != nil {
		return Run{}, err
	}
	state, blocker := "VERIFIED", ""
	verifiedAt := &now
	if len(inventory.Blockers) > 0 {
		sort.Strings(inventory.Blockers)
		state, blocker, verifiedAt = "BLOCKED", strings.Join(inventory.Blockers, ","), nil
		if len(blocker) > 2048 {
			return Run{}, ErrUnavailable
		}
	}
	run := Run{ID: request.RunID, SourceManifestSHA256: request.SourceManifestSHA256,
		TransformSHA256: request.TransformSHA256, SourceCount: inventory.SourceCount,
		TargetCount: inventory.TargetCount, SourceRootSHA256: inventory.SourceRootSHA256,
		TargetRootSHA256: inventory.TargetRootSHA256, State: state, BlockerCode: blocker,
		CutoverGeneration: request.CutoverGeneration, CreatedAt: now, VerifiedAt: verifiedAt}
	stored, err := FinalizeRun(ctx, store, run)
	if err != nil {
		return Run{}, err
	}
	return stored, nil
}

func (application *Application) validateRegistry(ctx context.Context,
	scope platformtx.TenantScope, references []RegistryReference,
) ([]string, error) {
	if len(references) == 0 {
		return nil, nil
	}
	if application == nil || application.registry == nil {
		return nil, ErrUnavailable
	}
	reader, err := application.registry.Bind(ctx, scope)
	if err != nil || reader == nil {
		return nil, ErrUnavailable
	}
	blockers := make([]string, 0)
	for _, reference := range references {
		blocker, validateErr := validateRegistryReference(ctx, reader, reference)
		if validateErr != nil {
			return nil, validateErr
		}
		if blocker != "" {
			blockers = append(blockers, blocker)
		}
	}
	return blockers, nil
}

func validateRegistryReference(ctx context.Context, reader packageregistry.Reader,
	reference RegistryReference,
) (string, error) {
	if ctx == nil || ctx.Err() != nil || reader == nil || reference.Identity == "" {
		return "", ErrUnavailable
	}
	if reference.Kind == "module" {
		return validateModuleReference(ctx, reader, reference)
	}
	if reference.Kind == "manifest" {
		return validateManifestReference(ctx, reader, reference)
	}
	return "", ErrUnavailable
}

func validateModuleReference(ctx context.Context, reader packageregistry.Reader,
	reference RegistryReference,
) (string, error) {
	value, err := reader.GetPackage(ctx, reference.PackageID)
	if missingRegistryRecord(err) || err == nil && value.Slug != reference.Slug {
		return "UNKNOWN_MODULE:" + reference.Identity, nil
	}
	if err != nil {
		return "", ErrUnavailable
	}
	_, err = reader.ResolveVersion(ctx, value.ID, reference.Version)
	if missingRegistryRecord(err) {
		return "UNKNOWN_VERSION:" + reference.Identity, nil
	}
	if err != nil {
		return "", ErrUnavailable
	}
	return "", nil
}

func validateManifestReference(ctx context.Context, reader packageregistry.Reader,
	reference RegistryReference,
) (string, error) {
	value, err := reader.ReadPackageBySlug(ctx, reference.Slug)
	if missingRegistryRecord(err) {
		return "UNKNOWN_MANIFEST_PACKAGE:" + reference.Identity, nil
	}
	if err != nil {
		return "", ErrUnavailable
	}
	version, err := reader.ResolveVersion(ctx, value.ID, reference.Version)
	if missingRegistryRecord(err) {
		return "UNKNOWN_MANIFEST_VERSION:" + reference.Identity, nil
	}
	if err != nil {
		return "", ErrUnavailable
	}
	if !digestPattern.MatchString(reference.ManifestSHA256) ||
		reference.ManifestSHA256 != version.ManifestSHA256 {
		return "MANIFEST_HASH_MISMATCH:" + reference.Identity, nil
	}
	return "", nil
}

func missingRegistryRecord(err error) bool {
	return errors.Is(err, packageregistry.ErrNotFound) ||
		errors.Is(err, packageregistry.ErrInvalid)
}
