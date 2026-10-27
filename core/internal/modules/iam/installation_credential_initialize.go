package iam

import (
	"context"
	"fmt"

	installationcredentialpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/installationcredential"
	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/google/uuid"
)

func InitializeInstallationCredential(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationCredentialInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		input.OperationID == [16]byte{} || input.TenantID == [16]byte{} ||
		input.EnvironmentID == [16]byte{} || input.UserID == [16]byte{} ||
		!input.Credential.Valid() {
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
		return fmt.Errorf("bind IAM installation credential owner: %w", err)
	}
	operationID, tenantID := uuid.UUID(input.OperationID), uuid.UUID(input.TenantID)
	userID := uuid.UUID(input.UserID)
	credentialID, err := installationcredential.InitialCredentialID(operationID, userID)
	if err != nil {
		return ErrInstallationCredentialInvalid
	}
	_, err = installationcredential.Create(ctx, installationcredential.ActivationBinding{
		OperationID: operationID, EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: tenantID, FirstUserID: userID, Kind: "environment_initialize",
	}, store, installationcredential.CreateCommand{
		OperationID: operationID, TenantID: tenantID, UserID: userID,
		CredentialID: credentialID, CredentialVersion: 1,
		Credential: input.Credential.value,
	})
	if err != nil {
		return fmt.Errorf("initialize IAM installation credential: %w", err)
	}
	return nil
}
