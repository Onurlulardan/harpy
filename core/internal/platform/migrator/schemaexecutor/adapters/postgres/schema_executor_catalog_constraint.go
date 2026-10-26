package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type schemaExecutorConstraintContract struct {
	owner      string
	kind       string
	validated  bool
	definition string
}

func schemaExecutorReadConstraint(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
	constraintName string,
) (schemaExecutorConstraintContract, bool, error) {
	var contract schemaExecutorConstraintContract
	err := querier.QueryRow(
		ctx, schemaExecutorReadConstraintSQL,
		schemaName, tableName, constraintName,
	).Scan(&contract.owner, &contract.kind, &contract.validated, &contract.definition)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorConstraintContract{}, false, nil
	}
	return contract, err == nil, err
}
