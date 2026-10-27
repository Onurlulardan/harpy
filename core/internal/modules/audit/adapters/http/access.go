package audithttp

import (
	"context"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/google/uuid"
)

var operationPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,127}$`)

type Access struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	permission                                             application.Permission
	operation                                              string
}

func NewAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission application.Permission, operation string,
) (Access, error) {
	access := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission, operation: operation}
	if !access.ValidFor(permission, operation) {
		return Access{}, application.ErrForbidden
	}
	return access, nil
}

func (access Access) ValidFor(permission application.Permission, operation string) bool {
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.permission == permission &&
		access.operation == operation && operationPattern.MatchString(operation)
}

func (access Access) PrincipalID() uuid.UUID      { return access.principalID }
func (access Access) TenantID() uuid.UUID         { return access.tenantID }
func (access Access) TokenVersion() int64         { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) AuthorizationClock() int64   { return access.authorizationClock }

type AccessRequest struct {
	AuthorizationValues, TenantIDValues, TenantSlugValues []string
	Permission                                            application.Permission
	Operation                                             string
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}
