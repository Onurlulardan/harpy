package application

import (
	"context"
	"errors"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (coordinator *RecoveryCoordinator) RecordRecoverySigning(
	ctx context.Context,
	request api.RecoverySigningRequest,
) (api.RecoverySigningBundle, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil || request.EnvironmentID == "" {
		return api.RecoverySigningBundle{}, api.ErrInvalid
	}
	bundle, err := recoverysigningv1.Parse(request.BundleBytes)
	if err != nil || bundle.EnvironmentID() != request.EnvironmentID ||
		bundle.GeneratedAt().After(coordinator.clock.Now()) {
		return api.RecoverySigningBundle{}, api.ErrInvalid
	}
	record := ports.SigningBundleRecord{EnvironmentID: bundle.EnvironmentID(),
		Generation: bundle.Generation(), SHA256: bundle.SHA256(),
		BundleBytes: bundle.CanonicalBytes(), CreatedAt: bundle.GeneratedAt()}
	if err = coordinator.store.RecordSigningBundle(ctx, record); err != nil {
		return api.RecoverySigningBundle{}, mapRecoveryStoreError(err)
	}
	return signingBundleAPI(record), nil
}

func mapRecoveryStoreError(err error) error {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		return api.ErrNotFound
	case errors.Is(err, ports.ErrConflict):
		return api.ErrConflict
	default:
		return api.ErrUnavailable
	}
}
