package iam

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type SystemTransferAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Resource            SystemTransferResource
	Permission          string
	Operation           string
}

type SystemTransferAccessVerifier interface {
	VerifySystemTransfer(context.Context, SystemTransferAccessRequest) (
		SystemTransferAccess, error,
	)
}

type SystemTransferAccess struct {
	principalID, tenantID              uuid.UUID
	tokenVersion, authorizationVersion int64
	authorizationClock                 int64
	resource                           SystemTransferResource
	permission, operation              string
}

func NewSystemTransferAccess(principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	resource SystemTransferResource, permission, operation string,
) (SystemTransferAccess, error) {
	value := SystemTransferAccess{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, resource: resource,
		permission: permission, operation: operation}
	if !value.ValidFor(resource, permission, operation) {
		return SystemTransferAccess{}, ErrSystemTransferForbidden
	}
	return value, nil
}

func (access SystemTransferAccess) ValidFor(resource SystemTransferResource,
	permission, operation string) bool {
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.resource == resource &&
		permission == resource.Permission(systemTransferAction(operation)) &&
		access.permission == permission && access.operation == operation
}

func (access SystemTransferAccess) PrincipalID() uuid.UUID { return access.principalID }
func (access SystemTransferAccess) TenantID() uuid.UUID    { return access.tenantID }
func (access SystemTransferAccess) TokenVersion() int64    { return access.tokenVersion }
func (access SystemTransferAccess) AuthorizationVersion() int64 {
	return access.authorizationVersion
}
func (access SystemTransferAccess) AuthorizationClock() int64 {
	return access.authorizationClock
}

func systemTransferAction(operation string) string {
	if strings.HasPrefix(operation, "iam.record_export.") {
		return "export"
	}
	if strings.HasPrefix(operation, "iam.record_import.") {
		return "import"
	}
	return ""
}
