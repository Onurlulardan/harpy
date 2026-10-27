package postgres

import (
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

var schemaNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func tenantSearchPath(schemaName string) (string, error) {
	if len(schemaName) < 1 || len(schemaName) > 63 ||
		!schemaNamePattern.MatchString(schemaName) || strings.ContainsRune(schemaName, 0) {
		return "", ErrTenantSchemaInvalid
	}
	return "pg_catalog," + pgx.Identifier{schemaName}.Sanitize(), nil
}
