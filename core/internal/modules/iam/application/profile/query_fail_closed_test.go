package profile

import (
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestQueryTypesFailClosedFromZeroValues(t *testing.T) {
	var field QueryField
	if field.Validate() == nil {
		t.Fatal("zero query field is valid")
	}
	if selectable, err := field.Selectable(); err == nil || selectable {
		t.Fatal("zero query field became selectable")
	}
	if field.RequireOperator(restquery.OperatorEq) == nil ||
		field.RequireSort() == nil || field.RequireSearch() == nil {
		t.Fatal("zero query field opened a capability")
	}
	var catalog QueryCatalog
	if catalog.Validate() == nil {
		t.Fatal("zero query catalog is valid")
	}
	if _, err := catalog.Field(fieldOneID); err == nil {
		t.Fatal("zero query catalog resolved a field")
	}
}

func TestQueryCatalogHashChangesWhenFieldIsRetired(t *testing.T) {
	active := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	retiredInput := baseFieldInput(fieldOneID, "nickname", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	activeSchema := mustSchema(t, schemaOneID, 1, active)
	retiredSchema := mustSchema(t, schemaOneID, 1, retired)
	if activeSchema.DefinitionSHA256() == retiredSchema.DefinitionSHA256() {
		t.Fatal("historical schema definition ignored its field lifecycle snapshot")
	}
	activeCatalog, err := NewQueryCatalog(activeSchema)
	if err != nil {
		t.Fatal(err)
	}
	retiredCatalog, err := NewQueryCatalog(retiredSchema)
	if err != nil {
		t.Fatal(err)
	}
	activeHash, _ := activeCatalog.Hash()
	retiredHash, _ := retiredCatalog.Hash()
	if activeHash == retiredHash {
		t.Fatal("query catalog hash ignored field retirement")
	}
}

func TestQueryCatalogLookupDoesNotGuessNames(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	catalog, _ := NewQueryCatalog(mustSchema(t, schemaOneID, 1, field))
	if _, err := catalog.FieldByPublicName("nickname"); err != nil {
		t.Fatalf("exact public name: %v", err)
	}
	for _, name := range []string{"Nickname", " nickname", "nickname "} {
		if _, err := catalog.FieldByPublicName(name); err == nil {
			t.Fatalf("non-exact public name %q was accepted", name)
		}
	}
}
