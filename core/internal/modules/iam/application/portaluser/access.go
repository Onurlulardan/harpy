package portaluser

import "github.com/google/uuid"

type AccessBinding struct {
	PrincipalID, BindingID, TenantID       uuid.UUID
	FormID, FormVersionID, SubjectRecordID uuid.UUID
	IdentifierFieldID                      uuid.UUID
	SubjectType                            string
	RoleNames                              []string
	TokenVersion, AuthorizationVersion     int64
	AuthorizationClock, BindingRevision    int64
}

func (value AccessBinding) Valid() bool {
	return value.PrincipalID != uuid.Nil && value.BindingID != uuid.Nil &&
		value.TenantID != uuid.Nil && value.FormID != uuid.Nil &&
		value.FormVersionID != uuid.Nil && value.SubjectRecordID != uuid.Nil &&
		value.IdentifierFieldID != uuid.Nil && value.SubjectType != "" && len(value.RoleNames) > 0 &&
		value.TokenVersion > 0 && value.AuthorizationVersion > 0 &&
		value.AuthorizationClock > 0 && value.BindingRevision > 0
}
