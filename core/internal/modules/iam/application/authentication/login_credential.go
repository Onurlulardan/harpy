package authentication

import (
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type LoginCredential struct {
	IdentifierID          uuid.UUID
	UserID                uuid.UUID
	UserRevision          int64
	TokenVersion          int64
	AuthorizationVersion  int64
	AccessModeID          uuid.UUID
	AccessModeRevision    int64
	AccessMode            AccessMode
	PortalBindingID       uuid.UUID
	PortalBindingRevision int64
	CredentialVersion     int64
	CredentialRevision    int64
	Password              PasswordCredential
	CredentialCreatedAt   time.Time
}

func (credential LoginCredential) validate(
	scope platformtx.CredentialTenantScope,
) error {
	if credential.IdentifierID == uuid.Nil || credential.UserID == uuid.Nil ||
		credential.IdentifierID != scope.SourceIdentifierID() ||
		credential.UserID != scope.SourceUserID() || credential.UserRevision <= 0 ||
		credential.TokenVersion <= 0 || credential.AuthorizationVersion <= 0 ||
		credential.AccessModeID == uuid.Nil || credential.AccessModeRevision <= 0 ||
		credential.CredentialVersion <= 0 || credential.CredentialRevision <= 0 ||
		credential.CredentialCreatedAt.IsZero() || credential.passwordShapeInvalid() {
		return ErrAuthenticationStorage
	}
	kind := scope.SourceIdentifierKind()
	operator := kind == platformtx.CredentialEmailIdentifier &&
		credential.AccessMode == AccessModeOperator &&
		credential.PortalBindingID == uuid.Nil && credential.PortalBindingRevision == 0
	portal := kind == platformtx.CredentialPortalIdentifier &&
		credential.AccessMode == AccessModePortal &&
		credential.PortalBindingID != uuid.Nil && credential.PortalBindingRevision > 0
	if !operator && !portal {
		return ErrAuthenticationStorage
	}
	return nil
}

func (credential LoginCredential) passwordShapeInvalid() bool {
	return credential.Password.invalid()
}
