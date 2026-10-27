package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type schemaExecutorKeyContract struct {
	owner        string
	kind         string
	validated    bool
	deleteAction string
	targetSchema string
	targetTable  string
	targetOwner  string
	updateAction string
	matchType    string
	deferrable   bool
	deferred     bool
	sourceFields []string
	targetFields []string
}

func schemaExecutorReadKey(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	constraintName string,
) (schemaExecutorKeyContract, bool, error) {
	var contract schemaExecutorKeyContract
	err := querier.QueryRow(
		ctx, schemaExecutorReadKeyContractSQL,
		schemaName, tableName, constraintName,
	).Scan(
		&contract.owner, &contract.kind, &contract.validated, &contract.deleteAction,
		&contract.targetSchema, &contract.targetTable, &contract.targetOwner,
		&contract.updateAction, &contract.matchType,
		&contract.deferrable, &contract.deferred,
		&contract.sourceFields, &contract.targetFields,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorKeyContract{}, false, nil
	}
	return contract, err == nil, err
}

func schemaExecutorTableCount(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	query string,
	schemaName string,
	tableName string,
) (int64, error) {
	var count int64
	err := querier.QueryRow(ctx, query, schemaName, tableName).Scan(&count)
	return count, err
}
