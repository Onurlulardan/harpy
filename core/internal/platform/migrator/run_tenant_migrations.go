package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (runner *Runner) runTenantMigrations(
	ctx context.Context,
	connection *sql.Conn,
	runID string,
) ([]int64, error) {
	if len(runner.tenantSelected) == 0 {
		return nil, nil
	}
	ready, err := tenantInfrastructureRequired(ctx, connection)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, fmt.Errorf("tenant migration infrastructure is absent")
	}
	targets, err := listTenantTargets(ctx, connection)
	if err != nil {
		return nil, err
	}
	applied := make(map[int64]bool, len(runner.tenantSelected))
	var committedErr error
	for _, target := range targets {
		versions, err := runner.executeTenantTarget(ctx, target, runID)
		if err != nil && !errors.Is(err, ErrCommitted) {
			return nil, fmt.Errorf("migrate tenant %s: %w", target.TenantID, err)
		}
		committedErr = errors.Join(committedErr, err)
		for _, version := range versions {
			applied[version] = true
		}
	}
	versions := make([]int64, 0, len(applied))
	for _, selected := range runner.tenantSelected {
		if applied[selected.Version] {
			versions = append(versions, selected.Version)
		}
	}
	return versions, committedErr
}
