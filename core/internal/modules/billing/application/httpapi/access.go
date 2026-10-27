package httpapi

import (
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Permission string

const (
	PermissionBillingRead            Permission = "billing.read"
	PermissionBillingManage          Permission = "billing.manage"
	PermissionBillingUsageRead       Permission = "billing.usage.read"
	PermissionPlatformBillingRead    Permission = "platform.billing.read"
	PermissionPlatformBillingManage  Permission = "platform.billing.manage"
	PermissionPlatformBillingAudit   Permission = "platform.billing.audit.read"
	PermissionPlatformBillingRevenue Permission = "platform.billing.revenue.read"
)

type AccessScope uint8

const (
	AccessScopeTenant AccessScope = iota + 1
	AccessScopeGlobal
)

type Access struct {
	principalID          uuid.UUID
	tenantID             uuid.UUID
	authorizationClock   int64
	tokenVersion         int64
	authorizationVersion int64
	permission           Permission
	scope                AccessScope
	globalCommand        platformtx.GlobalCommandCapability
	globalCommandAccess  platformtx.GlobalCommandAccess
	globalSnapshot       platformtx.GlobalSnapshotCapability
	globalMode           globalAccessMode
}

type globalAccessMode uint8

const (
	globalCommandAccess globalAccessMode = iota + 1
	globalSnapshotAccess
)

func NewTenantAccess(
	principalID uuid.UUID,
	tenantID uuid.UUID,
	authorizationClock int64,
	permission Permission,
) (Access, error) {
	if principalID == uuid.Nil || tenantID == uuid.Nil || authorizationClock < 1 ||
		!tenantPermission(permission) {
		return Access{}, fmt.Errorf("billing tenant access is invalid")
	}
	return Access{
		principalID: principalID, tenantID: tenantID,
		authorizationClock: authorizationClock,
		tokenVersion:       1, authorizationVersion: 1,
		permission: permission, scope: AccessScopeTenant,
	}, nil
}

func NewTenantAccessEvidence(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission Permission,
) (Access, error) {
	access, err := NewTenantAccess(
		principalID, tenantID, authorizationClock, permission,
	)
	if err != nil || tokenVersion < 1 || authorizationVersion < 1 {
		return Access{}, fmt.Errorf("billing tenant access evidence is invalid")
	}
	access.tokenVersion = tokenVersion
	access.authorizationVersion = authorizationVersion
	return access, nil
}

func NewGlobalAccess(
	principalID uuid.UUID,
	authorizationClock int64,
	permission Permission,
) (Access, error) {
	if principalID == uuid.Nil || authorizationClock < 1 || !globalPermission(permission) {
		return Access{}, fmt.Errorf("billing global access is invalid")
	}
	return Access{
		principalID: principalID, authorizationClock: authorizationClock,
		tokenVersion: 1, authorizationVersion: 1,
		permission: permission, scope: AccessScopeGlobal,
	}, nil
}

func NewGlobalAccessEvidence(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission Permission,
	capability platformtx.GlobalCommandCapability,
) (Access, error) {
	access, err := NewGlobalAccess(principalID, authorizationClock, permission)
	if err != nil || tokenVersion < 1 || authorizationVersion < 1 {
		return Access{}, fmt.Errorf("billing global access evidence is invalid")
	}
	access.tokenVersion = tokenVersion
	access.authorizationVersion = authorizationVersion
	access.globalCommand = capability
	access.globalMode = globalCommandAccess
	return access, nil
}

func NewGlobalSnapshotAccessEvidence(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission Permission,
	capability platformtx.GlobalSnapshotCapability,
) (Access, error) {
	access, err := NewGlobalAccess(principalID, authorizationClock, permission)
	if err != nil || tokenVersion < 1 || authorizationVersion < 1 {
		return Access{}, fmt.Errorf("billing global snapshot access evidence is invalid")
	}
	access.tokenVersion = tokenVersion
	access.authorizationVersion = authorizationVersion
	access.globalSnapshot = capability
	access.globalMode = globalSnapshotAccess
	return access, nil
}

func (access Access) ValidFor(permission Permission) bool {
	return access.principalID != uuid.Nil && access.authorizationClock > 0 &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.permission == permission &&
		(access.scope == AccessScopeTenant && access.tenantID != uuid.Nil ||
			access.scope == AccessScopeGlobal && access.tenantID == uuid.Nil)
}

func (access Access) PrincipalID() uuid.UUID      { return access.principalID }
func (access Access) TenantID() uuid.UUID         { return access.tenantID }
func (access Access) AuthorizationClock() int64   { return access.authorizationClock }
func (access Access) TokenVersion() int64         { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) Scope() AccessScope          { return access.scope }

func (access Access) GlobalCommandCapability() (
	platformtx.GlobalCommandCapability,
	bool,
) {
	if command, _, valid := access.globalCommandAccess.Capabilities(); valid &&
		access.globalMode == globalCommandAccess {
		return command, true
	}
	return access.globalCommand, access.globalMode == globalCommandAccess
}

func (access Access) GlobalSnapshotCapability() (
	platformtx.GlobalSnapshotCapability,
	bool,
) {
	return access.globalSnapshot, access.globalMode == globalSnapshotAccess
}

func (access Access) GlobalSnapshotCapabilityFor(
	operation string,
) (platformtx.GlobalSnapshotCapability, bool) {
	capability, available := access.GlobalSnapshotCapability()
	valid := available && capability.Allows(
		access.principalID, operation, string(access.permission),
		access.authorizationClock, access.authorizationVersion,
	)
	return capability, valid
}

func tenantPermission(permission Permission) bool {
	return permission == PermissionBillingRead || permission == PermissionBillingManage ||
		permission == PermissionBillingUsageRead
}

func globalPermission(permission Permission) bool {
	return permission == PermissionPlatformBillingRead ||
		permission == PermissionPlatformBillingManage ||
		permission == PermissionPlatformBillingAudit ||
		permission == PermissionPlatformBillingRevenue
}
