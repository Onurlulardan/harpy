package ddl

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

// CreateTenantSchema renders the dynamic namespace and activation ACL owned by
// the new-tenant migrator.
func CreateTenantSchema(schemaName string) (string, error) {
	token, err := sqlidentifier.MintStatic(schemaName)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("CREATE SCHEMA ")
	if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
		return "", err
	}
	builder.WriteString(" AUTHORIZATION harpy_owner")
	builder.WriteString("; REVOKE CREATE ON SCHEMA ")
	if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
		return "", err
	}
	builder.WriteString(" FROM PUBLIC, harpy_runtime, harpy_v2_runtime")
	builder.WriteString("; GRANT USAGE ON SCHEMA ")
	if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
		return "", err
	}
	builder.WriteString(" TO harpy_runtime")
	return builder.String(), nil
}
