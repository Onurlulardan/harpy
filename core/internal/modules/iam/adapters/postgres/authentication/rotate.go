package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) Rotate(
	ctx context.Context,
	scope authentication.CredentialScope,
	rotation authentication.RefreshRotation,
) (authentication.RefreshRotated, error) {
	bound, err := store.bindRefresh(ctx, scope)
	if err != nil {
		return authentication.RefreshRotated{}, err
	}
	row, err := bound.performRotationStorage(ctx, rotation)
	if err != nil {
		return authentication.RefreshRotated{}, err
	}
	defer row.clear()
	if row.outcome != "rejected" {
		if err := bound.advanceLocator(ctx, scope, row.advance); err != nil {
			return authentication.RefreshRotated{}, err
		}
	}
	if row.outcome == "rotated" {
		if err := bound.createLocator(
			ctx, scope, rotation.SuccessorID, rotation.FamilyID, row.successor,
		); err != nil {
			return authentication.RefreshRotated{}, err
		}
	}
	result := authentication.RefreshRotated{
		Outcome: authentication.RefreshOutcome(row.outcome),
	}
	if row.issuedAt != nil {
		result.IssuedAt = row.issuedAt.UTC()
	}
	return result, nil
}
