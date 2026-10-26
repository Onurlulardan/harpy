package postgres

import (
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
)

var schemaExecutorSchemaNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func schemaExecutorSearchPath(schemaName string) (string, error) {
	if len(schemaName) < 1 || len(schemaName) > 63 ||
		!schemaExecutorSchemaNamePattern.MatchString(schemaName) {
		return "", fmt.Errorf("tenant schema name is invalid")
	}
	return "pg_catalog," + pgx.Identifier{schemaName}.Sanitize(), nil
}
