package installationdelivery

import (
	"context"
	_ "embed"

	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

//go:embed sql/platform/record_recovery.sql
var recordRecoverySQL string

func (storage *store) recordRecovery(
	ctx context.Context,
	record onboarding.RecoveryRecord,
) (onboarding.RecoveryTransition, error) {
	if !storage.usable(ctx) {
		return onboarding.RecoveryTransition{}, errRecoveryUnavailable
	}
	binding := record.Binding
	var result onboarding.RecoveryTransition
	var state string
	err := storage.queryRow(ctx, recordRecoverySQL,
		binding.OperationID, binding.CurrentAttemptID, binding.CurrentCredentialID,
		binding.CurrentCredentialVersion, binding.NextAttemptID, binding.NextCredentialID,
		binding.NextCredentialVersion, record.ExpectedRevision,
		record.RecoveryTokenSHA256, binding.ManifestSHA256, string(binding.Cause)).Scan(
		&state, &result.Revision, &result.DatabaseNow,
	)
	if err != nil {
		return onboarding.RecoveryTransition{}, errRecoveryUnavailable
	}
	result.State = state
	return result, nil
}
