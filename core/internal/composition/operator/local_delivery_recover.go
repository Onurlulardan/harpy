package operator

import (
	"context"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func (runtime *deliveryRuntime) recoverUnknown(
	ctx context.Context,
	delivery tenancy.InstallationDelivery,
) error {
	cleanup := func() error {
		if err := cleanupToken(runtime.statePath(recoveryTokenName)); err != nil {
			return err
		}
		return cleanupToken(runtime.statePath(claimTokenName))
	}
	return recoverUnknownInstallationDelivery(ctx, runtime.delivery, delivery,
		runtime.passwordPath(), runtime.statePath(recoveryTokenName), cleanup)
}

func recoverUnknownInstallationDelivery(
	ctx context.Context,
	runtime *tenancy.InstallationDeliveryRuntime,
	delivery tenancy.InstallationDelivery,
	passwordPath string,
	recoveryPath string,
	cleanup func() error,
) error {
	if _, err := os.Lstat(recoveryPath); os.IsNotExist(err) {
		if err = removePasswordIfPresent(passwordPath); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	token, digest, err := loadOrCreateToken(recoveryPath)
	if err != nil {
		return err
	}
	defer clear(token)
	prepared, err := prepareRecoveryPassword(passwordPath)
	if err != nil {
		return err
	}
	request, err := buildInstallationRecoveryRequest(delivery, digest)
	if err != nil {
		return err
	}
	err = runtime.Recover(ctx, request,
		func(scoped context.Context, binding tenancy.InstallationRecoveryBinding,
			executor tenancy.InstallationRecoveryOwnerExecutor) error {
			return iam.SupersedeInstallationCredential(scoped,
				iam.InstallationActivationExecutor{QueryRow: func(
					ctx context.Context, statement string, arguments ...any,
				) iam.InstallationActivationRow {
					return executor.QueryRow(ctx, statement, arguments...)
				}}, iam.InstallationRecoveryBinding{
					OperationID: binding.OperationID, TenantID: binding.TenantID,
					FirstOwnerUserID:         binding.FirstOwnerUserID,
					CurrentCredentialID:      binding.CurrentCredentialID,
					NextCredentialID:         binding.NextCredentialID,
					CurrentCredentialVersion: binding.CurrentCredentialVersion,
					NextCredentialVersion:    binding.NextCredentialVersion,
					CurrentAttemptID:         binding.CurrentAttemptID,
					NextAttemptID:            binding.NextAttemptID,
					Cause:                    string(binding.Cause), ManifestSHA256: binding.ManifestSHA256,
				}, prepared)
		})
	if err == nil {
		return cleanup()
	}
	current, readErr := runtime.Read(ctx, request.OperationID)
	if readErr == nil && current.State == tenancy.InstallationDeliveryStateReady &&
		current.CurrentAttemptID == request.NextAttemptID &&
		current.CurrentCredentialID == request.NextCredentialID &&
		current.CurrentCredentialVersion == request.NextCredentialVersion {
		return cleanup()
	}
	return err
}

func buildInstallationRecoveryRequest(
	delivery tenancy.InstallationDelivery,
	digest string,
) (tenancy.InstallationRecoveryRequest, error) {
	nextVersion := delivery.CurrentCredentialVersion + 1
	nextCredentialID := localinstall.NextCredentialID(delivery.OperationID, nextVersion)
	nextAttemptID, err := tenancy.InstallationDeliveryInitialAttemptID(
		delivery.OperationID, nextCredentialID,
	)
	if err != nil {
		return tenancy.InstallationRecoveryRequest{}, err
	}
	schemaName, err := tenancy.OnboardingSchemaName(tenancy.UUID(delivery.TenantID))
	if err != nil {
		return tenancy.InstallationRecoveryRequest{}, err
	}
	return tenancy.InstallationRecoveryRequest{
		OperationID: delivery.OperationID, CurrentAttemptID: delivery.CurrentAttemptID,
		CurrentCredentialID:      delivery.CurrentCredentialID,
		CurrentCredentialVersion: delivery.CurrentCredentialVersion,
		NextAttemptID:            nextAttemptID, NextCredentialID: nextCredentialID,
		NextCredentialVersion: nextVersion, ExpectedRevision: delivery.Revision,
		RecoveryTokenSHA256: digest, SchemaName: schemaName, LeaseSeconds: 300,
		DeliveryManifestSHA256: tenancy.InstallationDeliveryManifestSHA256(),
	}, nil
}
