package api

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestFieldStorageCompanionCombinations(t *testing.T) {
	column, _ := sqlidentifier.MintDynamic("fld_title_g1")
	folded, _ := sqlidentifier.MintDynamic("fld_title_folded_g1")
	accent, _ := sqlidentifier.MintDynamic("fld_title_accent_g1")
	tests := []struct {
		name      string
		folded    *sqlidentifier.Token
		accent    *sqlidentifier.Token
		hasFolded bool
		hasAccent bool
	}{
		{"column only", nil, nil, false, false},
		{"folded", &folded, nil, true, false},
		{"folded and accent", &folded, &accent, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage, err := NewFieldStorage(column, test.folded, test.accent)
			if err != nil || storage.Validate() != nil {
				t.Fatalf("valid field storage rejected: %v", err)
			}
			actualColumn, columnErr := storage.Column()
			actualFolded, hasFolded, foldedErr := storage.FoldedCompanion()
			actualAccent, hasAccent, accentErr := storage.AccentCompanion()
			if columnErr != nil || foldedErr != nil || accentErr != nil ||
				actualColumn != column || hasFolded != test.hasFolded ||
				hasAccent != test.hasAccent {
				t.Fatal("field storage accessors did not preserve the contract")
			}
			if hasFolded && actualFolded != folded || hasAccent && actualAccent != accent {
				t.Fatal("field storage returned the wrong companion token")
			}
		})
	}
}

func TestFieldStorageAcceptsStaticTokensAndOwnsCopies(t *testing.T) {
	column, _ := sqlidentifier.MintStatic("forms_title")
	folded, _ := sqlidentifier.MintStatic("forms_title_folded")
	storage, err := NewFieldStorage(column, &folded, nil)
	if err != nil {
		t.Fatal(err)
	}
	actual, present, err := storage.FoldedCompanion()
	if err != nil || !present || actual != folded {
		t.Fatal("static companion token was not accepted")
	}
	actual = sqlidentifier.Token{}
	again, present, err := storage.FoldedCompanion()
	if err != nil || !present || again != folded {
		t.Fatal("field storage token was externally mutated")
	}
}

func TestFieldStorageRejectsInvalidCompanionCombinations(t *testing.T) {
	column, _ := sqlidentifier.MintDynamic("fld_title_g1")
	folded, _ := sqlidentifier.MintDynamic("fld_title_folded_g1")
	accent, _ := sqlidentifier.MintDynamic("fld_title_accent_g1")
	zero := sqlidentifier.Token{}
	for _, input := range []struct {
		column sqlidentifier.Token
		folded *sqlidentifier.Token
		accent *sqlidentifier.Token
	}{
		{zero, nil, nil}, {column, &zero, nil}, {column, nil, &accent},
		{column, &folded, &zero}, {column, &column, nil},
		{column, &folded, &column}, {column, &folded, &folded},
	} {
		if _, err := NewFieldStorage(input.column, input.folded, input.accent); err == nil {
			t.Fatal("invalid field storage accepted")
		}
	}
}

func TestFieldStorageZeroAndTamperFailClosed(t *testing.T) {
	column, _ := sqlidentifier.MintDynamic("fld_title_g1")
	folded, _ := sqlidentifier.MintDynamic("fld_title_folded_g1")
	accent, _ := sqlidentifier.MintDynamic("fld_title_accent_g1")
	for _, storage := range []FieldStorage{
		{},
		{column: column},
		{column: column, validated: true, folded: folded},
		{column: column, validated: true, hasFolded: true},
		{column: column, validated: true, accent: accent, hasAccent: true},
	} {
		if storage.Validate() == nil {
			t.Fatal("tampered field storage validated")
		}
		if _, err := storage.Column(); err == nil {
			t.Fatal("tampered field storage exposed a column")
		}
	}
}
