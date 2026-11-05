package operator

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func prepareInitialCredential(
	path string,
	delivery tenancy.InstallationDelivery,
	readErr error,
) error {
	if errors.Is(readErr, installationdelivery.ErrNotFound) {
		return ensureInitialPassword(path)
	}
	if readErr != nil {
		return readErr
	}
	return validateDeliveryCredentialFile(path, delivery)
}

func prepareDeliveryCredential(
	path string,
	delivery tenancy.InstallationDelivery,
	readErr error,
) error {
	if readErr != nil {
		return readErr
	}
	return validateDeliveryCredentialFile(path, delivery)
}

func validateDeliveryCredentialFile(
	path string,
	delivery tenancy.InstallationDelivery,
) error {
	switch delivery.State {
	case tenancy.InstallationDeliveryStateReady:
		return requireInitialPassword(path)
	case tenancy.InstallationDeliveryStateClaimed,
		tenancy.InstallationDeliveryStateOutcomeUnknown,
		tenancy.InstallationDeliveryStateAcknowledged:
		return nil
	case tenancy.InstallationDeliveryStateSecretDeleted,
		tenancy.InstallationDeliveryStateCompleted:
		return requireInitialPasswordAbsent(path)
	default:
		return tenancy.ErrInstallationDeliveryUnavailable
	}
}
