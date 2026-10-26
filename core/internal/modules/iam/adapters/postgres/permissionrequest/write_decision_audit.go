package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) WriteDecisionAudit(
	ctx context.Context,
	scope tenantScope,
	record permissionrequest.DecisionRecord,
) error {
	if record.Validate() != nil || !permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != record.ReviewerID {
		return permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return err
	}
	var written bool
	err = bound.database.QueryRow(
		ctx, permissionRequestDecisionAuditSQL,
		record.OperationID, record.RequestSHA256[:], record.GroupID,
		record.ReviewerID, record.ReviewTaskID, string(record.Decision),
		record.Result.Code, record.Result.Revision,
		record.Result.AuthorizationClock,
	).Scan(&written)
	if err != nil || !written {
		return permissionrequest.ErrDecisionEffects
	}
	return nil
}
