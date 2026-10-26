package recoveryorchestration

import (
	"context"
	"fmt"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	restoreorchestrationv1 "github.com/Onurlulardan/harpy/contracts/restoreorchestration/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func loadRestoreAuthority(
	ctx context.Context,
	artifacts RestoreArtifacts,
	storage RestoreStorage,
	trust recoverysigningv1.Bundle,
	orchestrator restoreorchestrationv1.Trust,
	now time.Time,
) (databaseRestoreAuthority, error) {
	if artifacts.Branch == RestoreRehydrate {
		return loadRehydrateRestoreAuthority(
			ctx, artifacts, storage, trust, orchestrator, now,
		)
	}
	if artifacts.Branch != RestoreExistingReadable {
		return databaseRestoreAuthority{}, fmt.Errorf("restore branch is invalid")
	}
	authority, err := loadExistingRestoreAuthority(
		artifacts, trust, orchestrator, now,
	)
	if err != nil {
		return databaseRestoreAuthority{}, err
	}
	manifest, err := discoverManifest(
		ctx, storage, authority.plan.Identity.EnvironmentID,
		authority.plan.RecoverySetID, authority.trust,
	)
	if err != nil || manifest.artifact.SHA256() != authority.plan.ManifestSHA256 {
		authority.clear()
		return databaseRestoreAuthority{},
			fmt.Errorf("restore manifest discovery differs")
	}
	authority.manifest = manifest
	return authority, nil
}

func loadRestoreKeys(
	artifacts RestoreArtifacts,
	storage RestoreStorage,
	authority *databaseRestoreAuthority,
) (databaseRestoreKeys, error) {
	if artifacts.Branch == RestoreRehydrate {
		return loadRehydrateRestoreKeys(authority)
	}
	providerKind, providerHandle := storage.RecoveryKeyProvider()
	return loadExistingRestoreKeys(
		artifacts.DRKEKKeyring, providerKind, providerHandle,
		authority.manifest.payload,
	)
}

func seedRehydrateControl(
	ctx context.Context,
	execution RestoreExecution,
	authority databaseRestoreAuthority,
) error {
	return execution.SeedRehydrate(ctx, ports.RehydrateControlSeed{
		TrustSHA256:    authority.trust.SHA256(),
		TrustBytes:     authority.trust.CanonicalBytes(),
		PlanSHA256:     authority.planArtifact.SHA256(),
		PlanBytes:      authority.planArtifact.CanonicalBytes(),
		ManifestSHA256: authority.manifest.artifact.SHA256(),
		ManifestBytes:  authority.manifest.artifact.CanonicalBytes(),
	})
}
