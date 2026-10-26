package adoption

import (
	"context"
	"encoding/json"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (store *Store) Start(ctx context.Context, command app.StartCommand) (app.Run, bool, error) {
	var document []byte
	var replayed bool
	err := store.within(ctx, func(database Database) error {
		return database.QueryRow(ctx, startSQL, command.OperationID, command.RunID,
			string(command.Manifest.Kind), command.Manifest.Schema, command.Manifest.Relation,
			command.Manifest.RelationOID, command.Manifest.RowCount,
			command.Manifest.RootSHA256[:], command.Manifest.SnapshotID,
			command.TransformSHA256[:], command.ObservedAt).Scan(&document, &replayed)
	})
	if err != nil {
		return app.Run{}, false, normalize(err)
	}
	run, err := decodeRun(document)
	return run, replayed, normalize(err)
}

func (store *Store) Claim(ctx context.Context, command app.ClaimCommand) (app.Run, error) {
	return store.runDocument(ctx, claimSQL, command.OperationID, command.RunID,
		command.LeaseOwner, command.ExpectedRevision, command.Manifest.RelationOID,
		command.Manifest.RowCount, command.Manifest.RootSHA256[:], command.ObservedAt)
}

func (store *Store) ApplyPage(ctx context.Context, command app.ApplyPageCommand) (
	app.PageReceipt, error,
) {
	records, err := json.Marshal(command.Records)
	if err != nil || len(records) > 1024*1024 {
		return app.PageReceipt{}, domain.ErrInvalid
	}
	var document []byte
	err = store.within(ctx, func(database Database) error {
		manifest := command.Manifest
		return database.QueryRow(ctx, applySQL, command.OperationID, command.RunID,
			command.LeaseOwner, command.LeaseGeneration, command.PageNumber,
			nullableUUID(command.CursorBefore), command.CursorAfter, records,
			command.SourcePageSHA256[:], command.TargetSHA256[:], manifest.RelationOID,
			manifest.RowCount, manifest.RootSHA256[:], command.ObservedAt).Scan(&document)
	})
	if err != nil {
		return app.PageReceipt{}, normalize(err)
	}
	return decodeReceipt(document)
}

func (store *Store) Advance(ctx context.Context, command app.ApplyPageCommand) (app.Run, error) {
	manifest := command.Manifest
	return store.runDocument(ctx, advanceSQL, command.OperationID, command.RunID,
		command.LeaseOwner, command.LeaseGeneration, command.PageNumber,
		nullableUUID(command.CursorBefore), command.CursorAfter, command.SourcePageSHA256[:],
		command.TargetSHA256[:], manifest.RelationOID, manifest.RowCount,
		manifest.RootSHA256[:], command.ObservedAt)
}

func (store *Store) Reconcile(ctx context.Context, command app.LeaseCommand) (app.Run, error) {
	return store.leaseDocument(ctx, reconcileSQL, command)
}

func (store *Store) Complete(ctx context.Context, command app.LeaseCommand) (app.Run, error) {
	return store.leaseDocument(ctx, completeSQL, command)
}
