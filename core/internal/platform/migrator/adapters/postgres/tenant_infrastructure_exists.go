package postgres

import (
	"context"
	"fmt"
)

func TenantInfrastructureExists(
	ctx context.Context,
	database tenantRowQuerier,
) (bool, error) {
	var registry, checkpoints, attempts bool
	if err := database.QueryRowContext(ctx, tenantInfrastructureExistsSQL).Scan(
		&registry, &checkpoints, &attempts,
	); err != nil {
		return false, fmt.Errorf("check tenant migration infrastructure: %w", err)
	}
	return registry && checkpoints && attempts, nil
}
