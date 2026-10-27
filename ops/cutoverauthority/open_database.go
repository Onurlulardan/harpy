package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

const authorityApplicationName = "harpy-cutover-authority"

type databaseObservation struct {
	databaseName     string
	systemIdentifier string
	databaseOID      uint32
	sessionUser      string
	currentUser      string
	serverVersion    string
	inRecovery       bool
	superuser        bool
	createDB         bool
	createRole       bool
	inherit          bool
	replication      bool
	bypassRLS        bool
	login            bool
	connectionLimit  int
	validUntil       *time.Time
	membershipCount  int
}

func observeDatabase(
	ctx context.Context, connection *pgx.Conn,
) (databaseObservation, error) {
	var observed databaseObservation
	err := connection.QueryRow(ctx, `SELECT current_database(),
    control.system_identifier::text, database.oid, SESSION_USER, CURRENT_USER,
    pg_catalog.current_setting('server_version_num'), pg_catalog.pg_is_in_recovery(),
    role.rolsuper, role.rolcreatedb, role.rolcreaterole, role.rolinherit,
    role.rolreplication, role.rolbypassrls, role.rolcanlogin, role.rolconnlimit,
    role.rolvaliduntil,
    (SELECT pg_catalog.count(*) FROM pg_catalog.pg_auth_members AS membership
     WHERE membership.roleid=role.oid OR membership.member=role.oid)
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()
JOIN pg_catalog.pg_roles AS role ON role.rolname=SESSION_USER`).Scan(
		&observed.databaseName, &observed.systemIdentifier, &observed.databaseOID,
		&observed.sessionUser, &observed.currentUser, &observed.serverVersion,
		&observed.inRecovery, &observed.superuser, &observed.createDB,
		&observed.createRole, &observed.inherit, &observed.replication,
		&observed.bypassRLS, &observed.login, &observed.connectionLimit,
		&observed.validUntil, &observed.membershipCount,
	)
	if err != nil {
		return databaseObservation{}, fmt.Errorf("authority database preflight failed")
	}
	return observed, nil
}

func preflightRoot(
	ctx context.Context, connection *pgx.Conn,
	target cutoverv1.DatabaseIdentity, expectedSession string,
) error {
	observed, err := observeDatabase(ctx, connection)
	if err != nil {
		return err
	}
	if !matchesDatabaseIdentity(observed, target) || observed.serverVersion != "180004" ||
		observed.inRecovery || observed.sessionUser != expectedSession ||
		observed.currentUser != expectedSession || !observed.superuser ||
		strings.HasPrefix(observed.sessionUser, "harpy_") {
		return fmt.Errorf("external root database preflight differs")
	}
	return nil
}

func preflightClusterAdmin(
	ctx context.Context, connection *pgx.Conn,
	target cutoverv1.DatabaseIdentity, now time.Time,
) error {
	observed, err := observeDatabase(ctx, connection)
	if err != nil {
		return err
	}
	if !matchesDatabaseIdentity(observed, target) || observed.serverVersion != "180004" ||
		observed.inRecovery || observed.sessionUser != "harpy_cluster_admin" ||
		observed.currentUser != observed.sessionUser || observed.superuser ||
		!observed.createDB || !observed.createRole || observed.inherit ||
		observed.replication || observed.bypassRLS || !observed.login ||
		observed.connectionLimit != 1 || observed.validUntil == nil ||
		!now.Before(*observed.validUntil) || observed.membershipCount != 0 {
		return fmt.Errorf("cluster admin database preflight differs")
	}
	return nil
}

func preflightFinalizer(
	ctx context.Context, connection *pgx.Conn, target cutoverv1.DatabaseIdentity,
	now time.Time,
) (time.Time, error) {
	observed, err := observeDatabase(ctx, connection)
	if err != nil {
		return time.Time{}, err
	}
	if !matchesDatabaseIdentity(observed, target) || observed.serverVersion != "180004" ||
		observed.inRecovery || observed.sessionUser != "harpy_cutover_finalizer" ||
		observed.currentUser != observed.sessionUser || observed.superuser ||
		observed.createDB || observed.createRole || observed.inherit ||
		observed.replication || observed.bypassRLS || !observed.login ||
		observed.connectionLimit != 1 || observed.validUntil == nil ||
		!now.Before(*observed.validUntil) || observed.membershipCount != 0 {
		return time.Time{}, fmt.Errorf("finalizer database preflight differs")
	}
	return observed.validUntil.UTC(), nil
}

func matchesDatabaseIdentity(
	observed databaseObservation, expected cutoverv1.DatabaseIdentity,
) bool {
	return observed.databaseName == expected.DatabaseName &&
		observed.systemIdentifier == expected.SystemIdentifier &&
		observed.databaseOID == expected.DatabaseOID
}
