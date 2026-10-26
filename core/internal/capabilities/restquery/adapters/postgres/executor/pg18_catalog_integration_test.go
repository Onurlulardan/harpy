//go:build integration

package executor

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

type pg18PlanInput struct {
	tenant string
	owner  string
	after  *string
	where  *api.RequestFilter
	count  api.CountMode
	size   int
}

func pg18AuthorizedPlan(t *testing.T, input pg18PlanInput) resolved.AuthorizedPlan {
	t.Helper()
	catalogs := pg18CatalogSet(t, input.owner)
	page, err := api.NewRequestPage(input.size, input.after)
	if err != nil {
		t.Fatal(err)
	}
	sort, err := api.NewRequestSort("id", api.DirectionAscending)
	if err != nil {
		t.Fatal(err)
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"id", "name", "metadata"}, Where: input.where,
		Sort: []api.RequestSort{sort}, Page: page, Count: input.count,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	reference, _ := spec.Reference()
	path, _ := spec.RootPath()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("pg18-restquery-requirement")), reference, path,
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse(input.tenant), uuid.MustParse(pg18Actor),
		sha256.Sum256([]byte("pg18-restquery-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func pg18CatalogSet(t *testing.T, owner string) api.CatalogSet {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("d", 64))
	reference, err := api.NewCatalogReference(
		owner, fixtureResource, fixtureVersion, hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	idType, _ := api.NewUUIDFieldType()
	profile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	nameType, _ := api.NewTextFieldType(profile)
	jsonType, _ := api.NewJSONFieldType()
	fields := []api.FieldDefinition{
		fixtureField(t, fixtureID, "id", "id", idType,
			api.FieldCapabilitiesInput{
				Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}},
			}),
		fixtureField(t, fixtureName, "name", "name", nameType,
			api.FieldCapabilitiesInput{
				Selectable: true, Sortable: true, CursorCanonicalBytes: 512,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}},
			}),
		fixtureField(t, fixtureJSON, "metadata", "metadata", jsonType,
			api.FieldCapabilitiesInput{Selectable: true}),
	}
	table, _ := sqlidentifier.MintDynamic("frm_executor_fixture")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, DefaultSelect: []uuid.UUID{
			uuid.MustParse(fixtureID), uuid.MustParse(fixtureName),
			uuid.MustParse(fixtureJSON),
		}, StableID: uuid.MustParse(fixtureID),
	})
	if err != nil {
		t.Fatal(err)
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{catalog})
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func pg18NameFilter(t *testing.T, value string) *api.RequestFilter {
	t.Helper()
	literal, err := api.NewTextLiteral(value)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := api.NewRequestPredicate(
		"name", api.OperatorEq, []api.Literal{literal},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &filter
}
