package async

import (
	"context"
	"fmt"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	installationpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres/installationactivation"
	"github.com/google/uuid"
)

type InitialAdminSetupDeliveryInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	ArtifactID                           [16]byte
	SchemaName, ManifestSHA256           string
}

func EnqueueInitialAdminSetupDelivery(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InitialAdminSetupDeliveryInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		input.OperationID == [16]byte{} || input.TenantID == [16]byte{} ||
		input.EnvironmentID == [16]byte{} || input.ArtifactID == [16]byte{} ||
		len(input.SchemaName) < 3 ||
		!installationEventDigestPattern.MatchString(input.ManifestSHA256) {
		return ErrInstallationActivationEventInvalid
	}
	reference, err := lifecycledeliveryv1.NewReference(
		uuid.UUID(input.TenantID), uuid.UUID(input.ArtifactID),
	)
	if err != nil {
		return ErrInstallationActivationEventInvalid
	}
	binding, err := installationpostgres.Bind(installationpostgres.Database{
		QueryRow: func(ctx context.Context, statement string,
			arguments ...any) installationpostgres.Row {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind initial admin setup delivery: %w", err)
	}
	_, err = binding.EnqueueInitialAdminSetup(ctx,
		input.OperationID, input.EnvironmentID, input.TenantID, input.ArtifactID,
		reference.String(), input.SchemaName, input.ManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("enqueue initial admin setup delivery: %w", err)
	}
	return nil
}
