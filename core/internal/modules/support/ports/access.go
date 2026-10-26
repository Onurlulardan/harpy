package ports

import (
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type AccessScope string

const (
	ScopeReporter AccessScope = "reporter"
	ScopeTenant   AccessScope = "tenant"
	ScopePlatform AccessScope = "platform"
)

const (
	PermissionBugCreate             = "bug_reports.create"
	PermissionBugReadOwn            = "bug_reports.read_own"
	PermissionBugRead               = "bug_reports.read"
	PermissionBugManage             = "bug_reports.manage"
	PermissionPlatformBugRead       = "platform.bug_reports.read"
	PermissionPlatformBugManage     = "platform.bug_reports.manage"
	PermissionPlatformContactRead   = "platform.contact_messages.read"
	PermissionPlatformContactManage = "platform.contact_messages.manage"
)

type Access struct {
	actorID, tenantID                                   domain.ID
	tokenVersion, authorizationVersion, authorityClock  int64
	scope                                               AccessScope
	permission, operation                               string
	reporterName, reporterEmail, tenantSlug, tenantName string
}

func NewAccess(actorID, tenantID domain.ID, tokenVersion, authorizationVersion,
	authorityClock int64, scope AccessScope, permission, operation,
	reporterName, reporterEmail, tenantSlug, tenantName string,
) (Access, error) {
	access := Access{actorID: actorID, tenantID: tenantID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, authorityClock: authorityClock,
		scope: scope, permission: permission, operation: operation,
		reporterName:  strings.TrimSpace(reporterName),
		reporterEmail: strings.TrimSpace(strings.ToLower(reporterEmail)),
		tenantSlug:    strings.TrimSpace(tenantSlug), tenantName: strings.TrimSpace(tenantName)}
	if !access.valid() {
		return Access{}, ErrInvalid
	}
	return access, nil
}

func (access Access) valid() bool {
	if access.actorID == domain.NilID || access.tokenVersion < 1 ||
		access.authorizationVersion < 1 || access.authorityClock < 1 ||
		access.permission == "" || access.operation == "" ||
		!validText(access.reporterName, 0, 200) || !validText(access.reporterEmail, 0, 255) {
		return false
	}
	switch access.scope {
	case ScopeReporter, ScopeTenant:
		return access.tenantID != domain.NilID && validText(access.tenantSlug, 1, 100) &&
			validText(access.tenantName, 1, 200)
	case ScopePlatform:
		return access.tenantID == domain.NilID
	default:
		return false
	}
}

func (access Access) Allows(scope AccessScope, permission, operation string) bool {
	return access.valid() && access.scope == scope && access.permission == permission &&
		access.operation == operation
}

func (access Access) ActorID() domain.ID          { return access.actorID }
func (access Access) TenantID() domain.ID         { return access.tenantID }
func (access Access) ReporterName() string        { return access.reporterName }
func (access Access) ReporterEmail() string       { return access.reporterEmail }
func (access Access) TenantSlug() string          { return access.tenantSlug }
func (access Access) TenantName() string          { return access.tenantName }
func (access Access) Scope() AccessScope          { return access.scope }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) AuthorityClock() int64       { return access.authorityClock }

func validText(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && len(value) >= minimum && len([]byte(value)) <= maximum &&
		!strings.ContainsRune(value, 0)
}
