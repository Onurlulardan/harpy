package adoption

import "github.com/google/uuid"

type SourceOrganization struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Name     string
	Slug     string
	Status   string
	ParentID string
}

type SourceRole struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	OrganizationID string
	Default        bool
}

type SourceUser struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	Email                 string
	PasswordHash          string
	RoleIDs               []string
	OrganizationID        string
	Status                string
	TokenVersion          int64
	AccountType           string
	PortalSubjectType     string
	PortalSubjectID       *uuid.UUID
	PortalProfileFormSlug string
	PortalLoginIdentifier string
	PortalIdentifierField string
	PortalSetupRequired   bool
	PortalSetupCompleted  bool
	PortalWorkflowStatus  string
}
