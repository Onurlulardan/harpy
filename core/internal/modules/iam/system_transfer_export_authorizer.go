package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type systemTransferExportAuthorizer struct {
	runtime *WorkerAuthorizationRuntime
}

func newSystemTransferExportAuthorizer(
	runtime *WorkerAuthorizationRuntime,
) (*systemTransferExportAuthorizer, error) {
	if runtime == nil {
		return nil, systemtransfer.ErrExportUnavailable
	}
	return &systemTransferExportAuthorizer{runtime: runtime}, nil
}

func (authorizer *systemTransferExportAuthorizer) AuthorizeExportCheckpoint(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.ExportAttempt,
) (publicresource.ActorEvidence, error) {
	if scope.Mode() != platformtx.Command ||
		scope.OperationID() != SystemExportExecuteOperation {
		return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
	}
	return authorizer.authorize(ctx, scope, attempt)
}

func (authorizer *systemTransferExportAuthorizer) authorizeSnapshot(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.ExportAttempt,
) (publicresource.ActorEvidence, error) {
	if !systemTransferExportSnapshotScopeUsable(ctx, scope) {
		return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
	}
	return authorizer.authorize(ctx, scope, attempt)
}

func (authorizer *systemTransferExportAuthorizer) authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.ExportAttempt,
) (publicresource.ActorEvidence, error) {
	if authorizer == nil || authorizer.runtime == nil || ctx == nil || ctx.Err() != nil ||
		attempt.Validate() != nil || scope.TenantID() != attempt.TenantID ||
		scope.ActorID() != attempt.ActorID {
		return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
	}
	plan, err := loadSystemTransferExportPlan(ctx, scope, attempt)
	if err != nil {
		return publicresource.ActorEvidence{}, err
	}
	actions, err := systemTransferExportActions(plan)
	resource, resourceErr := attempt.Resource.AuthorityResource()
	if err != nil || resourceErr != nil {
		return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
	}
	var clock int64
	for index, action := range actions {
		allowed, permissionID, tokenVersion, authorizationVersion, currentClock, authErr :=
			authorizer.runtime.AuthorizePinned(
				ctx, scope, resource, action, attempt.AuthorizationVersion,
			)
		if authErr != nil || !allowed || permissionID == uuid.Nil ||
			tokenVersion != attempt.TokenVersion ||
			authorizationVersion != attempt.AuthorizationVersion || currentClock < 1 ||
			(index > 0 && currentClock != clock) {
			return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
		}
		clock = currentClock
	}
	evidence, err := publicresource.NewActorEvidence(
		publicresource.ID(attempt.TenantID), publicresource.ID(attempt.ActorID),
		attempt.TokenVersion, attempt.AuthorizationVersion,
	)
	if err != nil {
		return publicresource.ActorEvidence{}, systemtransfer.ErrExportAuthorizationRevoked
	}
	return evidence, nil
}

func systemTransferExportClassificationAction(
	classification profileapp.Classification,
) (string, error) {
	switch classification {
	case profileapp.ClassificationPublic, profileapp.ClassificationInternal:
		return "read", nil
	case profileapp.ClassificationConfidential:
		return "read.confidential", nil
	case profileapp.ClassificationRestricted:
		return "read.restricted", nil
	default:
		return "", systemtransfer.ErrExportAuthorizationRevoked
	}
}

var _ systemtransfer.ExportCheckpointAuthorizer = (*systemTransferExportAuthorizer)(nil)
