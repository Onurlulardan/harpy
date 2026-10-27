package licenseauthority

import (
	"context"
	"fmt"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func (storage *store) CommitRestoredLicense(
	ctx context.Context,
	commit licenseapp.RestoredLicenseCommit,
) (licenseapp.RestoredLicenseReceipt, error) {
	if !storage.usable(ctx) || !restoredSnapshotMatches(commit) {
		return licenseapp.RestoredLicenseReceipt{}, licenseapp.ErrAuthorityState
	}
	documents, err := encodeSnapshot(commit.Snapshot)
	if err != nil {
		return licenseapp.RestoredLicenseReceipt{}, fmt.Errorf(
			"encode restored license snapshot: %w", err,
		)
	}
	snapshot := commit.Snapshot
	receipt := licenseapp.RestoredLicenseReceipt{}
	err = storage.database.QueryRow(ctx, commitRestoredLicenseSQL,
		commit.OperationID, commit.RestoreReceiptID,
		commit.RestoreReceiptSHA256, commit.LicenseBindingID,
		commit.LicenseID, commit.LicenseeOrganizationID,
		commit.EnvironmentID, commit.PreviousSequence,
		commit.PreviousSHA256, commit.NextSequence, commit.NextSHA256,
		commit.ExpectedRevision, string(snapshot.DeliveryKind()),
		snapshot.Issuer(), snapshot.KeyID(), string(snapshot.Status()),
		snapshot.IssuedAt(), snapshot.NotBefore(), snapshot.RefreshAfter(),
		snapshot.LicenseGrantDeadline(), snapshot.GraceUntil(),
		snapshot.CommercialExpiresAt(), snapshot.Channel(),
		documents.features, documents.limits, snapshot.CanonicalPayload(),
		snapshot.SignedDocument(), documents.projection,
	).Scan(
		&receipt.OperationID, &receipt.RestoreReceiptID,
		&receipt.LicenseBindingID, &receipt.PreviousSequence,
		&receipt.PreviousSHA256, &receipt.AcceptedSequence,
		&receipt.AcceptedSHA256, &receipt.AuthorityRevision, &receipt.Replay,
	)
	if err != nil {
		return licenseapp.RestoredLicenseReceipt{}, fmt.Errorf(
			"persist restored license snapshot: %w", err,
		)
	}
	return receipt, nil
}

func restoredSnapshotMatches(commit licenseapp.RestoredLicenseCommit) bool {
	return commit.OperationID != uuid.Nil && commit.LicenseID != uuid.Nil &&
		commit.EnvironmentID != uuid.Nil && commit.NextSequence > commit.PreviousSequence &&
		commit.NextSHA256 != "" && commit.Snapshot.LicenseID() == commit.LicenseID &&
		commit.Snapshot.OrganizationID() == commit.LicenseeOrganizationID &&
		commit.Snapshot.EnvironmentID() == commit.EnvironmentID &&
		commit.Snapshot.Sequence() == commit.NextSequence &&
		commit.Snapshot.Digest() == commit.NextSHA256
}
