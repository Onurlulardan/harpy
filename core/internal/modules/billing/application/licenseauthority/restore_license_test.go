package licenseauthority

import (
	"context"
	"strings"
	"testing"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	restoredlicense "github.com/Onurlulardan/harpy/core/internal/capabilities/restoredlicense/api"
	"github.com/google/uuid"
)

type restoredStore struct {
	calls  int
	commit RestoredLicenseCommit
}

func (store *restoredStore) CommitRestoredLicense(
	_ context.Context,
	commit RestoredLicenseCommit,
) (RestoredLicenseReceipt, error) {
	store.calls++
	store.commit = commit
	return RestoredLicenseReceipt{
		OperationID: commit.OperationID, RestoreReceiptID: commit.RestoreReceiptID,
		LicenseBindingID: commit.LicenseBindingID,
		PreviousSequence: commit.PreviousSequence, PreviousSHA256: commit.PreviousSHA256,
		AcceptedSequence: commit.NextSequence, AcceptedSHA256: commit.NextSHA256,
		AuthorityRevision: commit.ExpectedRevision + 1,
	}, nil
}

func TestRestoreAcceptsOnlyReceiptBoundHigherSnapshot(t *testing.T) {
	fixture := newAuthorityFixture(t)
	document := fixture.document(t)
	snapshot, err := Verify(
		context.Background(), document, fixture.binding, fixture.now, fixture.resolver(),
	)
	if err != nil {
		t.Fatal(err)
	}
	grant := restoredGrant(t, fixture, snapshot)
	store := &restoredStore{}
	coordinator, _ := NewRestoredLicenseCoordinator(store, fixture.resolver())
	err = restoredlicense.Run(grant, fixture.now, func(scope restoredlicense.Scope) error {
		_, restoreErr := coordinator.Restore(context.Background(), scope, document, fixture.now)
		return restoreErr
	})
	if err != nil || store.calls != 1 || store.commit.NextSequence != snapshot.Sequence() ||
		store.commit.RestoreReceiptSHA256 != grant.Binding.RestoreReceiptSHA256 {
		t.Fatalf("restored snapshot commit drifted: calls=%d err=%v", store.calls, err)
	}
}

func TestRestoreRejectsWrongDigestWithoutMutation(t *testing.T) {
	fixture := newAuthorityFixture(t)
	document := fixture.document(t)
	snapshot, _ := Verify(
		context.Background(), document, fixture.binding, fixture.now, fixture.resolver(),
	)
	grant := restoredGrant(t, fixture, snapshot)
	grant.Binding.NextSHA256 = strings.Repeat("f", 64)
	store := &restoredStore{}
	coordinator, _ := NewRestoredLicenseCoordinator(store, fixture.resolver())
	err := restoredlicense.Run(grant, fixture.now, func(scope restoredlicense.Scope) error {
		_, restoreErr := coordinator.Restore(context.Background(), scope, document, fixture.now)
		return restoreErr
	})
	if err == nil || store.calls != 0 {
		t.Fatalf("wrong digest mutated restored head: calls=%d err=%v", store.calls, err)
	}
}

func restoredGrant(
	t *testing.T,
	fixture authorityFixture,
	snapshot VerifiedSnapshot,
) restoredlicense.Grant {
	t.Helper()
	operationID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return restoredlicense.Grant{Revision: 4, LeaseUntil: fixture.now.Add(1), Binding: restoredlicense.Binding{
		OperationID: operationID, EnvironmentID: fixture.binding.EnvironmentID,
		RestoreReceiptID: uuid.New(), RestoreReceiptSHA256: strings.Repeat("a", 64),
		LicenseBindingID: uuid.New(), LicenseID: fixture.binding.LicenseID,
		LicenseeOrganizationID: fixture.binding.LicenseeOrganizationID,
		DeliveryKind:           string(fixture.binding.DeliveryKind), Issuer: snapshot.Issuer(),
		PreviousSequence: snapshot.Sequence() - 1, PreviousSHA256: strings.Repeat("b", 64),
		NextSequence: snapshot.Sequence(), NextSHA256: snapshot.Digest(),
		ManifestSHA256: restoredlicensev1.ManifestSHA256(),
	}}
}
