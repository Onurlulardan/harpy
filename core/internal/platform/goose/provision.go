package goose

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

func ProvisionTenant(ctx context.Context, cfg foundation.Config, tenantID [16]byte) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	schemaName, err := foundation.TenantSchemaName(tenantID)
	if err != nil {
		return err
	}
	password, err := tenantPassword()
	if err != nil {
		return err
	}
	namespace, err := foundation.PrepareTenantNamespace(ctx, cfg, tenantID, password)
	if err != nil {
		return err
	}
	if namespace.Schema != schemaName {
		return fmt.Errorf("tenant schema %s was not prepared", schemaName)
	}
	dsn, err := userDSN(cfg.AdminDSN, namespace.Role, password, cfg.Database)
	if err != nil {
		return err
	}
	source, err := fs.Sub(tenantRoot, "sql/tenant")
	if err != nil {
		return fmt.Errorf("open tenant goose source: %w", err)
	}
	database, err := openGoose(ctx, dsn)
	if err != nil {
		return err
	}
	defer database.Close()
	if err = lockTenant(ctx, database, tenantID); err != nil {
		return err
	}
	if err = upDatabase(ctx, database, source); err != nil {
		return err
	}
	_, err = database.ExecContext(ctx, `INSERT INTO tenancy_schema_identity (
    singleton, tenant_id, schema_name, origin, created_at
) VALUES (true, $1, $2, 'new', pg_catalog.clock_timestamp())
ON CONFLICT (singleton) DO NOTHING`, formatUUID(tenantID), schemaName)
	return err
}

func tenantPassword() (string, error) {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate tenant role password: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

func formatUUID(id [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
}

func lockTenant(ctx context.Context, database *sql.DB, tenantID [16]byte) error {
	keyA := int32(int(tenantID[0])<<24 | int(tenantID[1])<<16 | int(tenantID[2])<<8 | int(tenantID[3]))
	keyB := int32(int(tenantID[4])<<24 | int(tenantID[5])<<16 | int(tenantID[6])<<8 | int(tenantID[7]))
	var locked bool
	err := database.QueryRowContext(
		ctx, "SELECT pg_catalog.pg_try_advisory_lock($1, $2)", keyA, keyB,
	).Scan(&locked)
	if err != nil {
		return fmt.Errorf("lock tenant provisioning: %w", err)
	}
	if !locked {
		return fmt.Errorf("tenant provisioning is already running")
	}
	return nil
}
