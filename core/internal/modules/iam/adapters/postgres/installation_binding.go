package postgres

import (
	"context"
	_ "embed"
	"errors"

	"github.com/google/uuid"
)

//go:embed sql/tenant/installation_initialize.sql
var installationInitializeSQL string

//go:embed sql/platform/installation_directory.sql
var installationDirectorySQL string

//go:embed sql/platform/installation_activation_assert.sql
var installationActivationAssertSQL string

//go:embed sql/platform/installation_security_epoch.sql
var installationSecurityEpochSQL string

//go:embed sql/platform/installation_platform_owner.sql
var installationPlatformOwnerSQL string

//go:embed sql/platform/installation_platform_owner_input.sql
var installationPlatformOwnerInputSQL string

type InstallationDatabase struct {
	QueryRow func(context.Context, string, ...any) InstallationRow
}

type InstallationRow interface{ Scan(...any) error }

type InstallationIdentityRecord struct {
	OperationID, EnvironmentID, TenantID             uuid.UUID
	UserID, OrganizationID, ProtectedAssignmentID    uuid.UUID
	Email, DisplayName, TenantName, BindingSetSHA256 string
	UserSchemaID, UserFieldID                        uuid.UUID
	OrganizationSchemaID, OrganizationFieldID        uuid.UUID
	RoleSchemaID, RoleFieldID                        uuid.UUID
	IdentifierID, AccessModeID, MembershipID         uuid.UUID
	RoleID, PermissionID, RolePermissionID           uuid.UUID
}

type InstallationDirectoryRecord struct {
	OperationID, EnvironmentID, TenantID uuid.UUID
	UserID, IdentifierID                 uuid.UUID
	SchemaName, KeyID                    string
	IdentifierHMAC, SourceSHA256         [32]byte
}

type InstallationPlatformOwnerRecord struct {
	OperationID, EnvironmentID, TenantID, UserID uuid.UUID
	SubjectID, OwnerRoleID, RolePermissionID     uuid.UUID
	OwnerBindingID                               uuid.UUID
	SchemaName, RequestSHA256                    string
	AuthorizationVersion                         int64
}

type InstallationPlatformOwnerInputRecord struct {
	OperationID, TenantID, UserID, OwnerBindingID uuid.UUID
	SchemaName, RequestSHA256                     string
}

type InstallationBinding struct {
	Assert      func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (bool, error)
	Epoch       func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (bool, error)
	Identity    func(context.Context, InstallationIdentityRecord) (bool, error)
	Directory   func(context.Context, InstallationDirectoryRecord) (int64, error)
	Owner       func(context.Context, InstallationPlatformOwnerRecord) (bool, error)
	VerifyOwner func(context.Context, InstallationPlatformOwnerRecord) (bool, error)
	OwnerInput  func(
		context.Context, uuid.UUID,
	) (InstallationPlatformOwnerInputRecord, error)
}

func BindInstallation(database InstallationDatabase) (InstallationBinding, error) {
	if database.QueryRow == nil {
		return InstallationBinding{}, errors.New("IAM installation database is invalid")
	}
	return InstallationBinding{
		Assert: func(
			ctx context.Context,
			operationID, environmentID, tenantID uuid.UUID,
			schemaName string,
		) (bool, error) {
			var valid bool
			err := database.QueryRow(ctx, installationActivationAssertSQL,
				operationID, environmentID, tenantID, schemaName).Scan(&valid)
			return valid, err
		},
		Epoch: func(
			ctx context.Context,
			operationID, environmentID, tenantID uuid.UUID,
			schemaName string,
		) (bool, error) {
			var valid bool
			err := database.QueryRow(ctx, installationSecurityEpochSQL,
				operationID, environmentID, tenantID, schemaName).Scan(&valid)
			return valid, err
		},
		Identity: func(ctx context.Context, input InstallationIdentityRecord) (bool, error) {
			var replay bool
			err := database.QueryRow(ctx, installationInitializeSQL,
				input.OperationID, input.EnvironmentID, input.TenantID, input.UserID,
				input.OrganizationID, input.ProtectedAssignmentID, input.Email,
				input.DisplayName, input.TenantName, input.BindingSetSHA256,
				input.UserSchemaID, input.UserFieldID, input.OrganizationSchemaID,
				input.OrganizationFieldID, input.RoleSchemaID, input.RoleFieldID,
				input.IdentifierID, input.AccessModeID, input.MembershipID, input.RoleID,
				input.PermissionID, input.RolePermissionID).Scan(&replay)
			return replay, err
		},
		Directory: func(ctx context.Context, input InstallationDirectoryRecord) (int64, error) {
			var revision int64
			err := database.QueryRow(ctx, installationDirectorySQL,
				input.OperationID, input.EnvironmentID, input.TenantID,
				input.SchemaName, input.UserID, input.IdentifierID,
				input.KeyID, input.IdentifierHMAC[:], input.SourceSHA256[:],
			).Scan(&revision)
			return revision, err
		},
		Owner: func(ctx context.Context, input InstallationPlatformOwnerRecord) (bool, error) {
			var replay bool
			err := database.QueryRow(ctx, installationPlatformOwnerSQL,
				input.OperationID, input.EnvironmentID, input.TenantID, input.SchemaName,
				input.UserID, input.SubjectID, input.OwnerRoleID, input.RolePermissionID,
				input.OwnerBindingID, input.AuthorizationVersion,
				input.RequestSHA256).Scan(&replay)
			return replay, err
		},
		VerifyOwner: bindInstallationPlatformOwnerVerifier(database),
		OwnerInput: func(
			ctx context.Context, environmentID uuid.UUID,
		) (InstallationPlatformOwnerInputRecord, error) {
			var result InstallationPlatformOwnerInputRecord
			err := database.QueryRow(ctx, installationPlatformOwnerInputSQL,
				environmentID).Scan(&result.OperationID, &result.TenantID,
				&result.SchemaName, &result.UserID, &result.OwnerBindingID,
				&result.RequestSHA256)
			return result, err
		},
	}, nil
}
