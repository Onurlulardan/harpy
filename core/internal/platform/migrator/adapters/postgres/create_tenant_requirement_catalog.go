package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_tenant_requirement_catalog.sql
var createTenantRequirementCatalogSQL string

//go:embed sql/create_tenant_requirements.sql
var createTenantRequirementsSQL string

//go:embed sql/create_tenant_requirement_targets.sql
var createTenantRequirementTargetsSQL string

func CreateTenantRequirementCatalog(
	ctx context.Context,
	database goosedb.DBTxConn,
	catalog TenantRequirementCatalog,
) error {
	if catalog.RequirementCount != len(catalog.Requirements) ||
		catalog.TargetCount != len(catalog.Targets) {
		return fmt.Errorf("tenant requirement catalog count differs from projections")
	}
	result, err := database.ExecContext(ctx, createTenantRequirementCatalogSQL,
		catalog.MigrationRunID, catalog.ContractVersion, catalog.ReleaseID,
		catalog.EnvironmentID, catalog.DatabaseOrigin, catalog.MigrationHash,
		catalog.CatalogHash, catalog.RequirementCount, catalog.TargetCount,
		catalog.Canonical,
	)
	if err != nil {
		return fmt.Errorf("create tenant requirement catalog: %w", err)
	}
	created, err := result.RowsAffected()
	if err != nil || (created != 0 && created != 1) {
		return fmt.Errorf("create tenant requirement catalog changed %d rows", created)
	}
	if created == 0 {
		return nil
	}
	document, err := json.Marshal(catalog.Requirements)
	if err != nil {
		return fmt.Errorf("encode tenant requirements: %w", err)
	}
	result, err = database.ExecContext(
		ctx, createTenantRequirementsSQL, catalog.MigrationRunID, document,
	)
	if err != nil {
		return fmt.Errorf("create tenant requirements: %w", err)
	}
	requirements, err := result.RowsAffected()
	if err != nil || requirements != int64(len(catalog.Requirements)) {
		return fmt.Errorf("create tenant requirements changed %d rows", requirements)
	}
	document, err = json.Marshal(catalog.Targets)
	if err != nil {
		return fmt.Errorf("encode tenant requirement targets: %w", err)
	}
	result, err = database.ExecContext(
		ctx, createTenantRequirementTargetsSQL, catalog.MigrationRunID, document,
	)
	if err != nil {
		return fmt.Errorf("create tenant requirement targets: %w", err)
	}
	targets, err := result.RowsAffected()
	if err != nil || targets != int64(len(catalog.Targets)) {
		return fmt.Errorf("create tenant requirement targets changed %d rows", targets)
	}
	return nil
}
