package platformops

import (
	"context"
	"errors"
	"fmt"

	installationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/installationactivation"
)

var ErrInstallationRecoveryIdentityInvalid = errors.New(
	"platform operations installation recovery identity is invalid",
)

func InitializeInstallationRecoveryIdentity(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationRecoveryIdentityInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validInstallationRecoveryIdentity(input) {
		return ErrInstallationRecoveryIdentityInvalid
	}
	binding, err := installationpostgres.Bind(installationpostgres.Database{
		QueryRow: func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) installationpostgres.Row {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind platform operations installation owner: %w", err)
	}
	var generation any
	var trustBundleSHA256 any
	if input.RecoverySigningPresent {
		generation = input.SigningGeneration
		trustBundleSHA256 = input.TrustBundleSHA256
	}
	_, err = binding.Initialize(ctx,
		input.OperationID, input.EnvironmentID, input.TenantID,
		input.SchemaName, input.DeploymentMode, generation,
		trustBundleSHA256, input.ManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("initialize platform operations recovery identity: %w", err)
	}
	return nil
}
