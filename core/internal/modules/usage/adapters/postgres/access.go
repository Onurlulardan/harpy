package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func accessArguments(access application.Access) []any {
	return []any{access.ActorID(), access.SubjectTenantID(), nullableUUID(access.TargetTenantID()),
		access.EnvironmentID(), access.TokenVersion(), access.AuthorizationVersion(),
		access.AuthorityClock(), string(access.Scope()), string(access.Mode()),
		string(access.Purpose()), access.ControlPrimary(), access.Permission(), access.Operation()}
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func allows(access application.Access, tenantOperation, platformOperation string) bool {
	expected := tenantOperation
	if access.Scope() == domain.PlatformScope {
		expected = platformOperation
	}
	return access.Allows(access.Scope(), expected)
}

func withArguments(access application.Access, trailing ...any) []any {
	return append(accessArguments(access), trailing...)
}
