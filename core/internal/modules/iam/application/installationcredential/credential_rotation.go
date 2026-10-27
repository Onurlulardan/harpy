package installationcredential

import (
	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type RotationBinding struct {
	OperationID, TenantID, FirstOwnerUserID         uuid.UUID
	CurrentCredentialID, NextCredentialID           uuid.UUID
	CurrentAttemptID, NextAttemptID                 uuid.UUID
	CurrentCredentialVersion, NextCredentialVersion int64
	Cause                                           credentialdeliveryv1.Cause
	ManifestSHA256                                  string
}

func (binding RotationBinding) valid() bool {
	validCause := binding.Cause == credentialdeliveryv1.CauseClaimLeaseExpired ||
		binding.Cause == credentialdeliveryv1.CauseOutcomeUnknown
	return binding.OperationID != uuid.Nil && binding.TenantID != uuid.Nil &&
		binding.FirstOwnerUserID != uuid.Nil && binding.CurrentCredentialID != uuid.Nil &&
		binding.NextCredentialID != uuid.Nil &&
		binding.NextCredentialVersion == binding.CurrentCredentialVersion+1 &&
		binding.CurrentAttemptID != uuid.Nil && binding.NextAttemptID != uuid.Nil &&
		validCause && binding.ManifestSHA256 == credentialdeliveryv1.ManifestSHA256()
}

type RotateCommand struct {
	Binding    RotationBinding
	Credential Credential
}

type RotateReceipt struct {
	OperationID, TenantID, UserID      uuid.UUID
	PreviousCredentialID, CredentialID uuid.UUID
	CredentialVersion                  int64
	CredentialRevision                 int64
	PreviousMarkerRevision             int64
	MarkerRevision                     int64
}

func (receipt RotateReceipt) valid(binding RotationBinding) bool {
	return receipt.OperationID == binding.OperationID && receipt.TenantID == binding.TenantID &&
		receipt.UserID == binding.FirstOwnerUserID &&
		receipt.PreviousCredentialID == binding.CurrentCredentialID &&
		receipt.CredentialID == binding.NextCredentialID &&
		receipt.CredentialVersion == binding.NextCredentialVersion &&
		receipt.CredentialRevision > 1 &&
		receipt.PreviousMarkerRevision > 1 && receipt.MarkerRevision == 1
}
