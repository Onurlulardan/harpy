package iam

import "net/http"

func (handler *RBACHTTP) createPermission(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureAuthentication)
		return
	}
	key, valid := rbacCreateKey(r)
	if !valid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	document, err := decodeIdentityDocument[rbacPermissionCreateDocument](w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	resourceID, resourceValid := decodeUUID(document.ResourceID)
	actionIDs, actionsValid := decodeUUIDList(document.ActionIDs)
	target, targetID, targetValid := rbacTarget(
		document.Target, document.UserID, document.RoleID, document.OrganizationID,
	)
	if !resourceValid || !actionsValid || !targetValid {
		handler.writeFailure(w, r, identityFailureInvalid)
		return
	}
	view, err := handler.application.CreatePermission(
		r.Context(), claims, RBACPermissionCreate{
			ResourceID: resourceID, Target: target, TargetID: targetID,
			ActionIDs: actionIDs, IdempotencyKey: key,
		},
	)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	handler.writeRecord(w, r, http.StatusCreated, view.Revision, view)
}
