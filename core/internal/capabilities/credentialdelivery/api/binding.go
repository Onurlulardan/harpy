package api

import (
	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type Binding struct {
	OperationID              uuid.UUID
	TenantID                 uuid.UUID
	FirstOwnerUserID         uuid.UUID
	CurrentCredentialID      uuid.UUID
	CurrentCredentialVersion int64
	NextCredentialID         uuid.UUID
	NextCredentialVersion    int64
	CurrentAttemptID         uuid.UUID
	NextAttemptID            uuid.UUID
	Cause                    credentialdeliveryv1.Cause
	ManifestSHA256           string
}

func (binding Binding) valid() bool {
	return validV7(binding.OperationID) && binding.TenantID != uuid.Nil &&
		binding.FirstOwnerUserID != uuid.Nil && binding.CurrentCredentialID != uuid.Nil &&
		binding.NextCredentialID != uuid.Nil &&
		binding.CurrentCredentialID != binding.NextCredentialID &&
		binding.CurrentCredentialVersion > 0 &&
		binding.NextCredentialVersion == binding.CurrentCredentialVersion+1 &&
		binding.CurrentAttemptID != uuid.Nil && binding.NextAttemptID != uuid.Nil &&
		binding.CurrentAttemptID != binding.NextAttemptID && validCause(binding.Cause) &&
		binding.ManifestSHA256 == credentialdeliveryv1.ManifestSHA256()
}

func validCause(cause credentialdeliveryv1.Cause) bool {
	return cause == credentialdeliveryv1.CauseClaimLeaseExpired ||
		cause == credentialdeliveryv1.CauseOutcomeUnknown
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}
