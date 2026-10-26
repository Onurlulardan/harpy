package recoveryorchestration

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) resumeComplete(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	identity ports.RecoveryExecutionIdentity,
	material recoverykeymaterial.Material,
	inventory map[string]recoveryset.ComponentIdentity,
) error {
	components := make(map[string]databaseBackupComponent, len(inventory))
	for kind, stored := range inventory {
		component, err := verifyBackupComponent(
			ctx, executor.storage.RecoveryReader(), kind, stored,
		)
		if err != nil {
			return fmt.Errorf("stored recovery component differs")
		}
		components[kind] = component
	}
	documentsDigest, _, err := executor.openDocumentsInventory(
		ctx, claim, material, components["documents_inventory"],
	)
	if err != nil {
		return err
	}
	if claim.RecoveryState == "copying" {
		if err = executor.verify.Start(ctx, identity); err != nil {
			return err
		}
	}
	return executor.finalize(
		ctx, claim, identity, components, documentsDigest,
	)
}

func (executor *databaseBackupExecutor) openDocumentsInventory(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	material recoverykeymaterial.Material,
	component databaseBackupComponent,
) (string, int64, error) {
	source, err := executor.storage.RecoveryReader().OpenExact(
		ctx, component.identity.Key, component.identity.VersionID,
	)
	if err != nil {
		return "", 0, err
	}
	defer source.Close()
	var plaintext bytes.Buffer
	_, err = backupcrypto.DecryptExpected(
		source, &plaintext,
		archiveContext(claim, "documents_inventory",
			material.DRKeyID(), material.RecoveryKeyID()),
		material.DROpenKeys(),
	)
	if err != nil {
		return "", 0, err
	}
	return parseDocumentsRecoveryInventory(
		plaintext.Bytes(), claim.EnvironmentID,
		claim.RecoverySetID, executor.target.TenantID(),
		executor.documents,
	)
}
