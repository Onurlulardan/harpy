package iam

import (
	"context"
	"errors"
	"fmt"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/google/uuid"
)

var ErrInstallationIdentityInvalid = errors.New("IAM installation identity is invalid")

func InitializeInstallationIdentity(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationIdentityInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validInstallationIdentity(input) {
		return ErrInstallationIdentityInvalid
	}
	binding, err := iampostgres.BindInstallation(iampostgres.InstallationDatabase{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) iampostgres.InstallationRow {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind IAM installation owner: %w", err)
	}
	valid, err := binding.Assert(ctx, uuid.UUID(input.OperationID),
		uuid.UUID(input.EnvironmentID), uuid.UUID(input.TenantID),
		input.SchemaName)
	if err != nil || !valid {
		return fmt.Errorf("assert IAM installation activation: %w", err)
	}
	epochReady, err := binding.Epoch(ctx, uuid.UUID(input.OperationID),
		uuid.UUID(input.EnvironmentID), uuid.UUID(input.TenantID),
		input.SchemaName)
	if err != nil || !epochReady {
		return fmt.Errorf("initialize IAM installation security epoch: %w", err)
	}
	ids := installationIDs(input.OperationID)
	bindingSet := ProtectedAdminBindingSetSHA256(input.OperationID,
		input.TenantID, input.FirstUserID, input.ProtectedAdminBindingID)
	_, err = binding.Identity(ctx, iampostgres.InstallationIdentityRecord{
		OperationID: uuid.UUID(input.OperationID), EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: uuid.UUID(input.TenantID), UserID: uuid.UUID(input.FirstUserID),
		OrganizationID:        uuid.UUID(input.RootOrganizationID),
		ProtectedAssignmentID: uuid.UUID(input.ProtectedAdminBindingID),
		Email:                 input.Email, DisplayName: input.DisplayName, TenantName: input.TenantName,
		BindingSetSHA256: bindingSet, UserSchemaID: uuid.UUID(ids.userSchema),
		UserFieldID: uuid.UUID(ids.userField), OrganizationSchemaID: uuid.UUID(ids.organizationSchema),
		OrganizationFieldID: uuid.UUID(ids.organizationField), RoleSchemaID: uuid.UUID(ids.roleSchema),
		RoleFieldID: uuid.UUID(ids.roleField), IdentifierID: uuid.UUID(ids.identifier),
		AccessModeID: uuid.UUID(ids.accessMode), MembershipID: uuid.UUID(ids.membership),
		RoleID: uuid.UUID(ids.role), PermissionID: uuid.UUID(ids.permission),
		RolePermissionID: uuid.UUID(ids.rolePermission),
	})
	if err != nil {
		return fmt.Errorf("initialize IAM installation identity: %w", err)
	}
	revision, err := binding.Directory(ctx, iampostgres.InstallationDirectoryRecord{
		OperationID: uuid.UUID(input.OperationID), EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: uuid.UUID(input.TenantID), SchemaName: input.SchemaName,
		UserID:       uuid.UUID(input.FirstUserID),
		IdentifierID: uuid.UUID(ids.identifier), KeyID: input.DirectoryKeyID,
		IdentifierHMAC: input.IdentifierHMAC, SourceSHA256: input.SourceSHA256,
	})
	if err != nil || revision != 1 {
		return fmt.Errorf("initialize IAM installation directory: revision=%d: %w", revision, err)
	}
	return nil
}
