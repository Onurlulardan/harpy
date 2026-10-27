package recoveryorchestration

import (
	"bytes"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) validateClaim(
	claim ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
) (recoveryset.ManifestPayload, error) {
	if executor == nil || executor.authority == nil ||
		claim.RestoreExecutionIdentity != identity ||
		claim.OperationID != executor.authority.planArtifact.OperationID() ||
		claim.PlanSHA256 != executor.authority.planArtifact.SHA256() ||
		!bytes.Equal(
			claim.PlanBytes, executor.authority.planArtifact.CanonicalBytes(),
		) {
		return recoveryset.ManifestPayload{},
			fmt.Errorf("restore execution authority differs")
	}
	plan := executor.authority.plan
	if claim.RecoverySetID != plan.RecoverySetID ||
		claim.EnvironmentID != plan.Identity.EnvironmentID ||
		claim.TargetDisposition != string(plan.TargetDisposition) ||
		claim.SafetyRecoverySetID != plan.SafetyRecoverySetID ||
		claim.SafetyRecoverySetSHA256 != plan.SafetyRecoverySetSHA256 ||
		claim.WaiverReason != plan.WaiverReason ||
		claim.ProbeSHA256 != plan.ProbeSHA256 ||
		claim.DestructiveConfirmed != plan.DestructiveConfirmed {
		return recoveryset.ManifestPayload{},
			fmt.Errorf("restore execution plan differs")
	}
	artifact, manifest, err := parseRecoveryManifest(
		claim.ManifestBytes, claim.EnvironmentID, claim.RecoverySetID,
		executor.authority.trust,
	)
	if err != nil || artifact.SHA256() != claim.ManifestSHA256 ||
		artifact.SHA256() != plan.ManifestSHA256 ||
		manifest.Identity != plan.Identity ||
		manifest.Portability != recoveryset.SameEnvironment ||
		manifest.DatabaseSystemID != claim.PreviousDatabaseSystemID {
		return recoveryset.ManifestPayload{},
			fmt.Errorf("restore recovery manifest differs")
	}
	if executor.authority.manifest.artifact.SHA256() != "" &&
		(executor.authority.manifest.artifact.SHA256() != artifact.SHA256() ||
			executor.authority.manifest.payload != manifest) {
		return recoveryset.ManifestPayload{},
			fmt.Errorf("rehydrate manifest discovery differs")
	}
	inventory, err := recoveryset.ParseComponentInventory(
		claim.ComponentInventory, true,
	)
	if err != nil || !restoreComponentsMatch(inventory, manifest.Components) {
		return recoveryset.ManifestPayload{},
			fmt.Errorf("restore component inventory differs")
	}
	return manifest, nil
}

func restoreComponentsMatch(
	inventory map[string]recoveryset.ComponentIdentity,
	components recoveryset.RecoveryComponents,
) bool {
	expected := map[string]recoveryset.ComponentIdentity{
		"database_archive":    components.DatabaseArchive,
		"documents_inventory": components.DocumentsInventory,
		"key_custody":         components.KeyCustody,
		"owner_anchor":        components.OwnerAnchor,
		"protected_authority": components.ProtectedAuthority,
	}
	if len(inventory) != len(expected) {
		return false
	}
	for kind, component := range expected {
		if inventory[kind] != component {
			return false
		}
	}
	return true
}
