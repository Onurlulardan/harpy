package postgres

import (
	"context"
	"fmt"
)

func ListTenantMigrationTargets(
	ctx context.Context,
	database tenantRowsQuerier,
) ([]TenantMigrationTarget, error) {
	rows, err := database.QueryContext(ctx, listTenantMigrationTargetsSQL)
	if err != nil {
		return nil, fmt.Errorf("list tenant migration targets: %w", err)
	}
	defer rows.Close()
	var targets []TenantMigrationTarget
	for rows.Next() {
		var target TenantMigrationTarget
		if err := rows.Scan(&target.TenantID, &target.SchemaName); err != nil {
			return nil, fmt.Errorf("scan tenant migration target: %w", err)
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant migration targets: %w", err)
	}
	return targets, nil
}
