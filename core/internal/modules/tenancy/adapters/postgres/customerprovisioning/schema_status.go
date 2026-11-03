package customerprovisioning

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const markProvisioningSQL = `INSERT INTO platform.tenants (
    id, name, slug, status, schema_name, settings_revision, lifecycle_revision,
    created_at, updated_at
) VALUES (
    $1, $2, $3, 'PROVISIONING', $4, 1, 1,
    pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp()
)
ON CONFLICT (id) DO UPDATE
SET status = 'PROVISIONING', name = EXCLUDED.name, slug = EXCLUDED.slug,
    schema_name = EXCLUDED.schema_name, updated_at = pg_catalog.clock_timestamp()
WHERE platform.tenants.status = 'PROVISIONING'
RETURNING id`

const markActiveSQL = `UPDATE platform.tenants
SET status = 'ACTIVE', updated_at = pg_catalog.clock_timestamp()
WHERE id = $1 AND status = 'PROVISIONING'
RETURNING id`

type SchemaStatus struct {
	query QueryRow
	run   func(context.Context, string) error
}

func NewSchemaStatus(query QueryRow, run func(context.Context, string) error) (*SchemaStatus, error) {
	if query == nil || run == nil {
		return nil, fmt.Errorf("tenant schema status writer is incomplete")
	}
	return &SchemaStatus{query: query, run: run}, nil
}

func (status *SchemaStatus) Begin(
	ctx context.Context, tenantID uuid.UUID, schemaName, name, slug string,
) error {
	var id uuid.UUID
	err := status.query(ctx, markProvisioningSQL, tenantID, name, slug, schemaName).Scan(&id)
	if err != nil || id != tenantID {
		return fmt.Errorf("mark tenant provisioning: %w", err)
	}
	if err = status.run(ctx, tenantID.String()); err != nil {
		return err
	}
	return nil
}

func (status *SchemaStatus) Activate(ctx context.Context, tenantID uuid.UUID) error {
	var id uuid.UUID
	err := status.query(ctx, markActiveSQL, tenantID).Scan(&id)
	if err != nil || id != tenantID {
		return fmt.Errorf("mark tenant active: %w", err)
	}
	return nil
}
