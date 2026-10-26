package licenseauthority

import (
	"context"
	"fmt"
	"time"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	restoredlicense "github.com/Onurlulardan/harpy/core/internal/capabilities/restoredlicense/api"
)

type RestoredLicenseCoordinator struct {
	store RestoredLicenseStore
	keys  KeyResolver
}

func NewRestoredLicenseCoordinator(
	store RestoredLicenseStore,
	keys KeyResolver,
) (*RestoredLicenseCoordinator, error) {
	if store == nil || keys == nil {
		return nil, ErrRestoredLicenseConfiguration
	}
	return &RestoredLicenseCoordinator{store: store, keys: keys}, nil
}

func (coordinator *RestoredLicenseCoordinator) Restore(
	ctx context.Context,
	scope restoredlicense.Scope,
	signedDocument []byte,
	databaseNow time.Time,
) (RestoredLicenseReceipt, error) {
	if ctx == nil || ctx.Err() != nil || databaseNow.IsZero() ||
		databaseNow.Location() != time.UTC {
		return RestoredLicenseReceipt{}, ErrRestoredLicenseScope
	}
	binding := scope.Binding()
	snapshot, err := Verify(ctx, signedDocument, Binding{
		LicenseID:              binding.LicenseID,
		LicenseeOrganizationID: binding.LicenseeOrganizationID,
		EnvironmentID:          binding.EnvironmentID, DeliveryKind: DeliveryKind(binding.DeliveryKind),
	}, databaseNow, coordinator.keys)
	if err != nil || !restoredSnapshotMatches(binding, snapshot) {
		return RestoredLicenseReceipt{}, ErrWrongBinding
	}
	if err = scope.Consume(restoredlicensev1.ReconcileRestoredLicense); err != nil {
		return RestoredLicenseReceipt{}, fmt.Errorf("consume restored license command: %w", err)
	}
	receipt, err := coordinator.store.CommitRestoredLicense(ctx, restoredCommit(
		binding, scope.Revision(), snapshot,
	))
	if err != nil {
		return RestoredLicenseReceipt{}, fmt.Errorf("commit restored license: %w", err)
	}
	if !validRestoredReceipt(binding, scope.Revision(), receipt) {
		return RestoredLicenseReceipt{}, ErrRestoredLicenseReceipt
	}
	return receipt, nil
}

func restoredSnapshotMatches(binding restoredlicense.Binding, snapshot VerifiedSnapshot) bool {
	return snapshot.LicenseID() == binding.LicenseID &&
		snapshot.OrganizationID() == binding.LicenseeOrganizationID &&
		snapshot.EnvironmentID() == binding.EnvironmentID &&
		string(snapshot.DeliveryKind()) == binding.DeliveryKind &&
		snapshot.Issuer() == binding.Issuer && snapshot.Sequence() == binding.NextSequence &&
		snapshot.Digest() == binding.NextSHA256
}

func restoredCommit(
	binding restoredlicense.Binding,
	revision int64,
	snapshot VerifiedSnapshot,
) RestoredLicenseCommit {
	return RestoredLicenseCommit{
		OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
		RestoreReceiptID:     binding.RestoreReceiptID,
		RestoreReceiptSHA256: binding.RestoreReceiptSHA256,
		LicenseBindingID:     binding.LicenseBindingID, LicenseID: binding.LicenseID,
		LicenseeOrganizationID: binding.LicenseeOrganizationID,
		PreviousSequence:       binding.PreviousSequence, PreviousSHA256: binding.PreviousSHA256,
		NextSequence: binding.NextSequence, NextSHA256: binding.NextSHA256,
		ExpectedRevision: revision, Snapshot: snapshot,
	}
}

func validRestoredReceipt(
	binding restoredlicense.Binding,
	revision int64,
	receipt RestoredLicenseReceipt,
) bool {
	return receipt.OperationID == binding.OperationID &&
		receipt.RestoreReceiptID == binding.RestoreReceiptID &&
		receipt.LicenseBindingID == binding.LicenseBindingID &&
		receipt.PreviousSequence == binding.PreviousSequence &&
		receipt.PreviousSHA256 == binding.PreviousSHA256 &&
		receipt.AcceptedSequence == binding.NextSequence &&
		receipt.AcceptedSHA256 == binding.NextSHA256 &&
		receipt.AuthorityRevision >= revision
}
