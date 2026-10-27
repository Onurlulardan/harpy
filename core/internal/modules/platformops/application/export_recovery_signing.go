package application

import (
	"context"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (coordinator *RecoveryCoordinator) ExportRecoverySigning(
	ctx context.Context,
	environmentID string,
	generation int64,
) (api.RecoverySigningBundle, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil ||
		environmentID == "" || generation < 1 {
		return api.RecoverySigningBundle{}, api.ErrInvalid
	}
	record, err := coordinator.store.LoadSigningBundle(ctx, environmentID, generation)
	if err != nil {
		return api.RecoverySigningBundle{}, mapRecoveryStoreError(err)
	}
	bundle, err := recoverysigningv1.Parse(record.BundleBytes)
	if err != nil || bundle.EnvironmentID() != record.EnvironmentID ||
		bundle.Generation() != record.Generation || bundle.SHA256() != record.SHA256 {
		return api.RecoverySigningBundle{}, api.ErrConflict
	}
	return signingBundleAPI(record), nil
}

func signingBundleAPI(record ports.SigningBundleRecord) api.RecoverySigningBundle {
	return api.RecoverySigningBundle{EnvironmentID: record.EnvironmentID,
		Generation: record.Generation, SHA256: record.SHA256,
		BundleBytes: append([]byte(nil), record.BundleBytes...)}
}

var _ api.RecoverySigning = (*RecoveryCoordinator)(nil)
