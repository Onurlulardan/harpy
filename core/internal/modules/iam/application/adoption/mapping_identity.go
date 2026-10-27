package adoption

import "github.com/google/uuid"

type AssignmentMapping struct {
	RoleID       uuid.UUID `json:"role_id"`
	AssignmentID uuid.UUID `json:"assignment_id"`
}

type UserMapping struct {
	UserID              uuid.UUID           `json:"user_id"`
	EmailIdentifierID   *uuid.UUID          `json:"email_identifier_id"`
	PortalIdentifierID  *uuid.UUID          `json:"portal_identifier_id"`
	OperatorModeID      *uuid.UUID          `json:"operator_mode_id"`
	PortalModeID        *uuid.UUID          `json:"portal_mode_id"`
	MembershipID        *uuid.UUID          `json:"membership_id"`
	Assignments         []AssignmentMapping `json:"assignments"`
	PortalBindingID     *uuid.UUID          `json:"portal_binding_id"`
	PortalFormID        *uuid.UUID          `json:"portal_form_id"`
	PortalFormVersionID *uuid.UUID          `json:"portal_form_version_id"`
	PortalFieldID       *uuid.UUID          `json:"portal_field_id"`
}

type IdentityMapping struct {
	Users                 []UserMapping `json:"users"`
	ProtectedRoleID       uuid.UUID     `json:"protected_role_id"`
	ProtectedUserID       uuid.UUID     `json:"protected_user_id"`
	ProtectedPermissionID uuid.UUID     `json:"protected_permission_id"`
	ProtectedBindingID    uuid.UUID     `json:"protected_binding_id"`
	ProtectedPolicyID     uuid.UUID     `json:"protected_policy_id"`
}
