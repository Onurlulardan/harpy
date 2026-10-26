package adoption

import (
	"context"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func (store *Store) AdoptionStart(ctx context.Context, command application.StartCommand) (
	application.Run, bool, error,
) {
	var document []byte
	var replayed bool
	err := store.within(ctx, func(database Database) error {
		locator, manifest := command.Locator, command.Manifest
		locatorSHA := locator.AttestationSHA256()
		return database.QueryRow(ctx, startSQL, command.OperationID, command.RunID,
			locator.Kind(), nullableUUID(locator.TenantID()), locator.Schema(), locator.Relation(),
			locator.RelationOID(), locator.TenantSlug(), locator.TenantName(),
			locatorSHA[:], manifest.ManifestSHA256[:],
			command.TransformSHA256[:], manifest.SnapshotID, manifest.RowCount,
			manifest.RootSHA256[:], command.ObservedAt).Scan(&document, &replayed)
	})
	if err != nil {
		return application.Run{}, false, normalize(err)
	}
	run, err := decodeRun(document)
	return run, replayed, normalize(err)
}
