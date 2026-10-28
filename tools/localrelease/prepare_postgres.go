package main

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

func preparePostgres(
	ctx context.Context, adminDSN string, passwords applicationPasswords,
	contract runtimev1.BinarySchemaContract,
) (liveDatabaseIdentity, error) {
	admin, config, err := openAdmin(ctx, adminDSN)
	if err != nil {
		return liveDatabaseIdentity{}, err
	}
	defer admin.Close(context.Background())
	if err := reconcileRoles(ctx, admin); err != nil {
		return liveDatabaseIdentity{}, err
	}
	credentials := map[string]string{
		"harpy_v2_runtime":            passwords.runtime,
		"harpy_migrator":              passwords.migrator,
		"harpy_backup":                passwords.backup,
		"harpy_restore":               passwords.restore,
		"harpy_reactivation_observer": passwords.observer,
	}
	for role, password := range credentials {
		if err := setRolePassword(ctx, admin, role, password); err != nil {
			return liveDatabaseIdentity{}, err
		}
	}
	if err := verifyApplicationCredentials(ctx, admin); err != nil {
		return liveDatabaseIdentity{}, err
	}
	if err := reconcileDatabase(ctx, admin); err != nil {
		return liveDatabaseIdentity{}, err
	}
	target, err := openTarget(ctx, config)
	if err != nil {
		return liveDatabaseIdentity{}, err
	}
	defer target.Close(context.Background())
	if err := reconcileExtensions(ctx, target); err != nil {
		return liveDatabaseIdentity{}, err
	}
	live, err := readLiveDatabaseIdentity(ctx, target)
	if err != nil {
		return liveDatabaseIdentity{}, err
	}
	live.environmentID, err = readExistingEnvironmentID(ctx, target, live)
	if err != nil {
		return liveDatabaseIdentity{}, err
	}
	if err := verifyNewReleaseStage(ctx, target, contract, live.environmentID); err != nil {
		return liveDatabaseIdentity{}, err
	}
	if err := grantExistingRuntimeReadiness(ctx, target); err != nil {
		return liveDatabaseIdentity{}, err
	}
	if err := cleanupPrivilegedRoles(ctx, admin); err != nil {
		return liveDatabaseIdentity{}, err
	}
	return live, nil
}

func verifyApplicationCredentials(ctx context.Context, admin *pgx.Conn) error {
	for _, role := range []string{
		"harpy_v2_runtime", "harpy_migrator",
		"harpy_backup", "harpy_restore",
		"harpy_reactivation_observer",
	} {
		var exact bool
		err := admin.QueryRow(ctx, `SELECT rolcanlogin AND NOT rolinherit
AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole
AND NOT rolreplication AND NOT rolbypassrls AND rolconnlimit=-1
AND rolpassword LIKE 'SCRAM-SHA-256$%' AND rolvaliduntil='infinity'::timestamptz
FROM pg_catalog.pg_authid WHERE rolname=$1`, role).Scan(&exact)
		if err != nil || !exact {
			return fmt.Errorf("application role %s credential state differs: %w", role, err)
		}
	}
	return nil
}

func ensureAdminConnectionAlive(ctx context.Context, admin *pgx.Conn) error {
	if err := admin.Ping(ctx); err != nil {
		return fmt.Errorf("admin connection is not alive: %w", err)
	}
	return nil
}
