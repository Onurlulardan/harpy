package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateUserPortal(
	user SourceUser, mapping UserMapping, seen map[string]uuid.UUID,
) {
	console, portal := sourceAccessModes(user.AccountType)
	if !console && !portal {
		state.add(issue("account_type_invalid", "portal", user.ID.String()))
		return
	}
	if console != !nilUUID(mapping.OperatorModeID) || portal != !nilUUID(mapping.PortalModeID) {
		state.add(issue("access_mode_mapping_invalid", "portal", user.ID.String()))
	}
	if !portal {
		if mapping.PortalIdentifierID != nil || mapping.PortalBindingID != nil {
			state.add(issue("portal_mapping_unexpected", "portal", user.ID.String()))
		}
		return
	}
	value, err := state.normalizers.portal(user.PortalLoginIdentifier)
	if err != nil {
		state.add(issue("portal_identifier_invalid", "identifiers", user.ID.String()))
	} else {
		if prior, collision := seen[value]; collision && prior != user.ID {
			state.add(issue("portal_collision", "identifiers", value))
		}
		seen[value] = user.ID
	}
	if nilUUID(mapping.PortalIdentifierID) || nilUUID(mapping.PortalBindingID) {
		state.add(issue("portal_mapping_missing", "portal", user.ID.String()))
	}
	state.validatePortalSubject(user, mapping)
	if !validPortalWorkflowState(user.PortalWorkflowStatus) {
		state.add(issue("portal_workflow_state_invalid", "portal", user.ID.String()))
	}
}

func sourceAccessModes(accountType string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(accountType)) {
	case "", "operator":
		return true, false
	case "portal":
		return false, true
	case "operator_portal":
		return true, true
	default:
		return false, false
	}
}

func (state *validationState) validatePortalSubject(user SourceUser, mapping UserMapping) {
	if user.PortalSubjectType == "operator" && user.PortalSubjectID != nil &&
		*user.PortalSubjectID == user.ID && user.PortalProfileFormSlug == "users" {
		if mapping.PortalFormID != nil || mapping.PortalFormVersionID != nil ||
			mapping.PortalFieldID != nil {
			state.add(issue("portal_self_mapping_invalid", "portal", user.ID.String()))
		}
		return
	}
	if user.PortalSubjectType == "" || user.PortalSubjectID == nil ||
		strings.TrimSpace(user.PortalProfileFormSlug) == "" ||
		nilUUID(mapping.PortalFormID) || nilUUID(mapping.PortalFormVersionID) ||
		nilUUID(mapping.PortalFieldID) {
		state.add(issue("portal_subject_mapping_missing", "portal", user.ID.String()))
	}
}

func validPortalWorkflowState(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "DRAFT", "PENDING_APPROVAL", "REVISION_REQUESTED", "APPROVED", "REJECTED":
		return true
	default:
		return false
	}
}
