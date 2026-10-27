package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ResolvePortalSetup(
	ctx context.Context, scope tenantScope, query lifecycle.PortalSetupQuery,
) (lifecycle.PortalSetup, error) {
	if !snapshotUsable(ctx, scope) || scope.ActorID() != query.ActorID {
		return lifecycle.PortalSetup{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.PortalSetup{}, err
	}
	var value lifecycle.PortalSetup
	var setupCompletedAt, submittedAt, decisionAt *time.Time
	err = bound.database.QueryRow(ctx, portalSetupReadSQL,
		scope.TenantID(), query.BindingID, query.ActorID, query.PermissionID,
		query.ExpectedTokenVersion, query.ExpectedAuthorizationVersion,
	).Scan(
		&value.BindingID, &value.UserID, &value.SetupState, &value.ApprovalState,
		&value.Revision, &setupCompletedAt, &submittedAt, &decisionAt,
		&value.UpdatedAt,
	)
	if err != nil {
		return lifecycle.PortalSetup{}, mapError(err)
	}
	value.UpdatedAt = value.UpdatedAt.UTC()
	if setupCompletedAt != nil {
		value.SetupCompletedAt = setupCompletedAt.UTC()
	}
	if submittedAt != nil {
		value.SubmittedAt = submittedAt.UTC()
	}
	if decisionAt != nil {
		value.DecisionAt = decisionAt.UTC()
	}
	if value.Validate() != nil || value.BindingID != query.BindingID {
		return lifecycle.PortalSetup{}, lifecycle.ErrUnavailable
	}
	return value, nil
}
