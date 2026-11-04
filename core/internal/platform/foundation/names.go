package foundation

import (
	"encoding/hex"
	"errors"
	"regexp"
)

const (
	RoleOwner       = "harpy_owner"
	RoleGoose       = "harpy_goose"
	RoleAPI         = "harpy_api"
	RoleWorker      = "harpy_worker"
	PlatformSchema  = "platform"
	VectorExtension = "vector"
)

var (
	databaseNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	tenantSchemaPattern = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)
	tenantRolePattern   = regexp.MustCompile(`^harpy_td_[0-9a-f]{32}$`)
)

func ValidDatabaseName(name string) bool {
	switch name {
	case "postgres", "template0", "template1":
		return false
	default:
		return databaseNamePattern.MatchString(name)
	}
}

func TenantSchemaName(tenantID [16]byte) (string, error) {
	if tenantID == ([16]byte{}) {
		return "", errors.New("tenant ID must be non-zero")
	}
	return "tenant_" + hex.EncodeToString(tenantID[:]), nil
}

func TenantRoleName(tenantID [16]byte) (string, error) {
	if tenantID == ([16]byte{}) {
		return "", errors.New("tenant ID must be non-zero")
	}
	return "harpy_td_" + hex.EncodeToString(tenantID[:]), nil
}

func AllowedSchema(name string) bool {
	switch name {
	case "public", PlatformSchema:
		return true
	default:
		return tenantSchemaPattern.MatchString(name)
	}
}

func AllowedExtension(name string) bool {
	return name == "plpgsql" || name == VectorExtension
}

func allowedRole(name string) bool {
	switch name {
	case RoleOwner, RoleGoose, RoleAPI, RoleWorker:
		return true
	default:
		return tenantRolePattern.MatchString(name)
	}
}
