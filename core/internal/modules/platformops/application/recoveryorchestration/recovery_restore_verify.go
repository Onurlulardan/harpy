package recoveryorchestration

import (
	"context"
	"fmt"

	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) verifyAndRehydrate(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	_ ports.RestoreExecutionIdentity,
	manifest recoveryset.ManifestPayload,
	documents documentsapi.Recovery,
) (string, string, string, error) {
	keyCustody, err := executor.readRestoreComponent(
		ctx, "key_custody", manifest.Components.KeyCustody,
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return "", "", "", terminalRestore(err)
	}
	defer clear(keyCustody)
	var custody restoreKeyCustodyEvidence
	if err = parseCanonicalRestoreEvidence(keyCustody, &custody); err != nil ||
		executor.validateKeyCustodyEvidence(
			custody, claim.EnvironmentID, claim.RecoverySetID,
		) != nil {
		return "", "", "", terminalRestore(
			fmt.Errorf("recovery key custody verification failed"),
		)
	}
	owner, err := executor.readRestoreComponent(
		ctx, "owner_anchor", manifest.Components.OwnerAnchor,
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return "", "", "", terminalRestore(err)
	}
	defer clear(owner)
	var anchor restoreOwnerAnchorEvidence
	if parseCanonicalRestoreEvidence(owner, &anchor) != nil ||
		validateOwnerAnchorEvidence(
			anchor, manifest.Identity, claim.RecoverySetID,
		) != nil {
		return "", "", "", terminalRestore(
			fmt.Errorf("owner anchor verification failed"),
		)
	}
	authority, err := executor.readRestoreComponent(
		ctx, "protected_authority", manifest.Components.ProtectedAuthority,
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return "", "", "", terminalRestore(err)
	}
	defer clear(authority)
	var protected restoreProtectedAuthorityEvidence
	if parseCanonicalRestoreEvidence(authority, &protected) != nil ||
		validateProtectedAuthorityEvidence(
			protected, manifest.Identity, claim.RecoverySetID,
			manifest.RecoveryTrustBundleSHA256,
		) != nil {
		return "", "", "", terminalRestore(
			fmt.Errorf("protected authority verification failed"),
		)
	}
	inventoryDocument, err := executor.readRestoreComponent(
		ctx, "documents_inventory", manifest.Components.DocumentsInventory,
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return "", "", "", terminalRestore(err)
	}
	defer clear(inventoryDocument)
	inventory, items, err := loadDocumentsRecoveryInventory(
		inventoryDocument, claim.EnvironmentID, claim.RecoverySetID,
		manifest.Identity.PrimaryBindingID, documents,
	)
	if err != nil ||
		inventory.InventorySHA256 != manifest.DocumentsMerkleSHA256 {
		return "", "", "", terminalRestore(
			fmt.Errorf("documents recovery inventory verification failed"),
		)
	}
	remap, err := executor.rehydrateDocuments(
		ctx, claim, inventory, items, documents,
	)
	if err != nil {
		return "", "", "", err
	}
	if _, err = documents.TerminalizeTransients(
		ctx, claim.OperationID,
	); err != nil {
		return "", "", "", terminalRestore(err)
	}
	safety, err := restoreSafetyDigest(claim)
	if err != nil {
		return "", "", "", terminalRestore(err)
	}
	return manifest.DocumentsMerkleSHA256, remap, safety, nil
}
