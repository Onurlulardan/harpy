package httptransport

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuthentication = errors.New("analytics HTTP authentication failed")
	ErrForbidden      = errors.New("analytics HTTP operation is forbidden")
	ErrInvalid        = errors.New("analytics HTTP input is invalid")
	ErrNotFound       = errors.New("analytics HTTP resource was not found")
	ErrConflict       = errors.New("analytics HTTP state conflicts")
	ErrRateLimited    = errors.New("analytics HTTP rate limit was exceeded")
	ErrUnavailable    = errors.New("analytics HTTP authority is unavailable")
)

type Permission string

type AccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          Permission
	Operation           string
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}

type Access struct {
	principalID, tenantID                                  uuid.UUID
	permission                                             Permission
	operation                                              string
	tokenVersion, authorizationVersion, authorizationClock int64
	stepUpAuthenticatedAt                                  time.Time
}

func NewAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time, permission Permission, operation string,
) (Access, error) {
	value := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock:    authorizationClock,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC().Truncate(time.Second),
		permission:            permission, operation: operation}
	if !stepUpAuthenticatedAt.IsZero() && stepUpAuthenticatedAt.Nanosecond() != 0 {
		return Access{}, ErrInvalid
	}
	if !value.ValidFor(permission, operation) {
		return Access{}, ErrInvalid
	}
	return value, nil
}

func (access Access) ValidFor(permission Permission, operation string) bool {
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.permission == permission &&
		access.operation == operation && validPermissionOperation(permission, operation)
}

func (access Access) PrincipalID() uuid.UUID { return access.principalID }
func (access Access) TenantID() uuid.UUID    { return access.tenantID }
func (access Access) TokenVersion() int64    { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 {
	return access.authorizationVersion
}
func (access Access) AuthorizationClock() int64 { return access.authorizationClock }
func (access Access) StepUpAuthenticatedAt() time.Time {
	return access.stepUpAuthenticatedAt
}
