package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) preflightDatabase(
	ctx context.Context,
	connection *sql.Conn,
) (postgresstore.LiveDatabaseIdentity, error) {
	live, err := postgresstore.ReadLiveDatabaseIdentity(ctx, connection)
	if err != nil {
		return postgresstore.LiveDatabaseIdentity{}, err
	}
	expected := runner.identity
	if live.Name != expected.databaseName || live.Owner != expected.databaseOwner ||
		live.Encoding != expected.databaseEncoding ||
		live.LocaleProvider != expected.databaseLocaleProvider ||
		live.Collation != expected.databaseCollation || live.CType != expected.databaseCType ||
		live.SystemID != expected.databaseSystemID || live.OID != expected.databaseOID ||
		live.Checksums != expected.dataChecksums || live.ServerVersion != 180004 {
		return postgresstore.LiveDatabaseIdentity{}, fmt.Errorf("live database differs from verified release evidence")
	}
	if live.SessionUser != "harpy_migrator" || live.CurrentUser != "harpy_owner" ||
		live.SearchPath != "pg_catalog" || live.ReadOnly || live.InRecovery {
		return postgresstore.LiveDatabaseIdentity{}, fmt.Errorf("migration database session preflight failed")
	}
	return live, nil
}
