package operator

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
)

type restoredReplayStore struct {
	state billing.LocalRestoredLicenseState
}

func (store restoredReplayStore) GetRestoredLicenseReceipt(
	context.Context,
	billing.UUID,
	billing.UUID,
) (billing.LocalRestoredLicenseState, error) {
	return store.state, nil
}

func TestRestoredLicenseReplayCompletesWithoutCallingSource(t *testing.T) {
	operation, err := billing.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	evidence := restoredLicenseEvidence{
		OperationID: operation, EnvironmentID: billing.NewUUID(),
		RestoreReceiptID:     billing.NewUUID(),
		RestoreReceiptSHA256: strings.Repeat("a", 64),
	}
	state := billing.LocalRestoredLicenseState{
		Found: true, OperationID: evidence.OperationID,
		RestoreReceiptID:     evidence.RestoreReceiptID,
		RestoreReceiptSHA256: evidence.RestoreReceiptSHA256,
		LicenseBindingID:     billing.NewUUID(), LicenseID: billing.NewUUID(),
		LicenseeOrganizationID: billing.NewUUID(),
		EnvironmentID:          evidence.EnvironmentID,
		PreviousSequence:       8, PreviousSHA256: strings.Repeat("b", 64),
		AcceptedSequence: 9, AcceptedSHA256: strings.Repeat("c", 64),
		AuthorityRevision: 11,
		DatabaseNow:       time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	outcome, found, err := readRestoredLicenseReplay(
		context.Background(), restoredReplayStore{state: state}, evidence,
	)
	if err != nil || !found || !outcome.Replayed ||
		outcome.AcceptedSequence != state.AcceptedSequence {
		t.Fatalf("replay outcome=%+v found=%v err=%v", outcome, found, err)
	}
}
