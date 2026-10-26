package iam

import "github.com/google/uuid"

func validRBACResourceView(view RBACResourceView) bool {
	_, idValid := decodeUUID(view.ID)
	return idValid && rbacCatalogKeyValid(view.Slug) &&
		rbacDisplayValid(view.Name) && view.Revision > 0 &&
		(view.State == "active" || view.State == "retired")
}

func validRBACActionView(view RBACActionView) bool {
	_, idValid := decodeUUID(view.ID)
	return idValid && rbacCatalogKeyValid(view.Slug) &&
		rbacDisplayValid(view.Name) && view.Revision > 0 &&
		(view.State == "active" || view.State == "retired")
}

func validRBACPermissionView(view RBACPermissionView) bool {
	_, idValid := decodeUUID(view.ID)
	_, resourceValid := decodeUUID(view.ResourceID)
	targets := 0
	for _, target := range []*string{view.UserID, view.RoleID, view.OrganizationID} {
		if target != nil {
			if parsed, err := uuid.Parse(*target); err != nil || parsed == uuid.Nil {
				return false
			}
			targets++
		}
	}
	if !idValid || !resourceValid || targets != 1 || view.Revision < 1 ||
		len(view.Actions) < 1 || len(view.Actions) > 100 ||
		!validRBACResourceView(view.Resource) {
		return false
	}
	for _, action := range view.Actions {
		if !validRBACActionView(action) {
			return false
		}
	}
	return view.State == "active" || view.State == "retired"
}
