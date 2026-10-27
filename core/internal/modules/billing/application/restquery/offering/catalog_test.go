package offering

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestCatalogPublishesQualifiedImmutableOfferingSurface(t *testing.T) {
	catalogs, err := NewCatalog()
	if err != nil {
		t.Fatal(err)
	}
	root, err := catalogs.set.RootCatalog()
	if err != nil {
		t.Fatal(err)
	}
	fields, fieldsErr := root.Fields()
	defaults, defaultsErr := root.DefaultSelect()
	searches, searchesErr := root.SearchFields()
	stable, stableErr := root.StableID()
	storage, storageErr := root.Storage()
	_, qualified, schemaErr := storage.Schema()
	if fieldsErr != nil || defaultsErr != nil || searchesErr != nil ||
		stableErr != nil || storageErr != nil || schemaErr != nil {
		t.Fatal("catalog access failed")
	}
	if len(fields) != len(offeringFields) || len(defaults) != len(offeringFields) ||
		len(searches) != 1 || stable != uuid.MustParse(fieldID) || !qualified {
		t.Fatalf("catalog shape fields=%d defaults=%d searches=%d", len(fields), len(defaults), len(searches))
	}
	key, found, err := root.FieldByPublicName("offering_key")
	if err != nil || !found {
		t.Fatal("offering_key is absent")
	}
	capabilities, err := key.Capabilities()
	searchable, searchErr := capabilities.Searchable()
	_, contains, containsErr := capabilities.Filter(api.OperatorContains)
	if err != nil || searchErr != nil || containsErr != nil || !searchable || !contains {
		t.Fatal("offering_key search capability is incomplete")
	}
}
