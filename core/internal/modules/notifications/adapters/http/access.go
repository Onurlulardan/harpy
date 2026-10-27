package notificationhttp

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
)

var (
	ErrAuthentication = errors.New("notification HTTP authentication failed")
	ErrForbidden      = errors.New("notification HTTP operation is forbidden")
	ErrInvalid        = errors.New("notification HTTP input is invalid")
	ErrNotFound       = errors.New("notification HTTP resource was not found")
	ErrConflict       = errors.New("notification HTTP operation conflicts")
	ErrUnavailable    = errors.New("notification HTTP authority is unavailable")
)

type Permission string

const (
	PermissionInboxRead       Permission = "notifications.inbox.read"
	PermissionInboxUpdate     Permission = "notifications.inbox.update"
	PermissionInboxDelete     Permission = "notifications.inbox.delete"
	PermissionPreferencesRead Permission = "notifications.preferences.read"
	PermissionPreferencesEdit Permission = "notifications.preferences.update"
	PermissionSend            Permission = "notifications.send"
)

type AccessRequest struct {
	AuthorizationValues     []string
	CookieValues            []string
	TenantIDValues          []string
	TenantSlugValues        []string
	CSRFValues              []string
	OriginValues            []string
	ConsumerManifestValues  []string
	ConsumerSignatureValues []string
	Permission              Permission
	Operation               string
	RequireSession          bool
	RequireCSRF             bool
	RequireHTTPSOrigin      bool
	RequireSecureTransport  bool
	RequireSignedConsumer   bool
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}

type Access struct {
	principalID, tenantID              uuid.UUID
	authorizationVersion, tokenVersion int64
	permission                         Permission
	operation, origin                  string
	sessionAuthenticated, csrfVerified bool
	consumerManifestDigest             [32]byte
	consumerManifestVerified           bool
}

func NewLegacyAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion int64, origin string, manifestDigest [32]byte,
	sessionAuthenticated bool,
) (Access, error) {
	value := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		permission: PermissionInboxRead, operation: OperationLegacyConnect,
		origin: origin, sessionAuthenticated: sessionAuthenticated,
		consumerManifestDigest: manifestDigest, consumerManifestVerified: true}
	if !value.ValidFor(PermissionInboxRead, OperationLegacyConnect) {
		return Access{}, ErrInvalid
	}
	return value, nil
}

func NewAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion int64, permission Permission, operation, origin string,
	sessionAuthenticated, csrfVerified bool,
) (Access, error) {
	value := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		permission: permission, operation: operation, origin: origin,
		sessionAuthenticated: sessionAuthenticated, csrfVerified: csrfVerified}
	if !value.ValidFor(permission, operation) {
		return Access{}, ErrInvalid
	}
	return value, nil
}

func (value Access) ValidFor(permission Permission, operation string) bool {
	if value.principalID == uuid.Nil || value.tenantID == uuid.Nil ||
		value.tokenVersion < 1 || value.authorizationVersion < 1 ||
		value.permission != permission || value.operation != operation ||
		!validPermissionOperation(permission, operation) {
		return false
	}
	requireSession, requireCSRF, requireOrigin := operationSecurity(operation)
	return (!requireSession || value.sessionAuthenticated) &&
		(!requireCSRF || value.csrfVerified) && (!requireOrigin || validOrigin(value.origin)) &&
		(operation != OperationLegacyConnect ||
			(value.consumerManifestVerified && value.consumerManifestDigest != [32]byte{}))
}

func (value Access) PrincipalID() uuid.UUID           { return value.principalID }
func (value Access) TenantID() uuid.UUID              { return value.tenantID }
func (value Access) Origin() string                   { return value.origin }
func (value Access) ConsumerManifestDigest() [32]byte { return value.consumerManifestDigest }

func validOrigin(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" &&
		parsed.User == nil && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}
