package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) RevokeSession(
	ctx context.Context,
	scope authentication.CredentialScope,
	mutation authentication.LogoutMutation,
) (authentication.LogoutResult, error) {
	bound, err := store.bindRefresh(ctx, scope)
	if err != nil {
		return authentication.LogoutResult{}, err
	}
	row, err := bound.revokeSession(ctx, mutation)
	if err != nil {
		return authentication.LogoutResult{}, err
	}
	defer clear(row.projection.sourceSHA256)
	if row.outcome == "revoked" {
		if err := bound.advanceLocator(ctx, scope, row.projection); err != nil {
			return authentication.LogoutResult{}, err
		}
	}
	return authentication.LogoutResult{
		Outcome: authentication.LogoutOutcome(row.outcome),
	}, nil
}

type sessionRevokeRow struct {
	outcome    string
	projection locatorAdvance
}

func (bound *boundStore) revokeSession(
	ctx context.Context,
	mutation authentication.LogoutMutation,
) (sessionRevokeRow, error) {
	var row sessionRevokeRow
	err := bound.database.QueryRow(ctx, revokeSessionSQL,
		mutation.TokenID, mutation.FamilyID, mutation.ExpectedTokenRevision,
		mutation.ExpectedFamilyRevision, mutation.SessionID,
		mutation.ExpectedSessionRevision, mutation.UserID,
		mutation.ExpectedUserRevision,
	).Scan(
		&row.outcome, &row.projection.selector, &row.projection.expected,
		&row.projection.revision, &row.projection.sourceSHA256,
		&row.projection.state, &row.projection.projectedAt,
		&row.projection.retainUntil,
	)
	valid := row.outcome == "terminal" && len(row.projection.selector) == 0 ||
		row.outcome == "revoked" && row.projection.valid()
	if err != nil || !valid {
		clear(row.projection.sourceSHA256)
		return sessionRevokeRow{}, errAuthenticationStorage
	}
	return row, nil
}
