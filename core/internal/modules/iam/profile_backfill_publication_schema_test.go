package iam

import (
	"context"
	"errors"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type profileHistoricalSchemaReaderStub struct {
	schema profileapp.Schema
	calls  int
}

func (reader *profileHistoricalSchemaReaderStub) HistoricalSchema(
	context.Context, uuid.UUID,
) (profileapp.Schema, error) {
	reader.calls++
	return reader.schema, nil
}

func TestProfileBackfillActualSchemasCachesImmutablePageSnapshot(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	current := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 8)
	historicalActive := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	historical, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: historicalActive.ID(), Resource: historicalActive.Resource(),
		Version: historicalActive.Version(), State: profileapp.SchemaRetired,
		Fields: historicalActive.Fields(),
	})
	if err != nil {
		t.Fatal(err)
	}
	reader := &profileHistoricalSchemaReaderStub{schema: historical}
	rows := []profilebackfillapp.SourceResource{
		{ID: uuid.MustParse("14000000-0000-4000-8000-000000000001"), SchemaVersionID: historical.ID()},
		{ID: uuid.MustParse("14000000-0000-4000-8000-000000000002"), SchemaVersionID: historical.ID()},
	}
	schemas, err := profileBackfillActualSchemas(context.Background(), reader, current, rows)
	if err != nil || len(schemas) != 2 || reader.calls != 1 ||
		schemas[historical.ID()].ID() != historical.ID() {
		t.Fatalf("actual schemas = %d/%d/%v", len(schemas), reader.calls, err)
	}
}

func TestProfileBackfillActualSchemasRejectsNonRetiredHistoricalSnapshot(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	current := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 8)
	historical := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	reader := &profileHistoricalSchemaReaderStub{schema: historical}
	_, err := profileBackfillActualSchemas(
		context.Background(), reader, current,
		[]profilebackfillapp.SourceResource{{
			ID:              uuid.MustParse("14000000-0000-4000-8000-000000000001"),
			SchemaVersionID: historical.ID(),
		}},
	)
	if !errors.Is(err, profilebackfillapp.ErrUnavailable) {
		t.Fatalf("active historical snapshot error = %v", err)
	}
}
