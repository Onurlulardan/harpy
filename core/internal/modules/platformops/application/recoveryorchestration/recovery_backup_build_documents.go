package recoveryorchestration

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) buildDocumentsComponent(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	snapshotID, prefix string,
	writer RecoveryObjectWriter,
	keys backupcrypto.SealKeys,
	drKeyID, recoveryKeyID string,
) (databaseBackupComponent, string, int64, error) {
	document, digest, count, err := buildDocumentsRecoveryInventory(
		ctx, executor.documents, executor.storage.DocumentsReader(), writer,
		claim.EnvironmentID, executor.target.TenantID(),
		claim.RecoverySetID, snapshotID,
	)
	if err != nil {
		return databaseBackupComponent{}, "", 0, err
	}
	component, err := putEncryptedDocument(
		ctx, writer, componentKey(prefix, "documents_inventory"), document,
		archiveContext(claim, "documents_inventory",
			drKeyID, recoveryKeyID),
		keys,
	)
	clear(document)
	if err != nil {
		return databaseBackupComponent{}, "", 0, err
	}
	component.kind = "documents_inventory"
	return component, digest, count, nil
}
