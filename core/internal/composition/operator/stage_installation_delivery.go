package operator

import (
	"context"
	"errors"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func stageInstallationDelivery(
	ctx context.Context,
	runtime *tenancy.InstallationDeliveryRuntime,
	manifest installationv1.Manifest,
	result environmentoperator.InitializeResult,
	passwordPath config.ArtifactPath,
) error {
	if runtime == nil || result.OperationID == [16]byte{} || result.FirstUserID == [16]byte{} {
		return installationdelivery.ErrInvalid
	}
	existing, err := runtime.Read(ctx, result.OperationID)
	if err == nil {
		return validateExistingDelivery(existing, result)
	}
	if !errors.Is(err, installationdelivery.ErrUnavailable) {
		return err
	}
	if err = installationsecret.Read(passwordPath.Path().Value(),
		func([]byte) error { return nil }); err != nil {
		return err
	}
	credentialID, err := installationcredential.InitialCredentialID(
		result.OperationID, result.FirstUserID)
	if err != nil {
		return err
	}
	attemptID, err := installationdelivery.InitialAttemptID(result.OperationID, credentialID)
	if err != nil {
		return err
	}
	_, err = runtime.Stage(ctx, installationdelivery.StageCommand{
		OperationID: result.OperationID, AttemptID: attemptID,
		CredentialID: credentialID, CredentialVersion: 1,
		InstallationManifestSHA256: manifest.SHA256(),
		DeliveryManifestSHA256:     credentialdeliveryv1.ManifestSHA256(),
	})
	return err
}

func validateExistingDelivery(
	delivery installationdelivery.Delivery,
	result environmentoperator.InitializeResult,
) error {
	if delivery.OperationID != result.OperationID || delivery.TenantID != result.TenantID ||
		delivery.FirstOwnerUserID != result.FirstUserID {
		return installationdelivery.ErrConflict
	}
	return nil
}
