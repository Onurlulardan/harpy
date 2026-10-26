package postgres

import (
	"context"
	"fmt"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store struct{ pool *pgxpool.Pool }

func Bind(pool *platformpostgres.Pool) (tenantmigrationexecutor.Store, error) {
	return platformpostgres.BindTenantMigrationExecutorStore(
		pool, func(database *pgxpool.Pool) tenantmigrationexecutor.Store {
			return &store{pool: database}
		},
	)
}

func seconds(duration time.Duration) int { return int(duration / time.Second) }

func (storage *store) usable(ctx context.Context) error {
	if storage == nil || storage.pool == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("tenant migration postgres store is unavailable")
	}
	return nil
}

func (storage *store) RegisterReady(
	ctx context.Context, profileID uuid.UUID,
	release tenantmigrationexecutor.Release, ttl time.Duration,
) (tenantmigrationexecutor.ReadyReceipt, error) {
	if err := storage.usable(ctx); err != nil {
		return tenantmigrationexecutor.ReadyReceipt{}, err
	}
	row := storage.pool.QueryRow(ctx, registerReadySQL, profileID, release.ID,
		release.DeploymentMode, release.RuntimeGeneration,
		release.MigrationManifestSHA256, seconds(ttl))
	var receipt tenantmigrationexecutor.ReadyReceipt
	err := row.Scan(&receipt.ProfileInstanceID, &receipt.ReleaseID,
		&receipt.EnvironmentID, &receipt.DeploymentMode,
		&receipt.RuntimeGeneration, &receipt.DatabaseGeneration,
		&receipt.MigrationManifestSHA256, &receipt.ReadyUntil)
	return receipt, err
}

func (storage *store) HeartbeatReady(
	ctx context.Context, profileID uuid.UUID, ttl time.Duration,
) (time.Time, error) {
	if err := storage.usable(ctx); err != nil {
		return time.Time{}, err
	}
	var until time.Time
	err := storage.pool.QueryRow(ctx, heartbeatReadySQL, profileID, seconds(ttl)).Scan(&until)
	return until, err
}

func (storage *store) RetireReady(ctx context.Context, profileID uuid.UUID) error {
	if err := storage.usable(ctx); err != nil {
		return err
	}
	var retired bool
	return storage.pool.QueryRow(ctx, retireReadySQL, profileID).Scan(&retired)
}

var _ tenantmigrationexecutor.Store = (*store)(nil)
