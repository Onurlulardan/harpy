package iam

import (
	"context"
	"fmt"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	installationcredentialpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/installationcredential"
	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/google/uuid"
)

func SupersedeInstallationCredential(
	ctx context.Context,
	executor InstallationActivationExecutor,
	binding InstallationRecoveryBinding,
	credential PreparedInstallationCredential,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil || !credential.Valid() {
		return ErrInstallationCredentialInvalid
	}
	store, err := installationcredentialpostgres.NewStore(
		installationcredentialpostgres.Executor{QueryRow: func(
			ctx context.Context, statement string, arguments ...any,
		) installationcredentialpostgres.Row {
			return executor.QueryRow(ctx, statement, arguments...)
		}},
	)
	if err != nil {
		return fmt.Errorf("bind IAM installation credential recovery: %w", err)
	}
	if _, err = installationcredential.Rotate(
		ctx, installationcredential.RotationBinding{
			OperationID:              uuid.UUID(binding.OperationID),
			TenantID:                 uuid.UUID(binding.TenantID),
			FirstOwnerUserID:         uuid.UUID(binding.FirstOwnerUserID),
			CurrentCredentialID:      uuid.UUID(binding.CurrentCredentialID),
			NextCredentialID:         uuid.UUID(binding.NextCredentialID),
			CurrentAttemptID:         uuid.UUID(binding.CurrentAttemptID),
			NextAttemptID:            uuid.UUID(binding.NextAttemptID),
			CurrentCredentialVersion: binding.CurrentCredentialVersion,
			NextCredentialVersion:    binding.NextCredentialVersion,
			Cause:                    credentialdeliveryv1.Cause(binding.Cause),
			ManifestSHA256:           binding.ManifestSHA256,
		}, store, credential.value,
	); err != nil {
		return fmt.Errorf("supersede IAM installation credential: %w", err)
	}
	return nil
}
