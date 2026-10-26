package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	humantaskapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/humantask"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type humanTaskCurrentPermission struct {
	authorizer authorization.Authorizer
}

func (port humanTaskCurrentPermission) ResolveCurrentHumanPermission(
	ctx context.Context,
	scope platformtx.TenantScope,
	actorID [16]byte,
	action string,
) (humantaskapp.CurrentPermissionEvidence, error) {
	request, err := authorization.NewCurrentActorRequest(
		humanTaskResource, action, uuid.UUID(actorID),
	)
	if err != nil {
		return humantaskapp.CurrentPermissionEvidence{}, humantaskapp.ErrInvalid
	}
	decision, err := port.authorizer.CurrentActorDecision(ctx, scope, request)
	permissionID, exists := decision.PermissionID()
	if err != nil || decision.Validate() != nil || !decision.Allowed() || !exists {
		return humantaskapp.CurrentPermissionEvidence{}, humantaskapp.ErrForbidden
	}
	return humantaskapp.CurrentPermissionEvidence{
		PermissionID:         [16]byte(permissionID),
		AuthorizationVersion: decision.AuthorizationVersion(),
	}, nil
}

const humanTaskResource = "workflow.human_tasks"
