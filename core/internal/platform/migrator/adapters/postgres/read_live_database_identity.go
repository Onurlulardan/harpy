package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/read_live_database_identity.sql
var readLiveDatabaseIdentitySQL string

type LiveDatabaseIdentity struct {
	Name           string
	Owner          string
	Encoding       string
	LocaleProvider string
	Collation      string
	CType          string
	SystemID       string
	OID            uint32
	Checksums      bool
	ServerVersion  int
	SessionUser    string
	CurrentUser    string
	SearchPath     string
	ReadOnly       bool
	InRecovery     bool
}

func ReadLiveDatabaseIdentity(
	ctx context.Context,
	connection *sql.Conn,
) (LiveDatabaseIdentity, error) {
	var live LiveDatabaseIdentity
	err := connection.QueryRowContext(ctx, readLiveDatabaseIdentitySQL).Scan(
		&live.Name, &live.Owner, &live.Encoding, &live.LocaleProvider,
		&live.Collation, &live.CType, &live.SystemID, &live.OID,
		&live.Checksums, &live.ServerVersion, &live.SessionUser,
		&live.CurrentUser, &live.SearchPath, &live.ReadOnly, &live.InRecovery,
	)
	if err != nil {
		return LiveDatabaseIdentity{}, fmt.Errorf("read migration database identity: %w", err)
	}
	return live, nil
}
