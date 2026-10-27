package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) preflightExtensions(
	ctx context.Context,
	connection *sql.Conn,
) error {
	for _, expected := range runner.evidence.extensions {
		identity, err := postgresstore.ReadExtensionIdentity(ctx, connection, expected.name)
		if err != nil {
			return err
		}
		if identity.Version != expected.version || identity.SchemaName != expected.schemaName ||
			identity.OwnerRole != expected.ownerRole {
			return fmt.Errorf("extension %s differs from verified evidence", expected.name)
		}
	}
	return nil
}
