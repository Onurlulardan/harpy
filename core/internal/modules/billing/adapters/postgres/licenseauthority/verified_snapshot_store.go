package licenseauthority

import (
	"context"
	"fmt"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func (storage *store) PersistVerifiedSnapshot(
	ctx context.Context,
	claim licenseapp.RefreshClaim,
	snapshot licenseapp.VerifiedSnapshot,
) error {
	if !storage.usable(ctx) {
		return licenseapp.ErrAuthorityState
	}
	documents, err := encodeSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("encode verified license snapshot: %w", err)
	}
	var outcome string
	err = storage.database.QueryRow(ctx, acceptSnapshotSQL,
		claim.OperationID, claim.LeaseToken, snapshot.LicenseID(),
		snapshot.OrganizationID(), snapshot.EnvironmentID(),
		string(snapshot.DeliveryKind()), snapshot.Issuer(), snapshot.KeyID(),
		snapshot.Digest(), string(snapshot.Status()), snapshot.Sequence(),
		snapshot.IssuedAt(), snapshot.NotBefore(), snapshot.RefreshAfter(),
		snapshot.LicenseGrantDeadline(), snapshot.GraceUntil(),
		snapshot.CommercialExpiresAt(), snapshot.Channel(),
		documents.features, documents.limits, snapshot.CanonicalPayload(),
		snapshot.SignedDocument(), documents.projection,
	).Scan(&outcome)
	if err != nil {
		return fmt.Errorf("persist verified license snapshot: %w", err)
	}
	switch outcome {
	case "accepted", "noop":
		return nil
	case "sequence_rollback":
		return licenseapp.ErrSequenceRollback
	case "sequence_conflict":
		return licenseapp.ErrSequenceConflict
	default:
		return licenseapp.ErrAuthorityState
	}
}
