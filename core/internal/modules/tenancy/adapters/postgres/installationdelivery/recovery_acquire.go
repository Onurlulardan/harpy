package installationdelivery

import (
	"context"
	_ "embed"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

//go:embed sql/platform/recover.sql
var recoverSQL string

func (storage *store) Acquire(
	ctx context.Context,
	request onboarding.RecoveryRequest,
) (onboarding.RecoveryClaim, error) {
	if !storage.usable(ctx) {
		return onboarding.RecoveryClaim{}, errRecoveryUnavailable
	}
	var claim onboarding.RecoveryClaim
	var binding onboarding.RecoveryBinding
	var cause string
	err := storage.queryRow(ctx, recoverSQL, request.OperationID,
		request.CurrentAttemptID, request.CurrentCredentialID,
		request.CurrentCredentialVersion, request.NextAttemptID,
		request.NextCredentialID, request.NextCredentialVersion,
		request.ExpectedRevision, request.RecoveryTokenSHA256,
		request.DeliveryManifestSHA256, request.LeaseSeconds).Scan(
		&binding.OperationID, &binding.TenantID, &binding.FirstOwnerUserID,
		&binding.CurrentCredentialID, &binding.CurrentCredentialVersion,
		&binding.NextCredentialID, &binding.NextCredentialVersion,
		&binding.CurrentAttemptID, &binding.NextAttemptID, &cause,
		&binding.ManifestSHA256, &claim.Revision, &claim.DatabaseNow,
		&claim.LeaseUntil)
	if err != nil {
		return onboarding.RecoveryClaim{}, errRecoveryUnavailable
	}
	binding.Cause = credentialdeliveryv1.Cause(cause)
	claim.Binding = binding
	return claim, nil
}
