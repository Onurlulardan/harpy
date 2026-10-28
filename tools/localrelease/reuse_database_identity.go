package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func verifyReuseDatabase(
	ctx context.Context, adminDSN string, passwords applicationPasswords,
	expected existingEvidenceIdentity, foundation bool,
) error {
	admin, config, err := openAdmin(ctx, adminDSN)
	if err != nil {
		return err
	}
	defer admin.Close(context.Background())
	target, err := openTarget(ctx, config)
	if err != nil {
		return err
	}
	defer target.Close(context.Background())
	live, err := readLiveDatabaseIdentity(ctx, target)
	if err != nil {
		return err
	}
	if live.systemID != expected.systemID || live.databaseOID != expected.databaseOID {
		return fmt.Errorf("existing release belongs to a different physical database")
	}
	if err := verifyExtensions(ctx, target); err != nil {
		return err
	}
	if err := verifyManagedRoleSet(ctx, admin, rolesManifest()); err != nil {
		return err
	}
	if err := verifyRequiredMemberships(ctx, admin); err != nil {
		return err
	}
	if err := verifyApplicationCredentials(ctx, admin); err != nil {
		return err
	}
	for _, role := range privilegedRoles {
		if err := verifyCleanedRole(ctx, admin, role); err != nil {
			return err
		}
	}
	if err := verifyApplicationLogin(
		ctx, config, "harpy_v2_runtime", "harpy_runtime", passwords.runtime,
	); err != nil {
		return err
	}
	if err := verifyApplicationLogin(
		ctx, config, "harpy_migrator", "harpy_owner", passwords.migrator,
	); err != nil {
		return err
	}
	if err := verifyApplicationLogin(
		ctx, config, "harpy_backup", "harpy_backup_reader", passwords.backup,
	); err != nil {
		return err
	}
	if err := verifyApplicationLogin(
		ctx, config, "harpy_restore", "harpy_restore", passwords.restore,
	); err != nil {
		return err
	}
	if err := verifyApplicationLogin(
		ctx, config, "harpy_reactivation_observer",
		"harpy_reactivation_observer", passwords.observer,
	); err != nil {
		return err
	}
	if err := verifyStoredEnvironmentIdentity(ctx, target, expected, foundation); err != nil {
		return err
	}
	if !foundation {
		return verifyFoundationRelease(ctx, target, expected.environmentID)
	}
	return nil
}

func verifyApplicationLogin(
	ctx context.Context, admin *pgx.ConnConfig, sessionRole, currentRole, password string,
) error {
	config := admin.Copy()
	config.Database, config.User, config.Password = databaseName, sessionRole, password
	if config.RuntimeParams == nil {
		config.RuntimeParams = make(map[string]string)
	}
	config.RuntimeParams["role"] = currentRole
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("verify %s login: %w", sessionRole, err)
	}
	defer connection.Close(context.Background())
	var sessionUser, currentUser string
	if err := connection.QueryRow(ctx, "SELECT session_user, current_user").Scan(
		&sessionUser, &currentUser,
	); err != nil {
		return fmt.Errorf("verify %s role selection: %w", sessionRole, err)
	}
	if sessionUser != sessionRole || currentUser != currentRole {
		return fmt.Errorf("application role identity differs for %s", sessionRole)
	}
	return nil
}

func verifyStoredEnvironmentIdentity(
	ctx context.Context, target *pgx.Conn, expected existingEvidenceIdentity,
	requireBootstrapBinding bool,
) error {
	var exists bool
	if err := target.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass('platform.environment_identity') IS NOT NULL",
	).Scan(&exists); err != nil {
		return fmt.Errorf("inspect stored environment identity: %w", err)
	}
	if !exists {
		return fmt.Errorf("stored environment identity is missing; existing release cannot be reused")
	}
	var environmentID, origin, systemID, releaseID, authorityDigest string
	var databaseOID uint32
	err := target.QueryRow(ctx, `SELECT environment_id::text, database_origin,
database_system_identifier, database_oid, release_id::text, bootstrap_authority_sha256
FROM platform.environment_identity WHERE singleton=TRUE`).Scan(
		&environmentID, &origin, &systemID, &databaseOID, &releaseID, &authorityDigest,
	)
	if err != nil {
		return fmt.Errorf("read stored environment identity: %w", err)
	}
	if environmentID != expected.environmentID || origin != "fresh" ||
		systemID != expected.systemID || databaseOID != expected.databaseOID ||
		(requireBootstrapBinding && (releaseID != expected.releaseID ||
			authorityDigest != expected.authorityDigest)) {
		return fmt.Errorf("stored environment identity differs from existing release evidence")
	}
	return nil
}
