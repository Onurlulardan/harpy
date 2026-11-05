package operator

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/provisionlaunch"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

var tenantSchemaPattern = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)

const tenantIdentitySQL = `SELECT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_namespace AS namespace
    INNER JOIN pg_catalog.pg_class AS relation
      ON relation.relnamespace = namespace.oid
    WHERE namespace.nspname = $1
      AND relation.relname = 'tenancy_schema_identity'
      AND relation.relkind = 'r'
)`

const deliveryRecordedSQL = `SELECT operation_id
FROM platform.tenancy_installation_delivery_read_v1($1)`

type composeTenantProvisioner struct {
	query func(context.Context, string, ...any) platformpostgres.Row
}

func newComposeTenantProvisioner(
	pool *platformpostgres.Pool,
) (composeTenantProvisioner, error) {
	return platformpostgres.BindTenancyOperator(pool, func(
		database platformpostgres.TenancyOperatorDatabase,
	) (composeTenantProvisioner, error) {
		return composeTenantProvisioner{query: database.QueryRow}, nil
	})
}

func (provisioner composeTenantProvisioner) Provision(
	ctx context.Context, tenantID uuid.UUID,
) error {
	return provisioner.ensure(ctx, tenantID)
}

func (provisioner composeTenantProvisioner) SchemaReady(
	ctx context.Context, tenantID uuid.UUID,
) error {
	return provisioner.ensure(ctx, tenantID)
}

func (provisioner composeTenantProvisioner) TenantActivated(
	ctx context.Context, operationID uuid.UUID,
) (bool, error) {
	if provisioner.query == nil || ctx == nil || ctx.Err() != nil || operationID == uuid.Nil {
		return false, fmt.Errorf("%w: tenant schema name is invalid: <empty>",
			environmentoperator.ErrInvalid)
	}
	var recorded uuid.UUID
	err := provisioner.query(ctx, deliveryRecordedSQL, operationID).Scan(&recorded)
	if platformpostgres.RowNotFound(err) || recorded == uuid.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return recorded == operationID, nil
}

func (provisioner composeTenantProvisioner) ensure(
	ctx context.Context, tenantID uuid.UUID,
) error {
	schema, err := canonicalTenantSchema(tenantID)
	if err != nil {
		return err
	}
	if provisioner.query == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("%w: tenant schema name is invalid: %s",
			environmentoperator.ErrInvalid, schema)
	}
	var ready bool
	if err = provisioner.query(ctx, tenantIdentitySQL, schema).Scan(&ready); err != nil {
		return err
	}
	if ready {
		return nil
	}
	if err = provisionlaunch.Run(ctx, tenantID.String()); err != nil &&
		!errors.Is(err, provisionlaunch.ErrPending) {
		return err
	}
	return environmentoperator.ErrTenantProvisioningPending
}

func canonicalTenantSchema(tenantID uuid.UUID) (string, error) {
	if tenantID == uuid.Nil {
		return "", fmt.Errorf("%w: tenant schema name is invalid: <empty>",
			environmentoperator.ErrInvalid)
	}
	name := "tenant_" + hex.EncodeToString(tenantID[:])
	if !tenantSchemaPattern.MatchString(name) {
		return "", fmt.Errorf("%w: tenant schema name is invalid: %s",
			environmentoperator.ErrInvalid, name)
	}
	return name, nil
}
