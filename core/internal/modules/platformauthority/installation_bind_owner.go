package platformauthority

import (
	"context"
	"errors"
	"fmt"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
)

var ErrInstallationAuthorityInvalid = errors.New(
	"platform authority installation input is invalid",
)

func BindInstallationInstanceOwner(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationOwnerBindingInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validOwnerBinding(input) {
		return ErrInstallationAuthorityInvalid
	}
	binding, err := installationBinding(executor)
	if err != nil {
		return err
	}
	_, err = binding.BindOwner(ctx,
		input.OperationID, input.EnvironmentID, input.TenantID,
		input.SchemaName, input.DeploymentMode,
		input.InstanceOwnerBindingID, input.ProtectedAdminBindingID,
		input.RegistryBindingID, input.ManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("bind installation instance owner: %w", err)
	}
	return nil
}

func installationBinding(
	executor InstallationActivationExecutor,
) (postgresadapter.InstallationBinding, error) {
	return postgresadapter.BindInstallation(postgresadapter.InstallationDatabase{
		QueryRow: func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) postgresadapter.InstallationRow {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
}
