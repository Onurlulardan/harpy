package adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/google/uuid"
)

func (store *Store) leaseDocument(ctx context.Context, statement string,
	command app.LeaseCommand,
) (app.Run, error) {
	manifest := command.Manifest
	return store.runDocument(ctx, statement, command.OperationID, command.RunID,
		command.LeaseOwner, command.LeaseGeneration, manifest.RelationOID,
		manifest.RowCount, manifest.RootSHA256[:], command.ObservedAt)
}

func (store *Store) runDocument(ctx context.Context, statement string,
	arguments ...any,
) (app.Run, error) {
	var document []byte
	err := store.within(ctx, func(database Database) error {
		return database.QueryRow(ctx, statement, arguments...).Scan(&document)
	})
	if err != nil {
		return app.Run{}, normalize(err)
	}
	run, err := decodeRun(document)
	return run, normalize(err)
}

func nullableUUID(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return *value
}

var _ app.Target = (*Store)(nil)
