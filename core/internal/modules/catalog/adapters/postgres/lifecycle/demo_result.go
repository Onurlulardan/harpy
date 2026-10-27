package lifecycle

import (
	"context"
	_ "embed"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/insert_demo_receipt.sql
var insertDemoReceiptSQL string

//go:embed sql/tenant/get_demo_receipt.sql
var getDemoReceiptSQL string

//go:embed sql/tenant/update_demo_state.sql
var updateDemoStateSQL string

type demoReceipt struct {
	receiptID, installationID, operationID, componentID uuid.UUID
	action                                              string
	recordIDs                                           []uuid.UUID
	contentSHA256                                       string
	actorID                                             uuid.UUID
	idempotencyKey                                      string
	completedAt                                         time.Time
}

func (storage *store) ApplyDemoResult(
	ctx context.Context,
	mutation catalogapp.DemoResultMutation,
) (catalogapp.DemoComponent, error) {
	if !validDemoResultMutation(storage, mutation) {
		return catalogapp.DemoComponent{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return catalogapp.DemoComponent{}, unavailable()
	}
	current, err := storage.currentDemoComponent(
		ctx, mutation.InstallationID, mutation.ComponentID,
	)
	if err != nil {
		return catalogapp.DemoComponent{}, err
	}
	recordIDs, _ := encodeDemoRecordIDs(mutation.RecordIDs)
	inserted, err := storage.execute(ctx, insertDemoReceiptSQL,
		mutation.ReceiptID, mutation.InstallationID, mutation.OperationID,
		mutation.ComponentID, mutation.Action, recordIDs, mutation.ContentSHA256,
		storage.actorID, mutation.IdempotencyKey, mutation.CompletedAt,
	)
	if err != nil {
		return catalogapp.DemoComponent{}, unavailable()
	}
	receipt, err := storage.readDemoReceipt(ctx, mutation)
	if err != nil || !demoReceiptMatches(receipt, mutation, storage.actorID) {
		return catalogapp.DemoComponent{}, catalogapp.ErrConflict
	}
	if inserted == 0 {
		if current.State == resultDemoState(mutation.Action) &&
			current.Revision == mutation.ExpectedRevision+1 {
			return current, nil
		}
		return catalogapp.DemoComponent{}, catalogapp.ErrConflict
	}
	wantCurrent := catalogapp.DemoStateAbsent
	if mutation.Action == catalogapp.DemoActionRemove {
		wantCurrent = catalogapp.DemoStateInstalled
	}
	if current.State != wantCurrent || current.Revision != mutation.ExpectedRevision {
		return catalogapp.DemoComponent{}, catalogapp.ErrConflict
	}
	stateRecordIDs := recordIDs
	if mutation.Action == catalogapp.DemoActionRemove {
		stateRecordIDs = []byte("[]")
	}
	var revision int64
	err = storage.queryRow(ctx, updateDemoStateSQL,
		mutation.InstallationID, mutation.ComponentID, mutation.SourceRunID,
		resultDemoState(mutation.Action), stateRecordIDs, mutation.ReceiptID,
		mutation.CompletedAt, mutation.ContentSHA256, mutation.ExpectedRevision,
	).Scan(&revision)
	if err != nil || revision != int64(mutation.ExpectedRevision+1) {
		return catalogapp.DemoComponent{}, catalogapp.ErrConflict
	}
	return storage.currentDemoComponent(ctx, mutation.InstallationID, mutation.ComponentID)
}

func (storage *store) readDemoReceipt(
	ctx context.Context,
	mutation catalogapp.DemoResultMutation,
) (demoReceipt, error) {
	var value demoReceipt
	var recordIDs []byte
	err := storage.queryRow(ctx, getDemoReceiptSQL,
		mutation.InstallationID, mutation.ComponentID,
		mutation.Action, mutation.IdempotencyKey,
	).Scan(&value.receiptID, &value.installationID, &value.operationID,
		&value.componentID, &value.action, &recordIDs, &value.contentSHA256,
		&value.actorID, &value.idempotencyKey, &value.completedAt)
	if err != nil {
		return demoReceipt{}, unavailable()
	}
	value.completedAt = value.completedAt.UTC()
	value.recordIDs, err = decodeDemoRecordIDs(recordIDs)
	return value, err
}

func (storage *store) currentDemoComponent(
	ctx context.Context,
	installationID, componentID uuid.UUID,
) (catalogapp.DemoComponent, error) {
	components, err := storage.DemoComponents(ctx, installationID)
	if err != nil {
		return catalogapp.DemoComponent{}, err
	}
	for _, component := range components {
		if component.ComponentID == componentID {
			return component, nil
		}
	}
	return catalogapp.DemoComponent{}, catalogapp.ErrNotFound
}

func validDemoResultMutation(storage *store, value catalogapp.DemoResultMutation) bool {
	ids, err := canonicalDemoRecordIDs(value.RecordIDs)
	return storage != nil && value.OperationID.Version() == 7 && value.ReceiptID.Version() == 7 &&
		value.InstallationID.Version() == 7 && value.ComponentID.Version() == 7 &&
		value.SourceRunID.Version() == 7 &&
		(value.Action == catalogapp.DemoActionInstall || value.Action == catalogapp.DemoActionRemove) &&
		len(ids) > 0 && err == nil && sameDemoUUIDs(ids, value.RecordIDs) &&
		digestPattern.MatchString(value.ContentSHA256) &&
		idempotencyPattern.MatchString(value.IdempotencyKey) &&
		value.ExpectedRevision > 0 && canonicalDemoTime(value.CompletedAt)
}

func demoReceiptMatches(
	value demoReceipt,
	mutation catalogapp.DemoResultMutation,
	actorID uuid.UUID,
) bool {
	return value.receiptID == mutation.ReceiptID && value.installationID == mutation.InstallationID &&
		value.operationID == mutation.OperationID && value.componentID == mutation.ComponentID &&
		value.action == mutation.Action && sameDemoUUIDs(value.recordIDs, mutation.RecordIDs) &&
		value.contentSHA256 == mutation.ContentSHA256 && value.actorID == actorID &&
		value.idempotencyKey == mutation.IdempotencyKey && value.completedAt.Equal(mutation.CompletedAt)
}

func resultDemoState(action string) string {
	if action == catalogapp.DemoActionInstall {
		return catalogapp.DemoStateInstalled
	}
	return catalogapp.DemoStateAbsent
}
