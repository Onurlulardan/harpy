package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) mutationContext(
	claims accesstoken.Claims,
	command string,
	target uuid.UUID,
	idempotencyKey string,
	payload any,
) (uuid.UUID, [32]byte, rbacdomain.StepUpEvidence, error) {
	operationID := rbacOperationID(claims, command, idempotencyKey)
	requestSHA, err := identityRequestSHA(payload)
	if err != nil {
		return uuid.Nil, [32]byte{}, rbacdomain.StepUpEvidence{}, ErrRBACHTTPInvalid
	}
	stepUp, valid := application.stepUp(claims, command, target)
	if !valid {
		return uuid.Nil, [32]byte{}, rbacdomain.StepUpEvidence{}, ErrRBACHTTPForbidden
	}
	return operationID, requestSHA, stepUp, nil
}

func rbacCatalogMutationView(
	current rbacdomain.Resource,
	displayName string,
	description string,
	revision int64,
) RBACResourceView {
	current.DisplayName, current.Description = displayName, description
	current.Revision = revision
	return rbacResourceView(current)
}

func rbacActionMutationView(
	current rbacdomain.Action,
	displayName string,
	description string,
	revision int64,
) RBACActionView {
	current.DisplayName, current.Description = displayName, description
	current.Revision = revision
	return rbacActionView(current)
}

func rbacPermissionMutationView(
	id uuid.UUID,
	resource RBACResourceView,
	actions []RBACActionView,
	target rbacdomain.AssignmentTarget,
	targetID uuid.UUID,
	revision int64,
) RBACPermissionView {
	view := RBACPermissionView{
		ID: id.String(), Target: string(target), ResourceID: resource.ID,
		Resource: resource, Actions: actions, State: "active", Revision: revision,
	}
	setRBACPermissionTarget(&view, target, rbacdomain.ID(targetID))
	return view
}
