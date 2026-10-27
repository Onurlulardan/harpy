package api

import "time"

type OnboardingReceiptKind string

const (
	ReceiptMigrationClaimReady OnboardingReceiptKind = "migration_claim_ready"
	ReceiptMigrationCompleted  OnboardingReceiptKind = "migration_completed"
	ReceiptActivationCompleted OnboardingReceiptKind = "activation_completed"
	ReceiptAmbiguousReconciled OnboardingReceiptKind = "ambiguous_reconciled"
	ReceiptFailure             OnboardingReceiptKind = "failure"
)

type OnboardingReceipt struct {
	ReceiptID      UUID
	OperationID    UUID
	AttemptID      UUID
	Kind           OnboardingReceiptKind
	SHA256         string
	EvidenceSHA256 string
	OccurredAt     time.Time
}

func (receipt OnboardingReceipt) IsValid() bool {
	kindValid := receipt.Kind == ReceiptMigrationClaimReady ||
		receipt.Kind == ReceiptMigrationCompleted ||
		receipt.Kind == ReceiptActivationCompleted ||
		receipt.Kind == ReceiptAmbiguousReconciled ||
		receipt.Kind == ReceiptFailure
	return receipt.ReceiptID != NilUUID && receipt.OperationID != NilUUID &&
		receipt.AttemptID != NilUUID && kindValid &&
		validOnboardingDigest(receipt.SHA256) &&
		validOnboardingDigest(receipt.EvidenceSHA256) && microsecondUTC(receipt.OccurredAt)
}
