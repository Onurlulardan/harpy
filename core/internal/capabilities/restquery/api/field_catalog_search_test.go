package api

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestFieldCatalogSearchRequiresUniqueSelectableSearchableFields(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	tests := []struct {
		name   string
		mutate func(*FieldCatalogInput)
	}{
		{"duplicate", func(input *FieldCatalogInput) {
			input.SearchFields = append(input.SearchFields, input.SearchFields[0])
		}},
		{"unknown", func(input *FieldCatalogInput) {
			input.SearchFields = []uuid.UUID{uuid.MustParse("60000000-0000-4000-8000-000000000099")}
		}},
		{"not selectable", func(input *FieldCatalogInput) {
			addSearchProfileField(t, input, false, true, 1)
		}},
		{"not searchable", func(input *FieldCatalogInput) {
			addSearchProfileField(t, input, true, false, 2)
		}},
		{"over hard max", func(input *FieldCatalogInput) {
			for index := 1; index <= 16; index++ {
				addSearchProfileField(t, input, true, true, index)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := catalogFixtureInput(
				t, root, "root_records", rootStableField, rootSearchField, nil,
			)
			test.mutate(&input)
			if _, err := NewFieldCatalog(input); err == nil {
				t.Fatal("invalid search profile accepted")
			}
		})
	}
}

func TestFieldCatalogSearchHardMaxBoundary(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	input := catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField, nil,
	)
	for index := 1; index <= 15; index++ {
		addSearchProfileField(t, &input, true, true, index)
	}
	if len(input.SearchFields) != 16 {
		t.Fatal("boundary fixture does not contain 16 search fields")
	}
	if _, err := NewFieldCatalog(input); err != nil {
		t.Fatalf("16 search fields rejected: %v", err)
	}
	addSearchProfileField(t, &input, true, true, 16)
	if _, err := NewFieldCatalog(input); err == nil {
		t.Fatal("17 search fields accepted")
	}
}

func addSearchProfileField(
	t *testing.T,
	input *FieldCatalogInput,
	selectable bool,
	searchable bool,
	index int,
) {
	t.Helper()
	id := fmt.Sprintf("7%07d-0000-4000-8000-%012d", index, index)
	field := mustSearchCatalogField(
		t, id, fmt.Sprintf("extra_%d", index), fmt.Sprintf("fld_extra_%d", index),
		fmt.Sprintf("fld_extra_%d_folded", index), selectable, searchable,
	)
	input.Fields = append(input.Fields, field)
	input.SearchFields = append(input.SearchFields, field.id)
}
