package ddl

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestCreateFormTablePinsV1SystemColumnBudget(t *testing.T) {
	if schemav1.InitialFormSystemColumnCount != 5 {
		t.Fatalf("initial system column count = %d",
			schemav1.InitialFormSystemColumnCount)
	}
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("renderer test source path is unavailable")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(current), "render_form_table.go"))
	if err != nil {
		t.Fatal(err)
	}
	columns := []string{
		"id uuid PRIMARY KEY", "created_at timestamptz NOT NULL",
		"updated_at timestamptz NOT NULL", "row_version bigint NOT NULL",
		"deleted_at timestamptz",
	}
	for _, column := range columns {
		if strings.Count(string(source), column) != 1 {
			t.Fatalf("renderer does not contain exact system column %q", column)
		}
	}
}
