package iam

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type profileHistoricalSchemaReader interface {
	HistoricalSchema(context.Context, uuid.UUID) (profileapp.Schema, error)
}

func bindProfileHistoricalSchemaReader(
	scope platformtx.TenantScope,
) (profileHistoricalSchemaReader, error) {
	reader, err := iampostgres.BindProfileReader(scope)
	if err != nil || reader == nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return reader, nil
}

func profileBackfillActualSchemas(
	ctx context.Context,
	reader profileHistoricalSchemaReader,
	current profileapp.Schema,
	rows []profilebackfillapp.SourceResource,
) (map[uuid.UUID]profileapp.Schema, error) {
	if ctx == nil || ctx.Err() != nil || reader == nil || len(rows) == 0 ||
		len(rows) > profilebackfillapp.MaximumPageSize {
		return nil, profilebackfillapp.ErrUnavailable
	}
	schemas := map[uuid.UUID]profileapp.Schema{current.ID(): current}
	for _, row := range rows {
		if row.SchemaVersionID == current.ID() {
			continue
		}
		if _, found := schemas[row.SchemaVersionID]; found {
			continue
		}
		schema, err := reader.HistoricalSchema(ctx, row.SchemaVersionID)
		if err != nil || !validProfileBackfillActualSchema(current, schema, row.SchemaVersionID) {
			return nil, profilebackfillapp.ErrUnavailable
		}
		schemas[row.SchemaVersionID] = schema
	}
	return schemas, nil
}

func validProfileBackfillActualSchema(
	current profileapp.Schema,
	actual profileapp.Schema,
	expectedID uuid.UUID,
) bool {
	return expectedID != uuid.Nil && actual.ID() == expectedID &&
		actual.Resource() == current.Resource() &&
		actual.State() == profileapp.SchemaRetired &&
		actual.Version() > 0 && actual.Version() < current.Version()
}
