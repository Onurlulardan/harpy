package iam

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func rbacResourceView(value rbacdomain.Resource) RBACResourceView {
	return RBACResourceView{
		ID: value.ID.String(), Name: value.DisplayName, Slug: value.Key,
		Description: value.Description, OwnerDomain: value.OwnerDomain,
		Kind: value.Kind, Protected: value.Protected,
		State: value.State, Revision: value.Revision,
	}
}

func rbacActionView(value rbacdomain.Action) RBACActionView {
	return RBACActionView{
		ID: value.ID.String(), Name: value.DisplayName, Slug: value.Key,
		Description: value.Description, Protected: value.Protected,
		State: value.State, Revision: value.Revision,
	}
}

func (application *rbacHTTPApplication) assignmentView(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityEvidenceSource,
	permissionID uuid.UUID,
	assignment rbacdomain.Assignment,
) (RBACPermissionView, error) {
	queryEvidence := rbacQueryEvidence(evidence, permissionID)
	resource, err := application.runtime.resources.Get(
		ctx, scope, rbacdomain.GetQuery{
			ID: rbacdomain.ID(assignment.ResourceID), Evidence: queryEvidence,
		},
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	view := RBACPermissionView{
		ID: assignment.ID.String(), Target: string(assignment.Target),
		ResourceID: assignment.ResourceID.String(), Resource: rbacResourceView(resource),
		Actions: make([]RBACActionView, 0, len(assignment.Items)),
		State:   assignment.State, Revision: assignment.Revision,
	}
	setRBACPermissionTarget(&view, assignment.Target, assignment.TargetID)
	for _, item := range assignment.Items {
		action, actionErr := application.runtime.actions.Get(
			ctx, scope, rbacdomain.GetQuery{ID: item.ActionID, Evidence: queryEvidence},
		)
		if actionErr != nil {
			return RBACPermissionView{}, actionErr
		}
		view.Actions = append(view.Actions, rbacActionView(action))
	}
	return view, nil
}

func setRBACPermissionTarget(
	view *RBACPermissionView,
	target rbacdomain.AssignmentTarget,
	targetID rbacdomain.ID,
) {
	if view == nil {
		return
	}
	wire := targetID.String()
	switch target {
	case rbacdomain.AssignmentUser:
		view.UserID = &wire
	case rbacdomain.AssignmentRole:
		view.RoleID = &wire
	case rbacdomain.AssignmentOrganization:
		view.OrganizationID = &wire
	}
}

func rbacEffectiveView(value rbacdomain.EffectivePermission) RBACEffectiveView {
	view := RBACEffectiveView{
		PermissionID: value.PermissionID.String(), MatchKind: value.MatchKind,
		ResourceSlug: value.ResourceKey, ActionSlug: value.ActionKey,
		Source: string(value.Source), SourceID: value.SourceID.String(),
	}
	if value.ResourceID.Valid() {
		wire := value.ResourceID.String()
		view.ResourceID = &wire
	}
	if value.ActionID.Valid() {
		wire := value.ActionID.String()
		view.ActionID = &wire
	}
	return view
}
