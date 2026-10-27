package api

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestFieldCatalogDefaultSelectRequiresUniqueSelectableFields(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	tests := []struct {
		name   string
		mutate func(*FieldCatalogInput)
	}{
		{"empty", func(input *FieldCatalogInput) { input.DefaultSelect = nil }},
		{"duplicate", func(input *FieldCatalogInput) {
			input.DefaultSelect = []uuid.UUID{input.StableID, input.StableID}
		}},
		{"unknown", func(input *FieldCatalogInput) {
			input.DefaultSelect = []uuid.UUID{uuid.MustParse("60000000-0000-4000-8000-000000000099")}
		}},
		{"not selectable", func(input *FieldCatalogInput) {
			field := mustUUIDCatalogField(
				t, "60000000-0000-4000-8000-000000000001", "hidden",
				"fld_hidden", false, true, 36,
			)
			input.Fields = append(input.Fields, field)
			input.DefaultSelect = []uuid.UUID{field.id}
		}},
		{"over hard max", func(input *FieldCatalogInput) {
			for index := 1; index <= 49; index++ {
				id := fmt.Sprintf("9%07d-0000-4000-8000-%012d", index, index)
				field := mustUUIDCatalogField(
					t, id, fmt.Sprintf("default_%d", index),
					fmt.Sprintf("fld_default_%d", index), true, false, 0,
				)
				input.Fields = append(input.Fields, field)
				input.DefaultSelect = append(input.DefaultSelect, field.id)
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
				t.Fatal("invalid default select accepted")
			}
		})
	}
}

func TestFieldCatalogDefaultSelectHardMaxBoundary(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	input := catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField, nil,
	)
	for index := 1; index <= 48; index++ {
		id := fmt.Sprintf("a%07d-0000-4000-8000-%012d", index, index)
		field := mustUUIDCatalogField(
			t, id, fmt.Sprintf("boundary_%d", index),
			fmt.Sprintf("fld_boundary_%d", index), true, false, 0,
		)
		input.Fields = append(input.Fields, field)
		input.DefaultSelect = append(input.DefaultSelect, field.id)
	}
	if len(input.DefaultSelect) != 50 {
		t.Fatal("boundary fixture does not contain 50 fields")
	}
	if _, err := NewFieldCatalog(input); err != nil {
		t.Fatalf("50 default fields rejected: %v", err)
	}
	extra := mustUUIDCatalogField(
		t, "afffffff-0000-4000-8000-000000000001", "boundary_extra",
		"fld_boundary_extra", true, false, 0,
	)
	input.Fields = append(input.Fields, extra)
	input.DefaultSelect = append(input.DefaultSelect, extra.id)
	if _, err := NewFieldCatalog(input); err == nil {
		t.Fatal("51 default fields accepted")
	}
}

func TestFieldCatalogStableIDRequiresUUIDCursorSafety(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	tests := []struct {
		name   string
		mutate func(*FieldCatalogInput)
	}{
		{"unknown", func(input *FieldCatalogInput) {
			input.StableID = uuid.MustParse("60000000-0000-4000-8000-000000000099")
		}},
		{"not UUID", func(input *FieldCatalogInput) {
			input.StableID = uuid.MustParse(rootSearchField)
		}},
		{"not selectable", stableIDMutation(t, false, true, 36)},
		{"not sortable", stableIDMutation(t, true, false, 0)},
		{"short cursor bound", stableIDMutation(t, true, true, 35)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := catalogFixtureInput(
				t, root, "root_records", rootStableField, rootSearchField, nil,
			)
			test.mutate(&input)
			if _, err := NewFieldCatalog(input); err == nil {
				t.Fatal("invalid stable ID accepted")
			}
		})
	}
}

func stableIDMutation(
	t *testing.T, selectable, sortable bool, bound int,
) func(*FieldCatalogInput) {
	t.Helper()
	return func(input *FieldCatalogInput) {
		field := mustUUIDCatalogField(
			t, "60000000-0000-4000-8000-000000000002", "alternate_id",
			"fld_alternate_id", selectable, sortable, bound,
		)
		input.Fields = append(input.Fields, field)
		input.StableID = field.id
	}
}
