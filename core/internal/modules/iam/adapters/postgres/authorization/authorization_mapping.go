package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"

func mapActorFact(document *actorDocument) *authorization.ActorFact {
	if document == nil {
		return nil
	}
	return &authorization.ActorFact{
		ID: document.ID, Status: document.Status,
		TokenVersion:         document.TokenVersion,
		AuthorizationVersion: document.AuthorizationVersion,
	}
}

func mapResourceFact(document *catalogDocument) *authorization.ResourceFact {
	if document == nil {
		return nil
	}
	return &authorization.ResourceFact{
		ID: document.ID, State: document.State, Protected: document.Protected,
	}
}

func mapActionFact(document *catalogDocument) *authorization.ActionFact {
	if document == nil {
		return nil
	}
	return &authorization.ActionFact{
		ID: document.ID, State: document.State, Protected: document.Protected,
	}
}

func mapBindingFacts(documents []bindingDocument) []authorization.BindingFact {
	result := make([]authorization.BindingFact, 0, len(documents))
	for _, value := range documents {
		result = append(result, authorization.BindingFact{
			Path:      authorization.SubjectPath(value.Path),
			BindingID: value.BindingID, BindingState: value.BindingState,
			PermissionID:        value.PermissionID,
			PermissionMatchKind: value.PermissionMatchKind,
			PermissionState:     value.PermissionState,
			PermissionProtected: value.PermissionProtected,
			PermissionResource:  value.PermissionResource,
			PermissionAction:    value.PermissionAction,
			ResourceState:       value.ResourceState, ActionState: value.ActionState,
			AssignmentID: value.AssignmentID, AssignmentState: value.AssignmentState,
			RoleID: value.RoleID, RoleStatus: value.RoleStatus,
			RoleScopeKind: value.RoleScopeKind, RoleProtected: value.RoleProtected,
			MembershipID: value.MembershipID, MembershipState: value.MembershipState,
			OrganizationID:     value.OrganizationID,
			OrganizationStatus: value.OrganizationStatus,
			PolicyRoleID:       value.PolicyRoleID,
			PolicyPermissionID: value.PolicyPermissionID,
			PolicyBindingID:    value.PolicyBindingID, PolicyState: value.PolicyState,
		})
	}
	return result
}

func mapImplicationFacts(documents []implicationDocument) []authorization.ImplicationFact {
	result := make([]authorization.ImplicationFact, 0, len(documents))
	for _, value := range documents {
		result = append(result, authorization.ImplicationFact{
			ID: value.ID, Kind: value.Kind, State: value.State,
			SourcePermissionID:     value.SourcePermissionID,
			SourcePermissionState:  value.SourcePermissionState,
			SourceResourceID:       value.SourceResourceID,
			SourceActionID:         value.SourceActionID,
			SourceResourceState:    value.SourceResourceState,
			SourceActionState:      value.SourceActionState,
			ImpliedPermissionID:    value.ImpliedPermissionID,
			ImpliedPermissionState: value.ImpliedPermissionState,
			ImpliedResourceID:      value.ImpliedResourceID,
			ImpliedActionID:        value.ImpliedActionID,
			ImpliedResourceState:   value.ImpliedResourceState,
			ImpliedActionState:     value.ImpliedActionState,
		})
	}
	return result
}
