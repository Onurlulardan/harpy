package iam

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/google/uuid"
)

var ErrInstallationPlatformOwnerUnverified = errors.New(
	"IAM installation platform owner is not verified",
)

type InstallationPlatformOwnerInput struct {
	OperationID, EnvironmentID, TenantID, FirstUserID [16]byte
	InstanceOwnerBindingID                            [16]byte
	SchemaName, RequestSHA256                         string
}

func InitializeInstallationPlatformOwner(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationPlatformOwnerInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validInstallationPlatformOwner(input) {
		return ErrInstallationIdentityInvalid
	}
	binding, err := bindInstallationExecutor(executor)
	if err != nil {
		return err
	}
	return initializeInstallationPlatformOwner(ctx, binding, input)
}

func ReconcileInstallationPlatformOwner(
	ctx context.Context,
	executor InstallationActivationExecutor,
	environmentID [16]byte,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		zeroInstallationID(environmentID) {
		return ErrInstallationIdentityInvalid
	}
	binding, err := bindInstallationExecutor(executor)
	if err != nil {
		return err
	}
	stored, err := binding.OwnerInput(ctx, uuid.UUID(environmentID))
	if err != nil {
		return fmt.Errorf("read IAM installation platform owner input: %w", err)
	}
	record := installationPlatformOwnerRecord(InstallationPlatformOwnerInput{
		OperationID: uuidBytes(stored.OperationID), EnvironmentID: environmentID,
		TenantID: uuidBytes(stored.TenantID), FirstUserID: uuidBytes(stored.UserID),
		InstanceOwnerBindingID: uuidBytes(stored.OwnerBindingID),
		SchemaName:             stored.SchemaName, RequestSHA256: stored.RequestSHA256,
	})
	verified, err := binding.VerifyOwner(ctx, record)
	if err != nil {
		return fmt.Errorf("verify IAM installation platform owner: %w", err)
	}
	if !verified {
		return fmt.Errorf("verify IAM installation platform owner: %w",
			ErrInstallationPlatformOwnerUnverified)
	}
	return nil
}

func bindInstallationExecutor(
	executor InstallationActivationExecutor,
) (iampostgres.InstallationBinding, error) {
	return iampostgres.BindInstallation(iampostgres.InstallationDatabase{
		QueryRow: func(
			ctx context.Context, statement string, arguments ...any,
		) iampostgres.InstallationRow {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
}

func initializeInstallationPlatformOwner(
	ctx context.Context,
	binding iampostgres.InstallationBinding,
	input InstallationPlatformOwnerInput,
) error {
	replayed, err := binding.Owner(ctx, installationPlatformOwnerRecord(input))
	if err != nil {
		return fmt.Errorf("initialize IAM installation platform owner: %w", err)
	}
	_ = replayed
	return nil
}

func installationPlatformOwnerRecord(
	input InstallationPlatformOwnerInput,
) iampostgres.InstallationPlatformOwnerRecord {
	return iampostgres.InstallationPlatformOwnerRecord{
		OperationID: uuid.UUID(input.OperationID), EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: uuid.UUID(input.TenantID), UserID: uuid.UUID(input.FirstUserID),
		SubjectID:   uuid.UUID(deriveInstallationID(input.OperationID, "platform-owner-subject")),
		OwnerRoleID: uuid.UUID(deriveInstallationID(input.OperationID, "platform-owner-role")),
		RolePermissionID: uuid.UUID(deriveInstallationID(
			input.OperationID, "platform-owner-environment-all",
		)),
		OwnerBindingID: uuid.UUID(input.InstanceOwnerBindingID),
		SchemaName:     input.SchemaName, RequestSHA256: input.RequestSHA256,
		AuthorizationVersion: 1,
	}
}

func validInstallationPlatformOwner(input InstallationPlatformOwnerInput) bool {
	for _, value := range [][16]byte{input.OperationID, input.EnvironmentID,
		input.TenantID, input.FirstUserID, input.InstanceOwnerBindingID} {
		if zeroInstallationID(value) {
			return false
		}
	}
	decoded, err := hex.DecodeString(input.RequestSHA256)
	return err == nil && len(decoded) == 32 && len(input.SchemaName) >= 3 &&
		len(input.SchemaName) <= 63
}

func uuidBytes(value uuid.UUID) [16]byte { return [16]byte(value) }
