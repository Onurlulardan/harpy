package catalog

import (
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestOrganizationOwnerFieldIsTypedSelectableAndNonDefault(t *testing.T) {
	published := buildFixture(t, profile.ResourceOrganization, nil)
	set, err := published.CatalogSet()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := set.RootCatalog()
	if err != nil {
		t.Fatal(err)
	}
	ownerID := uuid.MustParse(organizationOwnerIDsFieldID)
	field, found, err := catalog.Field(ownerID)
	if err != nil || !found {
		t.Fatal("owner_ids field missing")
	}
	name, nameErr := field.PublicName()
	fieldType, typeErr := field.Type()
	kind, kindErr := fieldType.Kind()
	capabilities, capabilitiesErr := field.Capabilities()
	selectable, selectableErr := capabilities.Selectable()
	_, containsAny, containsAnyErr := capabilities.Filter(
		restquery.OperatorContainsAny,
	)
	_, containsAll, containsAllErr := capabilities.Filter(
		restquery.OperatorContainsAll,
	)
	if nameErr != nil || typeErr != nil || kindErr != nil ||
		capabilitiesErr != nil || selectableErr != nil ||
		containsAnyErr != nil || containsAllErr != nil ||
		name != "owner_ids" || kind != restquery.LogicalTypeUUIDList ||
		!selectable || !containsAny || !containsAll {
		t.Fatal("owner_ids field contract is incomplete")
	}
	defaults, err := catalog.DefaultSelect()
	if err != nil {
		t.Fatal(err)
	}
	for _, fieldID := range defaults {
		if fieldID == ownerID {
			t.Fatal("owner_ids must not expose raw identifiers by default")
		}
	}
	resource, err := published.Resource()
	if err != nil || resource.Table() != "iam_rest_query_organizations_v3" {
		t.Fatal("organization catalog does not use owner-aware query view")
	}
}
