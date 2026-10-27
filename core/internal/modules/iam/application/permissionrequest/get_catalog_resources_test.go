package permissionrequest

import "testing"

func TestGetCatalogResourcesIsBoundedAndKindClosed(t *testing.T) {
	if err := (CatalogQuery{Kind: CatalogKindAll, Limit: 100}).Validate(); err != nil {
		t.Fatalf("valid catalog query: %v", err)
	}
	for _, query := range []CatalogQuery{
		{Kind: "resource", Limit: 10},
		{Kind: CatalogKindAll, Limit: 101},
		{Kind: CatalogKindAll, Limit: 10, Search: " untrimmed"},
	} {
		if query.Validate() == nil {
			t.Fatalf("invalid catalog query accepted: %#v", query)
		}
	}
}
