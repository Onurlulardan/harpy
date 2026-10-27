package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type schemaExecutorColumnContract struct {
	typeName          string
	notNull           bool
	defaultExpression string
	defaultCollation  bool
}

func schemaExecutorReadColumn(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	columnName string,
) (schemaExecutorColumnContract, bool, error) {
	var contract schemaExecutorColumnContract
	err := querier.QueryRow(
		ctx, schemaExecutorReadColumnSQL, schemaName, tableName, columnName,
	).Scan(
		&contract.typeName, &contract.notNull,
		&contract.defaultExpression, &contract.defaultCollation,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorColumnContract{}, false, nil
	}
	return contract, err == nil, err
}
