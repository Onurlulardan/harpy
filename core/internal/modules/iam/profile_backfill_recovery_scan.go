package iam

import (
	"context"
	"errors"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ProfileBackfillRecoveryRuntime) scanRecoveryInventory(
	ctx context.Context,
	scope platformtx.TenantScope,
	contract profilebackfillpostgres.RecoveryContract,
	evidence profilebackfillapp.Evidence,
	classifier *ProfileRecoveryInventoryClassifier,
) (profilebackfillapp.RecoveryInventory, error) {
	if ctx == nil {
		return profilebackfillapp.RecoveryInventory{}, profilebackfillapp.ErrInvalid
	}
	scanContext, cancel := context.WithTimeoutCause(
		ctx, profilebackfillapp.MaximumRecoveryInventoryDuration,
		profilebackfillapp.ErrBudgetExceeded,
	)
	defer cancel()
	reader, err := bindProfileHistoricalSchemaReader(scope)
	if err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	cursor, err := runtime.source.OpenInventory(
		scanContext, scope, contract.Parent.ID, evidence,
	)
	if err != nil {
		if errors.Is(context.Cause(scanContext), profilebackfillapp.ErrBudgetExceeded) {
			return profilebackfillapp.RecoveryInventory{}, profilebackfillapp.ErrBudgetExceeded
		}
		return profilebackfillapp.RecoveryInventory{}, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	var after uuid.UUID
	for {
		if err := classifier.Check(scanContext); err != nil {
			return profilebackfillapp.RecoveryInventory{}, err
		}
		rows, readErr := cursor.Fetch(scanContext)
		if readErr != nil {
			if errors.Is(context.Cause(scanContext), profilebackfillapp.ErrBudgetExceeded) {
				return profilebackfillapp.RecoveryInventory{}, profilebackfillapp.ErrBudgetExceeded
			}
			return profilebackfillapp.RecoveryInventory{}, readErr
		}
		if len(rows) == 0 {
			return finishRecoveryInventory(ctx, scanContext, cursor, classifier)
		}
		if len(rows) > profilebackfillapp.RecoveryInventoryReadBatchSize ||
			!validRecoveryInventoryRows(contract, after, rows) {
			wipeRecoveryInventoryRows(rows)
			return profilebackfillapp.RecoveryInventory{}, profilebackfillapp.ErrUnavailable
		}
		schemas, schemaErr := recoveryActualSchemas(scanContext, reader, contract, rows)
		if schemaErr != nil {
			wipeRecoveryInventoryRows(rows)
			return profilebackfillapp.RecoveryInventory{}, schemaErr
		}
		after = rows[len(rows)-1].Resource.ID
		resources := detachRecoveryInventoryRows(rows)
		if err := classifier.AppendBatch(scanContext, schemas, resources); err != nil {
			return profilebackfillapp.RecoveryInventory{}, err
		}
		if len(rows) < profilebackfillapp.RecoveryInventoryReadBatchSize {
			return finishRecoveryInventory(ctx, scanContext, cursor, classifier)
		}
	}
}

func finishRecoveryInventory(
	ctx context.Context,
	scanContext context.Context,
	cursor profilebackfillpostgres.RecoveryInventoryCursor,
	classifier *ProfileRecoveryInventoryClassifier,
) (profilebackfillapp.RecoveryInventory, error) {
	if err := cursor.Close(ctx); err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	return classifier.Finish(scanContext)
}

func recoveryActualSchemas(
	ctx context.Context,
	reader profileHistoricalSchemaReader,
	contract profilebackfillpostgres.RecoveryContract,
	rows []profilebackfillpostgres.RecoveryInventoryRow,
) (map[uuid.UUID]profileapp.Schema, error) {
	result := map[uuid.UUID]profileapp.Schema{
		contract.Source.ID(): contract.Source,
		contract.Target.ID(): contract.Target,
	}
	for _, row := range rows {
		if _, found := result[row.Resource.SchemaVersionID]; found {
			continue
		}
		schema, err := reader.HistoricalSchema(ctx, row.Resource.SchemaVersionID)
		if err != nil {
			return nil, profilebackfillapp.ErrUnavailable
		}
		if !validProfileRecoverySource(
			contract.Source, schema, row.Resource.SchemaVersionID,
		) || schema.Version() != row.SchemaVersion {
			return nil, profilebackfillapp.ErrInvalid
		}
		result[row.Resource.SchemaVersionID] = schema
	}
	return result, nil
}

func detachRecoveryInventoryRows(
	rows []profilebackfillpostgres.RecoveryInventoryRow,
) []profilebackfillapp.SourceResource {
	resources := make([]profilebackfillapp.SourceResource, len(rows))
	for index := range rows {
		resources[index] = rows[index].Resource
		rows[index].Resource.Document = nil
	}
	return resources
}

func wipeRecoveryInventoryRows(
	rows []profilebackfillpostgres.RecoveryInventoryRow,
) {
	for index := range rows {
		clear(rows[index].Resource.Document)
		rows[index].Resource.Document = nil
	}
}
