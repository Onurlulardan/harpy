package tenancy

import (
	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	installationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

type InstallationDelivery = installationdeliveryapp.Delivery
type InstallationDeliveryState = installationdeliveryapp.State
type InstallationDeliveryClaimCommand = installationdeliveryapp.ClaimCommand
type InstallationDeliveryAcknowledgeCommand = installationdeliveryapp.AcknowledgeCommand
type InstallationDeliveryUnknownCommand = installationdeliveryapp.UnknownCommand
type InstallationDeliveryStateCommand = installationdeliveryapp.StateCommand
type InstallationRecoveryRequest = onboardingapp.RecoveryRequest
type InstallationRecoveryOwnerExecutor = onboardingapp.RecoveryOwnerExecutor
type InstallationRecoveryBinding = onboardingapp.RecoveryBinding

const (
	InstallationDeliveryStateReady          = installationdeliveryapp.StateReady
	InstallationDeliveryStateClaimed        = installationdeliveryapp.StateClaimed
	InstallationDeliveryStateOutcomeUnknown = installationdeliveryapp.StateOutcomeUnknown
	InstallationDeliveryStateAcknowledged   = installationdeliveryapp.StateAcknowledged
	InstallationDeliveryStateSecretDeleted  = installationdeliveryapp.StateSecretDeleted
	InstallationDeliveryStateCompleted      = installationdeliveryapp.StateCompleted
	InstallationDeliveryCauseOutcomeUnknown = credentialdeliveryv1.CauseOutcomeUnknown
	InstallationDeliveryCauseLeaseExpired   = credentialdeliveryv1.CauseClaimLeaseExpired
)

var (
	ErrInstallationDeliveryInvalid     = installationdeliveryapp.ErrInvalid
	ErrInstallationDeliveryUnavailable = installationdeliveryapp.ErrUnavailable
)

func InstallationDeliveryInitialAttemptID(
	operationID uuid.UUID,
	credentialID uuid.UUID,
) (uuid.UUID, error) {
	return installationdeliveryapp.InitialAttemptID(operationID, credentialID)
}

func InstallationDeliveryManifestSHA256() string {
	return credentialdeliveryv1.ManifestSHA256()
}
