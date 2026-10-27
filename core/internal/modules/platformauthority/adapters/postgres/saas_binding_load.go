package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) LoadSaaSBindingForEnvironment(
	ctx context.Context,
	environmentID string,
) (ports.BindingRecord, error) {
	var result ports.BindingRecord
	err := store.withBound(ctx, transactionSnapshotQuery, func(bound boundStore) error {
		var mode, kind string
		var licenseID, finalization, recovery *string
		err := bound.database.QueryRow(ctx, getSaaSBindingSQL, environmentID).Scan(
			&result.BindingID, &result.EnvironmentID, &mode, &kind,
			&licenseID, &finalization, &recovery, &result.SignerGeneration,
			&result.SignerKID, &result.SignerPublicKeySHA256, &result.SignerStatus,
		)
		result.DeploymentMode = artifact.DeploymentMode(mode)
		result.RegistryKind = artifact.RegistryKind(kind)
		result.LicenseBindingID = licenseID
		result.FinalizationReceiptSHA256 = finalization
		result.RecoveryTrustBundleSHA256 = recovery
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.BindingRecord{}, fmt.Errorf("get SaaS authority binding: %w", err)
	}
	return result, nil
}
