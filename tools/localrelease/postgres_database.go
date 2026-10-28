package main

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

type liveDatabaseIdentity struct {
	name, owner, encoding, provider, collation, ctype, systemID string
	environmentID, appliedReleaseID                             string
	databaseOID                                                 uint32
	checksums                                                   bool
	serverVersion                                               int
}

func reconcileDatabase(ctx context.Context, admin *pgx.Conn) error {
	var exists bool
	if err := admin.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_database WHERE datname=$1)", databaseName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("inspect target database: %w", err)
	}
	if !exists {
		_, err := admin.Exec(ctx, `CREATE DATABASE harpy
WITH OWNER harpy_owner TEMPLATE template0 ENCODING 'UTF8'
LOCALE_PROVIDER libc LC_COLLATE 'en_US.utf8' LC_CTYPE 'en_US.utf8'`)
		if err != nil {
			return fmt.Errorf("create target database: %w", err)
		}
	}
	if _, err := admin.Exec(ctx,
		"REVOKE CONNECT, TEMPORARY ON DATABASE harpy FROM PUBLIC",
	); err != nil {
		return fmt.Errorf("revoke public database privileges: %w", err)
	}
	tx, err := admin.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		return fmt.Errorf("assume database owner: %w", err)
	}
	if _, err := tx.Exec(ctx, `GRANT CONNECT ON DATABASE harpy TO
harpy_owner, harpy_runtime, harpy_v2_runtime, harpy_migrator,
harpy_cluster_admin, harpy_extension_admin, harpy_backup,
harpy_backup_reader, harpy_restore, harpy_reactivation_observer`); err != nil {
		return fmt.Errorf("grant target database connect: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit database privileges: %w", err)
	}
	return nil
}

func readLiveDatabaseIdentity(
	ctx context.Context, target *pgx.Conn,
) (liveDatabaseIdentity, error) {
	var identity liveDatabaseIdentity
	err := target.QueryRow(ctx, `SELECT current_database(), owner.rolname,
current_setting('server_encoding'), CASE database.datlocprovider
WHEN 'c' THEN 'libc' ELSE database.datlocprovider::text END,
database.datcollate, database.datctype, control.system_identifier::text,
database.oid, current_setting('data_checksums')='on',
current_setting('server_version_num')::integer
FROM pg_catalog.pg_database AS database
JOIN pg_catalog.pg_roles AS owner ON owner.oid=database.datdba
CROSS JOIN pg_catalog.pg_control_system() AS control
WHERE database.datname=current_database()`).Scan(
		&identity.name, &identity.owner, &identity.encoding, &identity.provider,
		&identity.collation, &identity.ctype, &identity.systemID,
		&identity.databaseOID, &identity.checksums, &identity.serverVersion,
	)
	if err != nil {
		return liveDatabaseIdentity{}, fmt.Errorf("read target database identity: %w", err)
	}
	if identity.name != databaseName || identity.owner != runtimev1.DatabaseOwnerRole ||
		identity.encoding != runtimev1.PostgresEncoding ||
		identity.provider != runtimev1.PostgresLocaleProvider ||
		identity.collation != runtimev1.PostgresCollation ||
		identity.ctype != runtimev1.PostgresCharacterType || !identity.checksums ||
		identity.serverVersion != runtimev1.PostgresServerVersionNumber ||
		identity.systemID == "" || identity.databaseOID == 0 {
		return liveDatabaseIdentity{}, fmt.Errorf("target database identity differs from PG18 contract")
	}
	return identity, nil
}
