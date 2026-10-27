package permissionrequest

import "testing"

func TestReconcileMissingCatalogResourcesRequiresActionAndBoundedPage(t *testing.T) {
	valid := MissingCatalogQuery{
		CatalogQuery: CatalogQuery{Kind: CatalogKindForm, Limit: 50},
		ActionKey:    "read",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid missing catalog query: %v", err)
	}
	valid.ActionKey = "*"
	if valid.Validate() == nil {
		t.Fatal("wildcard action was accepted")
	}
}
