package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_tenant_requirement_catalog.sql
var getTenantRequirementCatalogSQL string

//go:embed sql/list_tenant_requirements.sql
var listTenantRequirementsSQL string

//go:embed sql/list_tenant_requirement_targets.sql
var listTenantRequirementTargetsSQL string

func GetTenantRequirementCatalog(
	ctx context.Context,
	database goosedb.DBTxConn,
	runID string,
) (TenantRequirementCatalog, error) {
	var catalog TenantRequirementCatalog
	err := database.QueryRowContext(ctx, getTenantRequirementCatalogSQL, runID).Scan(
		&catalog.ContractVersion, &catalog.MigrationRunID, &catalog.ReleaseID,
		&catalog.EnvironmentID, &catalog.DatabaseOrigin, &catalog.MigrationHash,
		&catalog.CatalogHash, &catalog.RequirementCount, &catalog.TargetCount,
		&catalog.Canonical,
		&catalog.RecordedAt,
	)
	if err != nil {
		return TenantRequirementCatalog{}, fmt.Errorf("get tenant requirement catalog: %w", err)
	}
	rows, err := database.QueryContext(ctx, listTenantRequirementsSQL, runID)
	if err != nil {
		return TenantRequirementCatalog{}, fmt.Errorf("list tenant requirements: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var requirement TenantMigrationRequirement
		if err := rows.Scan(
			&requirement.Ordinal, &requirement.Owner, &requirement.Version,
			&requirement.Origin, &requirement.SelectedPath, &requirement.SelectedHash,
		); err != nil {
			return TenantRequirementCatalog{}, fmt.Errorf("scan tenant requirement: %w", err)
		}
		catalog.Requirements = append(catalog.Requirements, requirement)
	}
	if err := rows.Err(); err != nil {
		return TenantRequirementCatalog{}, fmt.Errorf("iterate tenant requirements: %w", err)
	}
	targetRows, err := database.QueryContext(ctx, listTenantRequirementTargetsSQL, runID)
	if err != nil {
		return TenantRequirementCatalog{}, fmt.Errorf("list tenant requirement targets: %w", err)
	}
	defer targetRows.Close()
	for targetRows.Next() {
		var target TenantRequirementTarget
		if err := targetRows.Scan(
			&target.Ordinal, &target.TenantID, &target.SchemaName,
		); err != nil {
			return TenantRequirementCatalog{}, fmt.Errorf("scan tenant requirement target: %w", err)
		}
		catalog.Targets = append(catalog.Targets, target)
	}
	if err := targetRows.Err(); err != nil {
		return TenantRequirementCatalog{}, fmt.Errorf("iterate tenant requirement targets: %w", err)
	}
	return catalog, nil
}
