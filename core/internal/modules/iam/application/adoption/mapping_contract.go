package adoption

import "github.com/google/uuid"

const MappingContractVersion = "harpy/iam-v1-adoption-mapping/v1"

type PermissionMapping struct {
	SourcePermissionID uuid.UUID `json:"source_permission_id"`
	ResourceID         uuid.UUID `json:"resource_id"`
	ActionID           uuid.UUID `json:"action_id"`
	PermissionID       uuid.UUID `json:"permission_id"`
	BindingID          uuid.UUID `json:"binding_id"`
}

type ResourceMapping struct {
	SourceID     uuid.UUID `json:"source_id"`
	TargetID     uuid.UUID `json:"target_id"`
	ResourceKey  string    `json:"resource_key"`
	OwnerDomain  string    `json:"owner_domain"`
	ResourceKind string    `json:"resource_kind"`
}

type ActionMapping struct {
	SourceID  uuid.UUID `json:"source_id"`
	TargetID  uuid.UUID `json:"target_id"`
	ActionKey string    `json:"action_key"`
}

type PermissionRequestMapping struct {
	SourceRequestID uuid.UUID `json:"source_request_id"`
	GroupID         uuid.UUID `json:"group_id"`
	RequestID       uuid.UUID `json:"request_id"`
}

type InvitationMapping struct {
	SourceInvitationID uuid.UUID `json:"source_invitation_id"`
	InvitationID       uuid.UUID `json:"invitation_id"`
	TokenID            uuid.UUID `json:"token_id"`
}

type ProfileSchemaMapping struct {
	ResourceKind    string    `json:"resource_kind"`
	SchemaVersionID uuid.UUID `json:"schema_version_id"`
}

type ProfileFieldMapping struct {
	ResourceKind string    `json:"resource_kind"`
	SourceKey    string    `json:"source_key"`
	FieldID      uuid.UUID `json:"field_id"`
}

type LayoutFieldMapping struct {
	ResourceKind string    `json:"resource_kind"`
	SourceKey    string    `json:"source_key"`
	LayoutID     uuid.UUID `json:"layout_id"`
	FieldID      uuid.UUID `json:"field_id"`
}

type Mapping struct {
	ContractVersion    string                     `json:"contract_version"`
	TenantID           uuid.UUID                  `json:"tenant_id"`
	Identity           IdentityMapping            `json:"identity"`
	Resources          []ResourceMapping          `json:"resources"`
	Actions            []ActionMapping            `json:"actions"`
	Permissions        []PermissionMapping        `json:"permissions"`
	PermissionRequests []PermissionRequestMapping `json:"permission_requests"`
	Invitations        []InvitationMapping        `json:"invitations"`
	ProfileSchemas     []ProfileSchemaMapping     `json:"profile_schemas"`
	ProfileFields      []ProfileFieldMapping      `json:"profile_fields"`
	LayoutFields       []LayoutFieldMapping       `json:"layout_fields"`
}
