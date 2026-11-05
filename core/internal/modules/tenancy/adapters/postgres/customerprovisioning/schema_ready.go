package customerprovisioning

import (
	"context"
	"fmt"
	"regexp"
)

var tenantSchemaName = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)

const tenantSchemaIdentitySQL = `SELECT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_class AS relation
    INNER JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = $1
      AND relation.relname = 'tenancy_schema_identity'
      AND relation.relkind = 'r'
)`

func (status *SchemaStatus) tenantSchemaProvisioned(
	ctx context.Context, schemaName string,
) (bool, error) {
	if !tenantSchemaName.MatchString(schemaName) {
		return false, fmt.Errorf("tenant schema name is invalid")
	}
	var exists bool
	err := status.query(ctx, tenantSchemaIdentitySQL, schemaName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("read tenant schema identity: %w", err)
	}
	return exists, nil
}
