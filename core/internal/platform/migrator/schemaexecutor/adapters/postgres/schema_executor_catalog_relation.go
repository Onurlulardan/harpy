package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type schemaExecutorCatalogQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func schemaExecutorOwnedTable(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
) (bool, error) {
	var kind, owner string
	err := querier.QueryRow(
		ctx, schemaExecutorReadRelationOwnerSQL, schemaName, tableName,
	).Scan(&kind, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if kind != "r" || owner != "harpy_owner" {
		return false, schemaExecutorBlocked("schema_object_ownership_invalid", nil)
	}
	return true, nil
}

func schemaExecutorRequireTable(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	tableName string,
) error {
	exists, err := schemaExecutorOwnedTable(ctx, querier, schemaName, tableName)
	if err != nil {
		return err
	}
	if !exists {
		return schemaExecutorBlocked("schema_object_missing", nil)
	}
	return nil
}
