package recoveryorchestration

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) buildComponents(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	material recoverykeymaterial.Material,
) (databaseBackupBuild, error) {
	snapshot, err := executor.beginSnapshot(ctx)
	if err != nil {
		return databaseBackupBuild{}, err
	}
	snapshotClosed := false
	defer func() {
		if !snapshotClosed {
			_ = snapshot.Close()
		}
	}()
	writer := executor.storage.Writer(claim.DatabaseNow)
	if writer == nil {
		return databaseBackupBuild{}, fmt.Errorf("recovery storage writer is unavailable")
	}
	prefix, err := executor.storage.Prefix(
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return databaseBackupBuild{}, err
	}
	keys := material.SealKeys(rand.Reader)
	database, err := putDatabaseArchive(
		ctx, writer, componentKey(prefix, "database_archive"), executor.dsn,
		snapshot.Identifier(), executor.archive,
		archiveContext(claim, "database_archive",
			material.DRKeyID(), material.RecoveryKeyID()), keys,
	)
	if err != nil {
		return databaseBackupBuild{}, err
	}
	database.kind = "database_archive"
	documents, documentsDigest, documentsCount, err :=
		executor.buildDocumentsComponent(
			ctx, claim, snapshot.Identifier(), prefix, writer, keys,
			material.DRKeyID(), material.RecoveryKeyID(),
		)
	if err != nil {
		return databaseBackupBuild{}, err
	}
	if err = snapshot.Close(); err != nil {
		return databaseBackupBuild{}, err
	}
	snapshotClosed = true
	components := map[string]databaseBackupComponent{
		database.kind: database, documents.kind: documents,
	}
	other, err := executor.buildEvidenceComponents(
		ctx, claim, prefix, writer, material, keys,
	)
	if err != nil {
		return databaseBackupBuild{}, err
	}
	for kind, component := range other {
		components[kind] = component
	}
	return databaseBackupBuild{
		components: components, documentsDigest: documentsDigest,
		documentsCount: documentsCount,
	}, nil
}

func componentKey(prefix, kind string) string {
	return prefix + "components/" + kind + ".fcbackup1"
}
