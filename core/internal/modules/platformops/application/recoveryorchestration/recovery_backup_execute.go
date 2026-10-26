package recoveryorchestration

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	"github.com/google/uuid"
)

type databaseBackupBuild struct {
	components      map[string]databaseBackupComponent
	documentsDigest string
	documentsCount  int64
}

func (executor *databaseBackupExecutor) execute(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	identity ports.RecoveryExecutionIdentity,
) (err error) {
	_, parseErr := uuid.Parse(claim.RecoverySetID)
	if parseErr != nil || executor.validateClaim(claim) != nil {
		return fmt.Errorf("database backup claim binding differs")
	}
	defer func() {
		if err != nil && ctx.Err() == nil {
			_ = executor.documents.FailHold(
				context.WithoutCancel(ctx), claim.RecoverySetID,
				"backup_execution_failed",
			)
		}
	}()
	inventory, err := recoveryset.ParseComponentInventory(
		claim.ComponentInventory, claim.RecoveryState == "verifying",
	)
	if err != nil || len(inventory) != 0 && len(inventory) != 5 {
		return fmt.Errorf("database backup component inventory is partial")
	}
	providerKind, providerHandle := executor.storage.RecoveryKeyProvider()
	material, err := loadDatabaseBackupKeyMaterial(
		providerKind, providerHandle,
		claim.EnvironmentID, claim.SigningGeneration,
	)
	if err != nil {
		return err
	}
	defer material.Clear()
	if len(inventory) == 5 {
		return executor.resumeComplete(ctx, claim, identity, material, inventory)
	}
	if claim.RecoveryState == "verifying" {
		return fmt.Errorf("database backup verifying state has no components")
	}
	if err = executor.documents.BeginHold(ctx, claim.RecoverySetID); err != nil {
		return err
	}
	if claim.RecoveryState == "holding" {
		if err = executor.snapshot.Start(ctx, identity); err != nil {
			return err
		}
	}
	build, err := executor.buildComponents(ctx, claim, material)
	if err != nil {
		return err
	}
	if claim.RecoveryState != "copying" {
		if err = executor.copy.Start(ctx, identity); err != nil {
			return err
		}
	}
	if err = executor.documents.CompleteHold(
		ctx, claim.RecoverySetID,
		decodeDigest32(build.documentsDigest), build.documentsCount,
	); err != nil {
		return err
	}
	if err = executor.execution.AttestDocuments(
		ctx, ports.DocumentsHoldAttestation{
			RecoveryExecutionIdentity: identity,
			InventorySHA256:           build.documentsDigest,
			InventoryCount:            build.documentsCount,
		},
	); err != nil {
		return err
	}
	if err = executor.recordComponentBatch(ctx, identity, build.components); err != nil {
		return err
	}
	if err = executor.verify.Start(ctx, identity); err != nil {
		return err
	}
	return executor.finalize(
		ctx, claim, identity, build.components, build.documentsDigest,
	)
}

func (executor *databaseBackupExecutor) validateClaim(
	claim ports.RecoveryExecutionClaim,
) error {
	targetHash := executor.target.TenantSetSHA256()
	targetDigest := hex.EncodeToString(targetHash[:])
	if executor.target.EnvironmentID() != claim.EnvironmentID ||
		executor.target.TenantID() != claim.PrimaryBindingID ||
		targetDigest != claim.TenantSetSHA256 ||
		executor.trust.EnvironmentID() != claim.EnvironmentID ||
		executor.trust.Generation() != claim.SigningGeneration ||
		executor.trust.SHA256() != claim.RecoveryTrustBundleSHA256 {
		return fmt.Errorf("recovery claim identity differs")
	}
	return nil
}

func decodeDigest32(value string) [32]byte {
	var result [32]byte
	decoded, _ := hex.DecodeString(value)
	copy(result[:], decoded)
	return result
}

func archiveContext(
	claim ports.RecoveryExecutionClaim,
	component string,
	drKeyID string,
	recoveryKeyID string,
) backupcrypto.Context {
	return backupcrypto.Context{
		EnvironmentID: claim.EnvironmentID,
		RecoverySetID: claim.RecoverySetID, ComponentID: component,
		DRKeyID: drKeyID, RecoveryKeyID: recoveryKeyID,
	}
}
