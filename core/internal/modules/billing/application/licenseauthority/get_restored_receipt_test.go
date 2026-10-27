package licenseauthority

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type restoredReceiptStoreStub struct {
	state RestoredLicenseState
	err   error
}

func (store restoredReceiptStoreStub) GetRestoredLicenseReceipt(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (RestoredLicenseState, error) {
	return store.state, store.err
}

func TestGetRestoredLicenseReceiptAcceptsOnlyExactPersistedEvidence(t *testing.T) {
	operation := uuid.Must(uuid.NewV7())
	environment, receipt := uuid.New(), uuid.New()
	digest := strings.Repeat("a", 64)
	state := RestoredLicenseState{
		Found: true, OperationID: operation,
		RestoreReceiptID: receipt, RestoreReceiptSHA256: digest,
		LicenseBindingID: uuid.New(), LicenseID: uuid.New(),
		LicenseeOrganizationID: uuid.New(), EnvironmentID: environment,
		PreviousSequence: 4, PreviousSHA256: strings.Repeat("b", 64),
		AcceptedSequence: 5, AcceptedSHA256: strings.Repeat("c", 64),
		AuthorityRevision: 7,
		DatabaseNow:       time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	lookup := RestoredLicenseLookup{
		OperationID: operation, EnvironmentID: environment,
		RestoreReceiptID: receipt, RestoreReceiptSHA256: digest,
	}
	got, err := GetRestoredLicenseReceipt(
		context.Background(), restoredReceiptStoreStub{state: state}, lookup,
	)
	if err != nil || !got.Found || got.AcceptedSequence != 5 {
		t.Fatalf("persisted receipt=%+v err=%v", got, err)
	}
	state.RestoreReceiptSHA256 = strings.Repeat("d", 64)
	if _, err = GetRestoredLicenseReceipt(
		context.Background(), restoredReceiptStoreStub{state: state}, lookup,
	); !errors.Is(err, ErrWrongBinding) {
		t.Fatalf("mismatched receipt error=%v", err)
	}
}

func TestGetRestoredLicenseReceiptAllowsMissingOperation(t *testing.T) {
	lookup := RestoredLicenseLookup{
		OperationID: uuid.Must(uuid.NewV7()), EnvironmentID: uuid.New(),
		RestoreReceiptID: uuid.New(), RestoreReceiptSHA256: strings.Repeat("a", 64),
	}
	state := RestoredLicenseState{
		DatabaseNow: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	got, err := GetRestoredLicenseReceipt(
		context.Background(), restoredReceiptStoreStub{state: state}, lookup,
	)
	if err != nil || got.Found {
		t.Fatalf("missing receipt=%+v err=%v", got, err)
	}
}
