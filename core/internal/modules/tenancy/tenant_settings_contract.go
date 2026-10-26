package tenancy

import (
	"context"
	"reflect"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/google/uuid"
)

type TenantSettingsPermission string

const (
	TenantSettingsPermissionRead   TenantSettingsPermission = "tenancy.settings.read"
	TenantSettingsPermissionManage TenantSettingsPermission = "tenancy.settings.manage"

	TenantSettingsOperationGet       = "tenancy.settings.get"
	TenantSettingsOperationUpdate    = "tenancy.settings.update"
	TenantSettingsOperationTestEmail = "tenancy.settings.test_email"
)

type TenantSettingsAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          TenantSettingsPermission
	Operation           string
}

type TenantSettingsAccessVerifier interface {
	Verify(context.Context, TenantSettingsAccessRequest) (TenantSettingsAccess, error)
}

type TenantSettingsAccess struct {
	principalID, tenantID                                  uuid.UUID
	permission                                             TenantSettingsPermission
	operation                                              string
	tokenVersion, authorizationVersion, authorizationClock int64
}

func NewTenantSettingsAccess(principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission TenantSettingsPermission, operation string,
) (TenantSettingsAccess, error) {
	value := TenantSettingsAccess{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission, operation: operation}
	if !value.validFor(permission, operation) {
		return TenantSettingsAccess{}, settings.ErrInvalid
	}
	return value, nil
}

func (access TenantSettingsAccess) validFor(permission TenantSettingsPermission,
	operation string,
) bool {
	validPair := permission == TenantSettingsPermissionRead &&
		operation == TenantSettingsOperationGet ||
		permission == TenantSettingsPermissionManage &&
			(operation == TenantSettingsOperationUpdate || operation == TenantSettingsOperationTestEmail)
	return access.principalID != uuid.Nil && access.tenantID != uuid.Nil &&
		access.tokenVersion > 0 && access.authorizationVersion > 0 &&
		access.authorizationClock > 0 && access.permission == permission &&
		access.operation == operation && validPair
}

func tenantSettingsPortMissing(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

type TenantSettingsProjectionReader = settings.ProjectionReader
type TenantSettingsMail = settings.MailSettings
type TenantSettingsClock = settings.Clock
type TenantSettingsProjection = settings.Projection
type TenantSettingsMailMutation = settings.MailMutation
type TenantSettingsTestMailRequest = settings.TestMailRequest
type TenantSettingsTestEmailResult = settings.TestEmailResult

var (
	ErrTenantSettingsAuthentication = settings.ErrAuthentication
	ErrTenantSettingsForbidden      = settings.ErrForbidden
	ErrTenantSettingsInvalid        = settings.ErrInvalid
	ErrTenantSettingsPrecondition   = settings.ErrPrecondition
	ErrTenantSettingsNotFound       = settings.ErrNotFound
	ErrTenantSettingsConflict       = settings.ErrConflict
	ErrTenantSettingsUnavailable    = settings.ErrUnavailable
)
