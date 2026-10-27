package recoveryorchestration

import (
	"context"
	"fmt"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func restoreDatabaseArchive(
	ctx context.Context,
	archive DatabaseArchive,
	reader RecoveryObjectReader,
	keys databaseRestoreKeys,
	dsn string,
	identity recoveryset.ComponentIdentity,
	environmentID, recoverySetID string,
) error {
	object := restoreObjectIdentity(identity)
	if err := reader.DownloadVerified(ctx, object, io.Discard); err != nil {
		return fmt.Errorf("database archive readback failed: %w", err)
	}
	source, err := reader.OpenExact(ctx, object.Key, object.VersionID)
	if err != nil {
		return err
	}
	defer source.Close()
	if keys.material != nil {
		expected := backupcrypto.Context{
			EnvironmentID: environmentID, RecoverySetID: recoverySetID,
			ComponentID: "database_archive", DRKeyID: keys.material.DRKeyID(),
			RecoveryKeyID: keys.material.RecoveryKeyID(),
		}
		_, err = archive.Restore(
			ctx, dsn, source, expected, keys.material.DROpenKeys(),
		)
	} else {
		_, _, err = archive.RestoreBound(
			ctx, dsn, source, backupcrypto.BoundContext{
				EnvironmentID: environmentID, RecoverySetID: recoverySetID,
				ComponentID: "database_archive",
			}, keys.resolveRecoveryKey,
		)
	}
	if err != nil {
		return fmt.Errorf("database archive restore failed: %w", err)
	}
	return nil
}

func restoreObjectIdentity(
	identity recoveryset.ComponentIdentity,
) ports.ObjectIdentity {
	return ports.ObjectIdentity{
		Key: identity.ObjectKey, VersionID: identity.ObjectVersionID,
		SHA256: identity.CiphertextSHA256, Bytes: identity.CiphertextBytes,
	}
}
