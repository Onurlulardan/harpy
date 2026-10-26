package operator

import "github.com/Onurlulardan/harpy/core/internal/modules/billing"

type restoredLicenseEvidence struct {
	OperationID          billing.UUID
	EnvironmentID        billing.UUID
	RestoreReceiptID     billing.UUID
	RestoreReceiptSHA256 string
}

type restoredLicenseOutcome struct {
	OperationID       billing.UUID
	EnvironmentID     billing.UUID
	PreviousSequence  int64
	AcceptedSequence  int64
	AuthorityRevision int64
	Replayed          bool
}
