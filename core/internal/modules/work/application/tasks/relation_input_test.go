package tasks

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestRelationInputNormalizesTypedLegacyValues(t *testing.T) {
	value := RelationInput{Labels: []string{" Urgent "},
		Checklist: []ChecklistValue{{Text: " Review "}}}
	if err := normalizeRelations(&value); err != nil {
		t.Fatal(err)
	}
	if value.Labels[0] != "Urgent" || value.Checklist[0].Text != "Review" {
		t.Fatalf("relations were not normalized: %#v", value)
	}
}

func TestRelationInputRejectsDuplicates(t *testing.T) {
	value := RelationInput{Labels: []string{"Urgent", "Urgent"}}
	if err := normalizeRelations(&value); !errors.Is(err, ports.ErrInvalid) {
		t.Fatalf("duplicate label result = %v", err)
	}
}
