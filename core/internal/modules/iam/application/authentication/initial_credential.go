package authentication

import (
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type InitialPasswordCredential struct {
	IdentifierID, UserID, AccessModeID       uuid.UUID
	InstallationOperationID, CredentialID    uuid.UUID
	UserRevision, TokenVersion               int64
	AuthorizationVersion, AccessModeRevision int64
	CredentialVersion, CredentialRevision    int64
	Password                                 PasswordCredential
	CreatedAt                                time.Time
}

func (credential InitialPasswordCredential) validate(
	scope platformtx.CredentialTenantScope,
) error {
	if credential.IdentifierID == uuid.Nil || credential.UserID == uuid.Nil ||
		credential.AccessModeID == uuid.Nil || credential.InstallationOperationID == uuid.Nil ||
		credential.CredentialID == uuid.Nil || credential.UserID != scope.SourceUserID() ||
		credential.IdentifierID != scope.SourceIdentifierID() || credential.UserRevision < 1 ||
		credential.TokenVersion < 1 || credential.AuthorizationVersion < 1 ||
		credential.AccessModeRevision < 1 || credential.CredentialVersion < 1 ||
		credential.CredentialRevision < 1 || credential.CreatedAt.IsZero() ||
		credential.Password.Algorithm != AlgorithmArgon2idV1 ||
		credential.Password.invalid() {
		return ErrAuthenticationStorage
	}
	return nil
}

type initialBinding struct {
	tenantID, userID, operationID, credentialID uuid.UUID
	credentialVersion                           int64
}

func (credential InitialPasswordCredential) binding(
	tenantID uuid.UUID,
) initialBinding {
	return initialBinding{tenantID: tenantID, userID: credential.UserID,
		operationID:       credential.InstallationOperationID,
		credentialID:      credential.CredentialID,
		credentialVersion: credential.CredentialVersion}
}
