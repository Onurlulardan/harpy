package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func (store *permissionRequestStore) WriteTransitionAudit(
	ctx context.Context,
	scope tenantScope,
	record permissionrequest.TransitionRecord,
) (int64, error) {
	if record.Validate() != nil || !permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != record.ActorID {
		return 0, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return 0, err
	}
	var reviewTask any
	if record.ReviewTaskID != uuid.Nil {
		reviewTask = record.ReviewTaskID
	}
	var fromState any
	if record.From != "" {
		fromState = string(record.From)
	}
	var clock int64
	err = bound.database.QueryRow(
		ctx, permissionRequestTransitionAuditSQL,
		record.OperationID, record.RequestSHA256[:], record.GroupID,
		record.ActorID, reviewTask, string(record.Command), fromState,
		string(record.To), record.Revision,
	).Scan(&clock)
	if err != nil || clock < 0 {
		return 0, permissionrequest.ErrDecisionEffects
	}
	return clock, nil
}
