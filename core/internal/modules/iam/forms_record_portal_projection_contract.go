package iam

import "github.com/google/uuid"

type PortalSubjectLinkProjectionInput struct {
	PrincipalID, TenantID                     uuid.UUID
	TokenVersion, AuthorizationVersion, Clock int64
	FormID, FormVersionID                     uuid.UUID
	RecordIDs                                 []uuid.UUID
}

func (input PortalSubjectLinkProjectionInput) valid() bool {
	return input.PrincipalID != uuid.Nil && input.TenantID != uuid.Nil &&
		input.TokenVersion > 0 && input.AuthorizationVersion > 0 && input.Clock > 0 &&
		input.FormID != uuid.Nil && input.FormVersionID != uuid.Nil &&
		len(input.RecordIDs) > 0 && len(input.RecordIDs) <= 500
}

type PortalSubjectLink struct {
	ID          uuid.UUID
	Status      string
	AccountType string
	Revision    int64
}

func (link PortalSubjectLink) Valid() bool {
	return link.ID != uuid.Nil && link.Revision > 0 &&
		(link.Status == "ACTIVE" || link.Status == "PENDING" || link.Status == "INACTIVE") &&
		(link.AccountType == "portal" || link.AccountType == "operator_portal")
}
