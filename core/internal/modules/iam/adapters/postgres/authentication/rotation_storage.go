package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (bound *boundStore) performRotationStorage(
	ctx context.Context,
	rotation authentication.RefreshRotation,
) (rotationRow, error) {
	portalID, portalRevision := portalArguments(
		rotation.PortalBindingID, rotation.ExpectedPortalBindingRevision,
	)
	selector := rotation.SuccessorSelector
	secretHMAC := rotation.SuccessorSecretHMAC
	var row rotationRow
	err := bound.database.QueryRow(ctx, rotateRefreshSQL,
		rotation.TokenID, rotation.FamilyID, rotation.ExpectedTokenRevision,
		rotation.ExpectedFamilyRevision, rotation.SessionID,
		rotation.ExpectedSessionRevision, rotation.UserID,
		rotation.ExpectedUserRevision, rotation.ExpectedTokenVersion,
		rotation.ExpectedAuthorizationVersion, rotation.AccessModeID,
		rotation.ExpectedAccessModeRevision, portalID, portalRevision,
		rotation.SuccessorID, selector[:], rotation.SuccessorHashKeyID,
		secretHMAC[:], rotation.ExpectedPlatformAuthorizationClock,
		rotation.ExpectedSecurityEpoch[:],
	).Scan(
		&row.outcome, &row.issuedAt, &row.advance.selector,
		&row.advance.expected, &row.advance.revision, &row.advance.sourceSHA256,
		&row.advance.state, &row.advance.projectedAt, &row.advance.retainUntil,
		&row.successor.selector, &row.successor.revision,
		&row.successor.sourceSHA256, &row.successor.state,
		&row.successor.issuedAt, &row.successor.projectedAt,
		&row.successor.familyExpires, &row.successor.retainUntil,
	)
	if err != nil || !row.valid() {
		row.clear()
		return rotationRow{}, errAuthenticationStorage
	}
	return row, nil
}

func (row rotationRow) valid() bool {
	switch row.outcome {
	case "rotated":
		return row.issuedAt != nil && !row.issuedAt.IsZero() &&
			row.advance.valid() && row.successor.valid()
	case "reused", "expired":
		return row.issuedAt == nil && row.advance.valid() &&
			len(row.successor.selector) == 0
	case "rejected":
		return row.issuedAt == nil && len(row.advance.selector) == 0 &&
			len(row.successor.selector) == 0
	default:
		return false
	}
}

func (row *rotationRow) clear() {
	if row == nil {
		return
	}
	clear(row.advance.sourceSHA256)
	clear(row.successor.sourceSHA256)
}
