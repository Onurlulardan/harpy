package adoption

import "github.com/google/uuid"

type SourceProfileField struct {
	ResourceKind string
	Key          string
	ValueKind    string
	Required     bool
}

type SourceLayoutField struct {
	ResourceKind string
	FieldKey     string
	SectionKey   string
	Position     int
}

type SourceInvitation struct {
	ID     uuid.UUID
	Email  string
	RoleID string
	State  string
}

type Snapshot struct {
	TenantID           uuid.UUID
	Organizations      []SourceOrganization
	Roles              []SourceRole
	Users              []SourceUser
	Resources          []SourceResource
	Actions            []SourceAction
	Permissions        []SourcePermission
	PermissionRequests []SourcePermissionRequest
	ProfileFields      []SourceProfileField
	LayoutFields       []SourceLayoutField
	Invitations        []SourceInvitation
}
