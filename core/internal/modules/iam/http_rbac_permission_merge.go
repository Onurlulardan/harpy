package iam

import (
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	"github.com/google/uuid"
)

func rbacMergePermissionUpdate(
	document rbacPermissionUpdateDocument,
	current RBACPermissionView,
	mutation RBACMutation,
) (RBACPermissionUpdate, bool) {
	if document.ResourceID == nil && document.Target == nil &&
		document.UserID == nil && document.RoleID == nil &&
		document.OrganizationID == nil && document.ActionIDs == nil {
		return RBACPermissionUpdate{}, false
	}
	resourceID, resourceValid := decodeUUID(current.ResourceID)
	target, targetID, targetValid := rbacTargetFromView(current)
	actionIDs, actionsValid := rbacActionIDsFromView(current)
	if !resourceValid || !targetValid || !actionsValid {
		return RBACPermissionUpdate{}, false
	}
	if document.ResourceID != nil {
		resourceID, resourceValid = decodeUUID(*document.ResourceID)
	}
	if document.ActionIDs != nil {
		actionIDs, actionsValid = decodeUUIDList(*document.ActionIDs)
	}
	target, targetID, targetValid = rbacMergedTarget(
		document, target, targetID,
	)
	if !resourceValid || !actionsValid || !targetValid {
		return RBACPermissionUpdate{}, false
	}
	return RBACPermissionUpdate{
		RBACMutation: mutation, ResourceID: resourceID,
		Target: target, TargetID: targetID, ActionIDs: actionIDs,
	}, true
}

func rbacMergedTarget(
	document rbacPermissionUpdateDocument,
	current rbacdomain.AssignmentTarget,
	currentID uuid.UUID,
) (rbacdomain.AssignmentTarget, uuid.UUID, bool) {
	selectorPresent := document.UserID != nil || document.RoleID != nil ||
		document.OrganizationID != nil
	if document.Target == nil && !selectorPresent {
		return current, currentID, true
	}
	target := current
	if document.Target != nil {
		target = rbacdomain.AssignmentTarget(*document.Target)
	}
	if !selectorPresent {
		return target, currentID, target == current
	}
	return rbacTarget(
		string(target), document.UserID, document.RoleID, document.OrganizationID,
	)
}
