package postgres

import (
	"context"
	_ "embed"
)

//go:embed sql/platform/installation_platform_owner_verify.sql
var installationPlatformOwnerVerifySQL string

func bindInstallationPlatformOwnerVerifier(
	database InstallationDatabase,
) func(context.Context, InstallationPlatformOwnerRecord) (bool, error) {
	return func(
		ctx context.Context,
		input InstallationPlatformOwnerRecord,
	) (bool, error) {
		var verified bool
		err := database.QueryRow(ctx, installationPlatformOwnerVerifySQL,
			input.OperationID, input.EnvironmentID, input.TenantID, input.UserID,
			input.SubjectID, input.OwnerRoleID, input.RolePermissionID,
			input.OwnerBindingID, input.AuthorizationVersion,
			input.RequestSHA256,
		).Scan(&verified)
		return verified, err
	}
}
