package storagecatalog

import "testing"

func TestWorkRESTQueryCatalogsAreExactAndComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, spec := range allQuerySpecs() {
		if seen[spec.name] {
			t.Fatalf("duplicate Work query resource %q", spec.name)
		}
		seen[spec.name] = true
		catalog, err := newCatalog(spec)
		if err != nil || catalog.Validate() != nil {
			t.Fatalf("build Work query resource %q: %v", spec.name, err)
		}
	}
	for _, resource := range []string{
		"boards", "tasks", "comments", "calendar", "inbox", "calendar-unified",
	} {
		if !seen[resource] {
			t.Fatalf("missing Work query resource %q", resource)
		}
	}
}
