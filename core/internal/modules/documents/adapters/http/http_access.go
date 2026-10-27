package documentshttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuthentication        = errors.New("Documents HTTP authentication failed")
	ErrForbidden             = errors.New("Documents HTTP operation is forbidden")
	ErrInvalid               = errors.New("Documents HTTP input is invalid")
	ErrNotFound              = errors.New("Documents HTTP resource was not found")
	ErrConflict              = errors.New("Documents HTTP state conflicts")
	ErrContentRejected       = errors.New("Documents HTTP content was rejected")
	ErrScanUnavailable       = errors.New("Documents HTTP scanner is unavailable")
	ErrClientUpgradeRequired = errors.New("Documents HTTP client upgrade is required")
	ErrUnavailable           = errors.New("Documents HTTP authority is unavailable")
)

type AccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
	PurposeTarget       string
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}

type Access struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	permission, operation                                  string
	stepUpAuthenticatedAt                                  time.Time
}

func NewAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion, authorizationClock int64, permission, operation string,
) (Access, error) {
	return NewAccessWithStepUp(principalID, tenantID, tokenVersion,
		authorizationVersion, authorizationClock, time.Time{}, permission, operation)
}

func NewAccessWithStepUp(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion, authorizationClock int64, stepUpAuthenticatedAt time.Time,
	permission, operation string,
) (Access, error) {
	access := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission, operation: operation,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC()}
	if !access.ValidFor(permission, operation) {
		return Access{}, ErrInvalid
	}
	return access, nil
}

func (access Access) ValidFor(permission, operation string) bool {
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.permission == permission &&
		access.operation == operation && permission != "" && operation != "" &&
		(access.stepUpAuthenticatedAt.IsZero() || access.stepUpAuthenticatedAt.Nanosecond() == 0)
}

func (access Access) PrincipalID() uuid.UUID { return access.principalID }
func (access Access) TenantID() uuid.UUID    { return access.tenantID }
func (access Access) TokenVersion() int64    { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 {
	return access.authorizationVersion
}
func (access Access) AuthorizationClock() int64 { return access.authorizationClock }
func (access Access) Permission() string        { return access.permission }
func (access Access) Operation() string         { return access.operation }
func (access Access) StepUpAuthenticatedAt() time.Time {
	return access.stepUpAuthenticatedAt
}

func (handler *Handler) accessFor(r *http.Request, permission, operation string) (Access, error) {
	if handler == nil || r == nil {
		return Access{}, ErrUnavailable
	}
	return handler.access.Verify(r.Context(), AccessRequest{
		AuthorizationValues: r.Header.Values("Authorization"),
		TenantIDValues:      r.Header.Values("X-Tenant-ID"),
		TenantSlugValues:    r.Header.Values("X-Tenant-Slug"),
		Permission:          permission, Operation: operation,
		PurposeTarget: accessPurposeTarget(operation, r),
	})
}

func accessPurposeTarget(operation string, request *http.Request) string {
	if operation == operationItemPurge && request != nil {
		return request.PathValue("document_id")
	}
	return ""
}
