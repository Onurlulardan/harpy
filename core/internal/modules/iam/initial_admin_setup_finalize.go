package iam

import (
	"context"
	"fmt"
	"regexp"

	setuppostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/initialadminsetup"
	"github.com/google/uuid"
)

var initialAdminSetupSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type InitialAdminSetupFinalizeInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	SchemaName, ManifestSHA256           string
}

func FinalizeInstallationInitialAdminSetup(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InitialAdminSetupFinalizeInput,
) (uuid.UUID, error) {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		input.OperationID == [16]byte{} || input.TenantID == [16]byte{} ||
		input.EnvironmentID == [16]byte{} || len(input.SchemaName) < 3 ||
		!initialAdminSetupSHA256Pattern.MatchString(input.ManifestSHA256) {
		return uuid.Nil, ErrInstallationIdentityInvalid
	}
	installation, err := setuppostgres.NewInstallation(func(
		ctx context.Context, statement string, arguments ...any,
	) setuppostgres.Row {
		return executor.QueryRow(ctx, statement, arguments...)
	})
	if err != nil {
		return uuid.Nil, err
	}
	artifactID, err := installation.Apply(ctx, setuppostgres.FinalizeInput{
		OperationID:   uuid.UUID(input.OperationID),
		EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID:      uuid.UUID(input.TenantID), SchemaName: input.SchemaName,
		ManifestSHA256: input.ManifestSHA256,
	})
	if err != nil || artifactID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("finalize initial admin setup: %w", err)
	}
	return artifactID, nil
}
