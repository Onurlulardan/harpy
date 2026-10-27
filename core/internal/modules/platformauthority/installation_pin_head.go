package platformauthority

import (
	"context"
	"fmt"
)

func PinInstallationAuthorityHead(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationHeadPinInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validHeadPin(input) {
		return ErrInstallationAuthorityInvalid
	}
	binding, err := installationBinding(executor)
	if err != nil {
		return err
	}
	var previousHeadSHA256 any
	if input.PreviousHeadSHA256 != "" {
		previousHeadSHA256 = input.PreviousHeadSHA256
	}
	_, err = binding.PinHead(ctx,
		input.OperationID, input.EnvironmentID, input.TenantID,
		input.SchemaName, input.RegistryBindingID, input.ScopeKind,
		input.ScopeID, input.Generation, previousHeadSHA256,
		input.StateSHA256, input.ProposalSHA256, input.AckSHA256,
		input.ManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("pin installation authority head: %w", err)
	}
	return nil
}
