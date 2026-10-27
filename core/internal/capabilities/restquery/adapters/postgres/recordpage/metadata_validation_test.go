package recordpage

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/google/uuid"
)

func TestValidateMetadataRejectsDuplicateAndReservedOutputKeys(t *testing.T) {
	metadata, sorts := metadataFixture(t)
	if err := validateMetadata(metadata, sorts); err != nil {
		t.Fatalf("valid metadata rejected: %v", err)
	}
	duplicate := metadata
	duplicate.columns = append([]columnMetadata(nil), metadata.columns...)
	duplicate.columns[1].outputKey = duplicate.columns[0].outputKey
	if validateMetadata(duplicate, sorts) != ErrContract {
		t.Fatal("duplicate output key was accepted")
	}
	reserved := metadata
	reserved.columns = append([]columnMetadata(nil), metadata.columns...)
	reserved.columns[0].outputKey = "_included"
	if validateMetadata(reserved, sorts) != ErrContract {
		t.Fatal("reserved output key was accepted")
	}
}

func TestValidateMetadataRejectsSortMismatch(t *testing.T) {
	metadata, sorts := metadataFixture(t)
	tests := map[string]func(*statementMetadata){
		"missing": func(value *statementMetadata) {
			value.columns[0].sorted = false
			value.columns[0].sortOrdinal = -1
		},
		"duplicate": func(value *statementMetadata) {
			value.columns[1].sorted = true
			value.columns[1].sortOrdinal = 0
		},
		"wrong type": func(value *statementMetadata) {
			value.columns[0].kind = api.LogicalTypeText
		},
		"direction transplant": func(value *statementMetadata) {
			value.sorts[0].direction = api.DirectionDescending
		},
		"bound transplant": func(value *statementMetadata) {
			value.sorts[0].canonicalBound++
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			changed := metadata
			changed.columns = append([]columnMetadata(nil), metadata.columns...)
			changed.sorts = append([]sortMetadata(nil), metadata.sorts...)
			mutate(&changed)
			if validateMetadata(changed, sorts) != ErrContract {
				t.Fatal("invalid sort metadata was accepted")
			}
		})
	}
}

func metadataFixture(t *testing.T) (statementMetadata, []cursor.SortDefinition) {
	t.Helper()
	fieldType, err := api.NewUUIDFieldType()
	if err != nil {
		t.Fatal(err)
	}
	first := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	second := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	sort, err := cursor.NewSortDefinition(
		first.String(), api.DirectionAscending, api.LogicalTypeUUID, 36,
	)
	if err != nil {
		t.Fatal(err)
	}
	metadata := statementMetadata{
		requestedSize: 2,
		sorts: []sortMetadata{{
			fieldID: first, direction: api.DirectionAscending,
			kind: api.LogicalTypeUUID, canonicalBound: 36,
		}},
		columns: []columnMetadata{
			{fieldID: first, fieldType: fieldType, kind: api.LogicalTypeUUID,
				outputKey: "id", hasOutput: true, sortOrdinal: 0, sorted: true},
			{fieldID: second, fieldType: fieldType, kind: api.LogicalTypeUUID,
				outputKey: "owner_id", hasOutput: true, sortOrdinal: -1},
		},
	}
	return metadata, []cursor.SortDefinition{sort}
}
