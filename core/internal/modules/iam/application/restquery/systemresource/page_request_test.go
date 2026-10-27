package systemresourcequery

import (
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestNewPageRequestPreservesQueryAndPinsProjectionCursor(t *testing.T) {
	base, err := DecodePOST([]byte(`{
		"where":{"field":"status","op":"eq","value":"active"},
		"sort":[{"field":"id","direction":"asc"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	after := "signed-cursor"
	paged, err := NewPageRequest(base, []string{"id", "status"}, 500, &after)
	if err != nil {
		t.Fatal(err)
	}
	value, err := paged.Value()
	actualAfter, present := value.Page().After()
	if err != nil || value.Count() != restquery.CountModeNone ||
		value.Page().Size() != 500 || !present || actualAfter != after ||
		len(value.Select()) != 2 || len(value.Sort()) != 1 {
		t.Fatalf("paged request drifted: err=%v", err)
	}
	if _, present := value.Where(); !present {
		t.Fatal("paged request lost its filter")
	}
}
