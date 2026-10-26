package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
)

func readRestoredLicenseReplay(
	ctx context.Context,
	store billing.LocalRestoredLicenseReceiptStore,
	evidence restoredLicenseEvidence,
) (restoredLicenseOutcome, bool, error) {
	state, err := billing.ReadLocalRestoredLicenseReceipt(
		ctx,
		store,
		billing.LocalRestoredLicenseLookup{
			OperationID:          evidence.OperationID,
			EnvironmentID:        evidence.EnvironmentID,
			RestoreReceiptID:     evidence.RestoreReceiptID,
			RestoreReceiptSHA256: evidence.RestoreReceiptSHA256,
		},
	)
	if err != nil {
		return restoredLicenseOutcome{}, false, err
	}
	if !state.Found {
		return restoredLicenseOutcome{}, false, nil
	}
	return restoredLicenseOutcome{
		OperationID:       state.OperationID,
		EnvironmentID:     state.EnvironmentID,
		PreviousSequence:  state.PreviousSequence,
		AcceptedSequence:  state.AcceptedSequence,
		AuthorityRevision: state.AuthorityRevision,
		Replayed:          true,
	}, true, nil
}
