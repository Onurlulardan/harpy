//go:build integration

package migrator

import (
	"context"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type builtEntrypointLiveIdentity struct {
	systemID    string
	databaseOID uint32
}

func readBuiltEntrypointLiveIdentity(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
) builtEntrypointLiveIdentity {
	t.Helper()
	var (
		name, owner, encoding, provider, collation, ctype, systemID string
		databaseOID                                                 uint32
		checksums                                                   bool
		serverVersion                                               int
	)
	err := database.admin.QueryRow(ctx, `
SELECT current_database(), owner.rolname, current_setting('server_encoding'),
CASE database.datlocprovider WHEN 'c' THEN 'libc' ELSE database.datlocprovider::text END,
database.datcollate, database.datctype, control.system_identifier::text,
database.oid, current_setting('data_checksums') = 'on',
current_setting('server_version_num')::integer
FROM pg_catalog.pg_database AS database
JOIN pg_catalog.pg_roles AS owner ON owner.oid=database.datdba
CROSS JOIN pg_catalog.pg_control_system() AS control
WHERE database.datname=current_database()`).Scan(
		&name, &owner, &encoding, &provider, &collation, &ctype,
		&systemID, &databaseOID, &checksums, &serverVersion,
	)
	if err != nil {
		t.Fatalf("read built entrypoint live database identity: %v", err)
	}
	if name != builtEntrypointDatabaseName || owner != runtimev1.DatabaseOwnerRole ||
		encoding != runtimev1.PostgresEncoding || provider != runtimev1.PostgresLocaleProvider ||
		collation != runtimev1.PostgresCollation || ctype != runtimev1.PostgresCharacterType ||
		!checksums || serverVersion != runtimev1.PostgresServerVersionNumber ||
		systemID == "" || databaseOID == 0 {
		t.Fatalf(
			"built entrypoint database identity differs: %s %s %s %s %s %s %t %d",
			name, owner, encoding, provider, collation, ctype, checksums, serverVersion,
		)
	}
	return builtEntrypointLiveIdentity{systemID: systemID, databaseOID: databaseOID}
}
