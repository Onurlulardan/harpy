package migrationmanifest

import "testing"

func TestSourceDocumentRequiresGroupedDollarQuotedSQL(t *testing.T) {
	path := "core/internal/example/migrations/fresh/20260722010000_example.fresh.sql"
	tests := []struct {
		name    string
		content string
		valid   bool
	}{
		{"grouped empty tag", groupedFunction("$$"), true},
		{"grouped named tag", groupedFunction("$function$"), true},
		{"ungrouped empty tag", ungroupedFunction("$$"), false},
		{"ungrouped named tag", ungroupedFunction("$function$"), false},
		{"partially grouped", partiallyGroupedFunctions(), false},
		{"later statements preserved", groupedFunction("$$") +
			"CREATE TRIGGER example_guard BEFORE UPDATE ON example " +
			"EXECUTE FUNCTION example();\n" +
			"REVOKE ALL ON FUNCTION example() FROM PUBLIC;\n", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewSourceDocument("example", path, []byte(test.content))
			if test.valid && err != nil {
				t.Fatalf("valid dollar-quoted SQL was rejected: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("invalid dollar-quoted SQL was accepted")
			}
		})
	}
}

func TestSourceDocumentRejectsUnbalancedStatementGroups(t *testing.T) {
	path := "core/internal/example/migrations/fresh/20260722010000_example.fresh.sql"
	tests := map[string]string{
		"unclosed":      "-- +goose Up\n-- +goose StatementBegin\nSELECT 1;\n",
		"unmatched end": "-- +goose Up\n-- +goose StatementEnd\n",
		"nested": "-- +goose Up\n-- +goose StatementBegin\n" +
			"-- +goose StatementBegin\nSELECT 1;\n",
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSourceDocument("example", path, []byte(content)); err == nil {
				t.Fatal("unbalanced goose statement group was accepted")
			}
		})
	}
}

func groupedFunction(tag string) string {
	return "-- +goose Up\n-- +goose StatementBegin\n" +
		functionSQL(tag) + "-- +goose StatementEnd\n"
}

func ungroupedFunction(tag string) string {
	return "-- +goose Up\n" + functionSQL(tag)
}

func partiallyGroupedFunctions() string {
	return groupedFunction("$$") + functionSQL("$function$")
}

func functionSQL(tag string) string {
	return "CREATE FUNCTION example() RETURNS void AS " + tag +
		" BEGIN NULL; END; " + tag + " LANGUAGE plpgsql;\n"
}
