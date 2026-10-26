package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/tenantmigrationexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/google/uuid"
)

type NewID func() (uuid.UUID, error)

func EnsureSchema(
	ctx context.Context, database *sql.DB, newID NewID,
	claim tenantmigrationexecutor.Claim,
) error {
	if ctx == nil || database == nil || newID == nil {
		return fmt.Errorf("tenant schema executor is unavailable")
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	tenantID, schemaName, namespaceExists, receiptExists, err :=
		inspectSchema(ctx, transaction, claim)
	if err != nil {
		return err
	}
	if tenantID != claim.TenantID.String() || schemaName != claim.SchemaName {
		return fmt.Errorf("tenant schema target differs from claim")
	}
	if namespaceExists != receiptExists {
		return tenantmigrationexecutor.ErrCleanupRequired
	}
	if namespaceExists {
		return transaction.Commit()
	}
	command, err := ddl.CreateTenantSchema(schemaName)
	if err != nil {
		return err
	}
	if _, err := transaction.ExecContext(ctx, command); err != nil {
		return err
	}
	if err := recordSchema(ctx, transaction, newID, claim); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return reconcileSchema(context.WithoutCancel(ctx), database, claim, err)
	}
	return nil
}

func inspectSchema(
	ctx context.Context, query interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}, claim tenantmigrationexecutor.Claim,
) (string, string, bool, bool, error) {
	var tenantID, schemaName string
	var namespaceExists, receiptExists bool
	err := query.QueryRowContext(ctx, inspectSchemaSQL, claim.JobID, claim.Attempt,
		claim.ProfileInstanceID, claim.LeaseID).Scan(
		&tenantID, &schemaName, &namespaceExists, &receiptExists)
	return tenantID, schemaName, namespaceExists, receiptExists, err
}

func recordSchema(
	ctx context.Context, transaction *sql.Tx, newID NewID,
	claim tenantmigrationexecutor.Claim,
) error {
	receiptID, err := newID()
	if err != nil {
		return err
	}
	evidence := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%s:%s", claim.TenantID, claim.SchemaName, claim.MigrationRunID,
	)))
	digest := hex.EncodeToString(evidence[:])
	var recorded bool
	err = transaction.QueryRowContext(ctx, recordSchemaSQL, claim.JobID,
		claim.Attempt, claim.ProfileInstanceID, claim.LeaseID, receiptID,
		digest, digest).Scan(&recorded)
	if err != nil || !recorded {
		return errors.Join(err, fmt.Errorf("tenant schema receipt was not recorded"))
	}
	return nil
}

func reconcileSchema(
	ctx context.Context, database *sql.DB,
	claim tenantmigrationexecutor.Claim, cause error,
) error {
	tenantID, schemaName, namespaceExists, receiptExists, err :=
		inspectSchema(ctx, database, claim)
	if err != nil {
		return errors.Join(tenantmigrationexecutor.ErrAmbiguousOutcome, cause, err)
	}
	if namespaceExists && receiptExists && tenantID == claim.TenantID.String() &&
		schemaName == claim.SchemaName {
		return nil
	}
	if namespaceExists != receiptExists {
		return errors.Join(tenantmigrationexecutor.ErrCleanupRequired, cause)
	}
	return cause
}
