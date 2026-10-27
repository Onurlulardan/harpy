package recoveryorchestration

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseRestoreExecutor) prepareScratch(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	identity ports.RestoreExecutionIdentity,
	manifest recoveryset.ManifestPayload,
	applyArchive bool,
) (databaseRestoreEvidence, error) {
	if applyArchive {
		if err := restoreDatabaseArchive(
			ctx, executor.archive, executor.storage.RecoveryReader(),
			executor.keys, executor.scratchDSN, manifest.Components.DatabaseArchive,
			claim.EnvironmentID, claim.RecoverySetID,
		); err != nil {
			return databaseRestoreEvidence{}, resumableRestore(err)
		}
	}
	digest, digestErr := decodeTenantSetDigest(manifest.Identity.TenantSetSHA256)
	if digestErr != nil {
		return databaseRestoreEvidence{}, terminalRestore(digestErr)
	}
	scratch, err := executor.openScratch(
		ctx, manifest.Identity.EnvironmentID,
		manifest.Identity.PrimaryBindingID, digest, manifest.DatabaseSystemID,
	)
	if err != nil {
		return databaseRestoreEvidence{}, dependencyRestore(err)
	}
	evidence := databaseRestoreEvidence{scratch: scratch}
	documents := scratch.Documents()
	if documents == nil {
		evidence.close()
		return databaseRestoreEvidence{},
			terminalRestore(fmt.Errorf("restored documents binding is unavailable"))
	}
	evidence.documents = documents
	evidence.archiveSHA256 = manifest.DatabaseArchiveSHA256
	evidence.newDatabaseID = claim.NewDatabaseSystemID
	if evidence.newDatabaseID == "" {
		evidence.newDatabaseID, err = executor.newDatabaseID()
		if err != nil {
			evidence.close()
			return databaseRestoreEvidence{}, resumableRestore(err)
		}
	}
	evidence.objectMerkle, evidence.locatorRemap, evidence.safety, err =
		executor.verifyAndRehydrate(ctx, claim, identity, manifest, documents)
	if err != nil {
		evidence.close()
		return databaseRestoreEvidence{}, err
	}
	evidence.componentInventory = append(
		[]byte(nil), claim.ComponentInventory...,
	)
	return evidence, nil
}

func decodeTenantSetDigest(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("restored tenant identity differs")
	}
	copy(result[:], decoded)
	return result, nil
}
