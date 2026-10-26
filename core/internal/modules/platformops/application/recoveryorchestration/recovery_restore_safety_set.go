package recoveryorchestration

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) verifySafetySet(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	restoreManifest recoveryset.ManifestPayload,
) error {
	if claim.TargetDisposition != string(recoveryset.TargetExistingReadable) {
		return nil
	}
	safety, err := discoverManifest(
		ctx, executor.storage, claim.EnvironmentID, claim.SafetyRecoverySetID,
		executor.authority.trust,
	)
	if err != nil {
		return dependencyRestore(err)
	}
	if safety.artifact.SHA256() != claim.SafetyRecoverySetSHA256 ||
		safety.payload.Identity != restoreManifest.Identity ||
		safety.payload.DatabaseSystemID != claim.PreviousDatabaseSystemID ||
		!safety.payload.CreatedAt.After(restoreManifest.CreatedAt) {
		return terminalRestore(fmt.Errorf("restore safety recovery set differs"))
	}
	return nil
}
