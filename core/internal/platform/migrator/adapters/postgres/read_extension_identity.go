package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed sql/read_extension_identity.sql
var readExtensionIdentitySQL string

type ExtensionIdentity struct {
	Version    string
	SchemaName string
	OwnerRole  string
}

func ReadExtensionIdentity(
	ctx context.Context,
	connection *sql.Conn,
	name string,
) (ExtensionIdentity, error) {
	var identity ExtensionIdentity
	if err := connection.QueryRowContext(ctx, readExtensionIdentitySQL, name).Scan(
		&identity.Version, &identity.SchemaName, &identity.OwnerRole,
	); err != nil {
		return ExtensionIdentity{}, fmt.Errorf("read extension %s inventory: %w", name, err)
	}
	return identity, nil
}
