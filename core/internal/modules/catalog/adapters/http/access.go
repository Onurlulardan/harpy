package cataloghttp

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAuthentication = errors.New("catalog HTTP authentication failed")
	ErrForbidden      = errors.New("catalog HTTP operation is forbidden")
	ErrInvalid        = errors.New("catalog HTTP input is invalid")
	ErrNotFound       = errors.New("catalog HTTP resource was not found")
	ErrConflict       = errors.New("catalog HTTP state conflicts")
	ErrUnavailable    = errors.New("catalog HTTP authority is unavailable")
)

type Permission string

const (
	PermissionRead       Permission = "catalog.read"
	PermissionInstall    Permission = "catalog.install"
	PermissionUpgrade    Permission = "catalog.upgrade"
	PermissionDisable    Permission = "catalog.disable"
	PermissionEnable     Permission = "catalog.enable"
	PermissionRestore    Permission = "catalog.restore"
	PermissionRetire     Permission = "catalog.retire"
	PermissionPurge      Permission = "catalog.purge"
	PermissionDemoManage Permission = "catalog.demo.manage"
	PermissionMenuManage Permission = "catalog.menu.manage"
)

type AccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          Permission
	Operation           string
	PurposeTarget       string
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (Access, error)
}

type Access struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	permission                                             Permission
	operation                                              string
	stepUpAuthenticatedAt                                  time.Time
	authorizationPermissionID                              uuid.UUID
	authorizationEvidenceSHA256, stepUpPurposeSHA256       [sha256.Size]byte
}

func NewAccess(principalID, tenantID uuid.UUID, tokenVersion, authorizationVersion,
	authorizationClock int64, stepUpAuthenticatedAt time.Time,
	permission Permission, operation string,
) (Access, error) {
	value := Access{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission, operation: operation,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC().Truncate(time.Second)}
	if !stepUpAuthenticatedAt.IsZero() && stepUpAuthenticatedAt.Nanosecond() != 0 ||
		!value.ValidFor(permission, operation) {
		return Access{}, ErrInvalid
	}
	return value, nil
}

func NewAccessWithAuthorizationEvidence(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
	authorizationPermissionID uuid.UUID,
	authorizationEvidenceSHA256, stepUpPurposeSHA256 [sha256.Size]byte,
	permission Permission,
	operation string,
) (Access, error) {
	access, err := NewAccess(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, stepUpAuthenticatedAt, permission, operation,
	)
	if err != nil || authorizationPermissionID == uuid.Nil ||
		authorizationEvidenceSHA256 == [sha256.Size]byte{} ||
		(!stepUpAuthenticatedAt.IsZero() &&
			stepUpPurposeSHA256 == [sha256.Size]byte{}) ||
		(stepUpAuthenticatedAt.IsZero() &&
			stepUpPurposeSHA256 != [sha256.Size]byte{}) {
		return Access{}, ErrInvalid
	}
	access.authorizationPermissionID = authorizationPermissionID
	access.authorizationEvidenceSHA256 = authorizationEvidenceSHA256
	access.stepUpPurposeSHA256 = stepUpPurposeSHA256
	return access, nil
}

func (access Access) ValidFor(permission Permission, operation string) bool {
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.permission == permission &&
		access.operation == operation && validPermissionOperation(permission, operation)
}

func (access Access) PrincipalID() uuid.UUID      { return access.principalID }
func (access Access) TenantID() uuid.UUID         { return access.tenantID }
func (access Access) TokenVersion() int64         { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) AuthorizationClock() int64   { return access.authorizationClock }
func (access Access) StepUpAt() time.Time         { return access.stepUpAuthenticatedAt }
func (access Access) Permission() Permission      { return access.permission }
func (access Access) Operation() string           { return access.operation }
func (access Access) AuthorizationPermissionID() uuid.UUID {
	return access.authorizationPermissionID
}
func (access Access) AuthorizationEvidenceSHA256() [sha256.Size]byte {
	return access.authorizationEvidenceSHA256
}
func (access Access) StepUpPurposeSHA256() [sha256.Size]byte {
	return access.stepUpPurposeSHA256
}

func validPermissionOperation(permission Permission, operation string) bool {
	expected, found := operationPermissions[operation]
	return found && permission == expected
}

var operationPermissions = map[string]Permission{
	"catalog.package.list": PermissionRead, "catalog.package.get": PermissionRead,
	"catalog.package_version.list": PermissionRead, "catalog.package_version.get": PermissionRead,
	"catalog.package_version_dependency.list": PermissionRead,
	"catalog.installation.list":               PermissionRead, "catalog.installation.get": PermissionRead,
	"catalog.installation.install": PermissionInstall, "catalog.installation.upgrade": PermissionUpgrade,
	"catalog.installation.disable": PermissionDisable, "catalog.installation.enable": PermissionEnable,
	"catalog.installation.restore": PermissionRestore, "catalog.installation.retire": PermissionRetire,
	"catalog.installation.purge":         PermissionPurge,
	"catalog.installation.purge.approve": PermissionPurge,
	"catalog.operation.get":              PermissionRead,
	"catalog.operation.cancel":           PermissionInstall, "catalog.operation.retry": PermissionInstall,
	"catalog.installation.demo.get":          PermissionRead,
	"catalog.installation.demo.removal_plan": PermissionDemoManage,
	"catalog.installation.demo.install":      PermissionDemoManage,
	"catalog.installation.demo.remove":       PermissionDemoManage,
	"catalog.menu.get":                       PermissionRead, "catalog.menu.update": PermissionMenuManage,
}
