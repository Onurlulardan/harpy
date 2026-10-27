package iam

import (
	"context"

	workflowinitiator "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workflowinitiator"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *WorkflowAccessRuntime) VerifyWorkflowInitiator(
	ctx context.Context,
	scope platformtx.TenantScope,
	request workflowinitiator.Request,
) (workflowinitiator.Evidence, error) {
	if runtime == nil || runtime.shared == nil || request.Validate() != nil ||
		ctx == nil || ctx.Err() != nil || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.NetworkAllowed(ctx) || scope.Mode() != platformtx.Command ||
		scope.OperationID() != "workflow.runtime.start" ||
		scope.ActorID() != uuid.UUID(request.ActorID) ||
		scope.TenantID() != uuid.UUID(request.TenantID) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return workflowinitiator.Evidence{}, workflowinitiator.ErrInvalid
	}
	probe, err := authorization.NewCurrentActorRequest(
		"workflow.instance", "start", uuid.UUID(request.ActorID),
	)
	if err != nil {
		return workflowinitiator.Evidence{}, workflowinitiator.ErrInvalid
	}
	decision, err := runtime.shared.authorizer.CurrentActorDecision(ctx, scope, probe)
	permissionID, allowed := decision.PermissionID()
	if err != nil || decision.Validate() != nil || !decision.Allowed() || !allowed ||
		decision.AuthorizationVersion() < 1 || decision.AuthorizationClock() < 1 {
		return workflowinitiator.Evidence{}, ErrWorkflowAccessForbidden
	}
	evidence := workflowinitiator.Evidence{Request: request,
		AuthorizationVersion: uint64(decision.AuthorizationVersion()),
		AuthorizationClock:   uint64(decision.AuthorizationClock()),
		PermissionID:         [16]byte(permissionID)}
	evidence.SHA256 = workflowInitiatorSHA256(evidence)
	return evidence, nil
}

var _ workflowinitiator.Verifier = (*WorkflowAccessRuntime)(nil)
