package transaction

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

type CredentialIdentifierKind string

const (
	CredentialEmailIdentifier  CredentialIdentifierKind = "email_v1"
	CredentialPortalIdentifier CredentialIdentifierKind = "portal_v1"
)

func (kind CredentialIdentifierKind) valid() bool {
	return kind == CredentialEmailIdentifier || kind == CredentialPortalIdentifier
}

type CredentialLookup struct {
	operationID string
	purpose     CredentialPurpose
	locatorKind CredentialLocatorKind
	identifier  CredentialIdentifierKind
	keyID       string
	tenantID    uuid.UUID
	key         [32]byte
	keyLength   uint8
}

func NewLoginCredentialLookup(
	operationID string,
	tenantID uuid.UUID,
	identifier CredentialIdentifierKind,
	keyID string,
	lookupKey [32]byte,
) (CredentialLookup, error) {
	lookup := credentialLookup(
		operationID, CredentialLogin, CredentialLoginDirectory, lookupKey, 32,
	)
	lookup.tenantID, lookup.identifier, lookup.keyID = tenantID, identifier, keyID
	return validateCredentialLookup(lookup)
}

func NewPasswordResetCredentialLookup(
	operationID string,
	tenantID uuid.UUID,
	identifier CredentialIdentifierKind,
	keyID string,
	lookupKey [32]byte,
) (CredentialLookup, error) {
	lookup := credentialLookup(
		operationID, CredentialPasswordReset, CredentialLoginDirectory, lookupKey, 32,
	)
	lookup.tenantID, lookup.identifier, lookup.keyID = tenantID, identifier, keyID
	return validateCredentialLookup(lookup)
}

func NewInitialPasswordCredentialLookup(
	operationID string,
	tenantID uuid.UUID,
	identifier CredentialIdentifierKind,
	keyID string,
	lookupKey [32]byte,
) (CredentialLookup, error) {
	lookup := credentialLookup(
		operationID, CredentialInitialPassword, CredentialLoginDirectory, lookupKey, 32,
	)
	lookup.tenantID, lookup.identifier, lookup.keyID = tenantID, identifier, keyID
	return validateCredentialLookup(lookup)
}

func NewRefreshCredentialLookup(
	operationID string,
	selector [16]byte,
) (CredentialLookup, error) {
	var key [32]byte
	copy(key[:], selector[:])
	return validateCredentialLookup(credentialLookup(
		operationID, CredentialRefresh, CredentialRefreshLocator, key, 16,
	))
}

func NewTenantRouteCredentialLookup(
	operationID string,
	purpose CredentialPurpose,
	tenantID uuid.UUID,
	routeKey [32]byte,
) (CredentialLookup, error) {
	lookup := credentialLookup(
		operationID, purpose, CredentialTenantRoute, routeKey, 32,
	)
	lookup.tenantID = tenantID
	return validateCredentialLookup(lookup)
}

func NewSessionCredentialLookup(
	operationID string,
	tenantID uuid.UUID,
	sessionBinding [32]byte,
) (CredentialLookup, error) {
	lookup := credentialLookup(
		operationID, CredentialReauthentication,
		CredentialSessionLocator, sessionBinding, 32,
	)
	lookup.tenantID = tenantID
	return validateCredentialLookup(lookup)
}

func credentialLookup(
	operationID string,
	purpose CredentialPurpose,
	locatorKind CredentialLocatorKind,
	key [32]byte,
	keyLength uint8,
) CredentialLookup {
	return CredentialLookup{
		operationID: operationID, purpose: purpose, locatorKind: locatorKind,
		key: key, keyLength: keyLength,
	}
}

func validateCredentialLookup(lookup CredentialLookup) (CredentialLookup, error) {
	if !lookup.valid() {
		return CredentialLookup{}, fmt.Errorf("credential lookup is invalid")
	}
	return lookup, nil
}

func (lookup CredentialLookup) OperationID() string { return lookup.operationID }

func (lookup CredentialLookup) Purpose() CredentialPurpose { return lookup.purpose }

func (lookup CredentialLookup) LocatorKind() CredentialLocatorKind { return lookup.locatorKind }

func (lookup CredentialLookup) IdentifierKind() CredentialIdentifierKind {
	return lookup.identifier
}

func (lookup CredentialLookup) KeyID() string { return lookup.keyID }

func (lookup CredentialLookup) TenantID() uuid.UUID { return lookup.tenantID }

func (lookup CredentialLookup) Key() []byte {
	if !lookup.valid() {
		return nil
	}
	result := make([]byte, lookup.keyLength)
	copy(result, lookup.key[:lookup.keyLength])
	return result
}

func (lookup CredentialLookup) valid() bool {
	if !operationIDPattern.MatchString(lookup.operationID) ||
		!lookup.purpose.valid() || !lookup.locatorKind.validFor(lookup.purpose) ||
		lookup.key == [32]byte{} {
		return false
	}
	if lookup.locatorKind == CredentialRefreshLocator {
		return lookup.identifier == "" && lookup.keyID == "" &&
			lookup.tenantID == uuid.Nil && lookup.keyLength == 16 &&
			[16]byte(lookup.key[16:]) == [16]byte{}
	}
	if lookup.locatorKind == CredentialLoginDirectory {
		return lookup.tenantID != uuid.Nil && lookup.keyLength == 32 &&
			(lookup.identifier == CredentialEmailIdentifier ||
				lookup.identifier == CredentialPortalIdentifier) &&
			validCredentialKeyID(lookup.keyID)
	}
	return lookup.identifier == "" && lookup.keyID == "" &&
		lookup.tenantID != uuid.Nil && lookup.keyLength == 32
}

func validCredentialKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}
