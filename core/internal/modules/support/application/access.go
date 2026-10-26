package application

import (
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	supportports "github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
	"github.com/google/uuid"
)

type Access = supportports.Access
type AccessScope = supportports.AccessScope

const (
	ScopeReporter = supportports.ScopeReporter
	ScopeTenant   = supportports.ScopeTenant
	ScopePlatform = supportports.ScopePlatform

	PermissionBugCreate             = supportports.PermissionBugCreate
	PermissionBugReadOwn            = supportports.PermissionBugReadOwn
	PermissionBugRead               = supportports.PermissionBugRead
	PermissionBugManage             = supportports.PermissionBugManage
	PermissionPlatformBugRead       = supportports.PermissionPlatformBugRead
	PermissionPlatformBugManage     = supportports.PermissionPlatformBugManage
	PermissionPlatformContactRead   = supportports.PermissionPlatformContactRead
	PermissionPlatformContactManage = supportports.PermissionPlatformContactManage
)

func NewAccess(actorID, tenantID uuid.UUID, tokenVersion, authorizationVersion,
	authorityClock int64, scope AccessScope, permission, operation,
	reporterName, reporterEmail, tenantSlug, tenantName string,
) (Access, error) {
	return supportports.NewAccess(domain.ID(actorID), domain.ID(tenantID), tokenVersion,
		authorizationVersion, authorityClock, scope, permission, operation,
		reporterName, reporterEmail, tenantSlug, tenantName)
}

func domainText(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && len(value) >= minimum && len([]byte(value)) <= maximum &&
		!strings.ContainsRune(value, 0)
}
