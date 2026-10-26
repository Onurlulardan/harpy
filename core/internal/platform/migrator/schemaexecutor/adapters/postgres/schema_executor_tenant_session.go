package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaExecutorTenantSession struct {
	connection    *pgxpool.Conn
	tenantID      string
	schemaName    string
	trigramSchema string
	cleanup       time.Duration
}

func (store *schemaExecutorStore) openTenantSession(
	ctx context.Context,
	lease schemaexecutor.Lease,
) (*schemaExecutorTenantSession, error) {
	connection, err := store.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire schema executor connection: %w", err)
	}
	session := &schemaExecutorTenantSession{
		connection: connection, tenantID: lease.TenantID(), cleanup: store.cleanupTimeout,
	}
	preflightCtx, cancelPreflight := schemaExecutorOperationContext(ctx)
	schemaName, err := schemaExecutorResolveTenant(
		preflightCtx, connection, lease.TenantID(),
	)
	cancelPreflight()
	if err != nil {
		connection.Release()
		return nil, err
	}
	session.schemaName = schemaName
	var lockedTenant string
	if err := connection.QueryRow(
		ctx, schemaExecutorAcquireTenantLockSQL, lease.TenantID(),
	).Scan(&lockedTenant); err != nil {
		connection.Release()
		return nil, fmt.Errorf("acquire tenant schema lock: %w", err)
	}
	if lockedTenant != lease.TenantID() {
		connection.Release()
		return nil, schemaExecutorBlocked("schema_tenant_lock_mismatch", nil)
	}
	verifyCtx, cancelVerify := schemaExecutorOperationContext(ctx)
	err = session.verifyTenant(verifyCtx)
	cancelVerify()
	if err != nil {
		session.close(store.cleanupTimeout)
		return nil, err
	}
	return session, nil
}

func schemaExecutorResolveTenant(
	ctx context.Context,
	querier interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	},
	tenantID string,
) (string, error) {
	var schemaName, status string
	var owner *string
	var notDeleted bool
	err := querier.QueryRow(ctx, schemaExecutorResolveTenantSQL, tenantID).Scan(
		&schemaName, &status, &notDeleted, &owner,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", schemaExecutorBlocked("schema_tenant_missing", nil)
	}
	if err != nil {
		return "", fmt.Errorf("resolve schema executor tenant: %w", err)
	}
	if _, err := schemaExecutorSearchPath(schemaName); err != nil || !notDeleted ||
		(status != "ACTIVE" && status != "TRIAL") || owner == nil ||
		*owner != "harpy_owner" {
		return "", schemaExecutorBlocked("schema_tenant_invalid", err)
	}
	return schemaName, nil
}

func (session *schemaExecutorTenantSession) verifyTenant(ctx context.Context) error {
	transaction, err := session.connection.Begin(ctx)
	if err != nil {
		return err
	}
	defer session.rollback(transaction)
	schemaName, err := schemaExecutorResolveTenant(ctx, transaction, session.tenantID)
	if err != nil {
		return err
	}
	if schemaName != session.schemaName {
		return schemaExecutorBlocked("schema_tenant_stale", nil)
	}
	searchPath, err := schemaExecutorSearchPath(session.schemaName)
	if err != nil {
		return schemaExecutorBlocked("schema_tenant_invalid", err)
	}
	var appliedPath, scopedTenant string
	if err := transaction.QueryRow(
		ctx, schemaExecutorSetSearchPathSQL, searchPath, session.tenantID,
	).Scan(&appliedPath, &scopedTenant); err != nil {
		return err
	}
	if appliedPath != searchPath || scopedTenant != session.tenantID {
		return schemaExecutorBlocked("schema_tenant_search_path_invalid", nil)
	}
	var tenantID, sentinelSchema string
	err = transaction.QueryRow(
		ctx, schemaExecutorVerifySentinelSQL, session.tenantID, session.schemaName,
	).Scan(&tenantID, &sentinelSchema)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorBlocked("schema_tenant_sentinel_invalid", nil)
	}
	if err != nil {
		return err
	}
	if tenantID != session.tenantID || sentinelSchema != session.schemaName {
		return schemaExecutorBlocked("schema_tenant_sentinel_invalid", nil)
	}
	var trigramSchema string
	err = transaction.QueryRow(ctx, schemaExecutorResolveTrigramSchemaSQL).Scan(&trigramSchema)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorBlocked("schema_trigram_registry_invalid", nil)
	}
	if err != nil {
		return err
	}
	if _, err := schemaExecutorSearchPath(trigramSchema); err != nil {
		return schemaExecutorBlocked("schema_trigram_registry_invalid", err)
	}
	session.trigramSchema = trigramSchema
	return transaction.Commit(ctx)
}

func (session *schemaExecutorTenantSession) rollback(transaction pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), session.cleanup)
	defer cancel()
	_ = transaction.Rollback(ctx)
}

func schemaExecutorBlocked(code string, cause error) error {
	return schemaexecutor.NewFault(schemaexecutor.FaultBlocked, code, cause)
}
