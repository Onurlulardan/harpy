//go:build integration

package executor

import (
	"crypto/sha256"
	"math"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	pg18AggregateResource = "61000000-0000-4000-8000-000000000001"
	pg18AggregateVersion  = "61000000-0000-4000-8000-000000000002"
	pg18AggregateID       = "61000000-0000-4000-8000-000000000003"
	pg18AggregateStatus   = "61000000-0000-4000-8000-000000000004"
	pg18AggregateQuantity = "61000000-0000-4000-8000-000000000005"
	pg18AggregateAmount   = "61000000-0000-4000-8000-000000000006"
)

type pg18AggregatePlanInput struct {
	groups   []string
	measures []api.RequestMeasure
	where    *api.RequestFilter
	having   *api.RequestHaving
	sorts    []api.RequestAggregateSort
	limit    int
}

func pg18AggregatePlan(
	t *testing.T, input pg18AggregatePlanInput,
) resolved.AuthorizedPlan {
	t.Helper()
	aggregate, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy: input.groups, Measures: input.measures, Having: input.having,
		Sort: input.sorts, Limit: &input.limit,
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Where: input.where, Aggregate: aggregate,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, pg18AggregateCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	reference, _ := spec.Reference()
	path, _ := spec.RootPath()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("pg18-aggregate-requirement")), reference, path,
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse(pg18TenantA), uuid.MustParse(pg18Actor),
		sha256.Sum256([]byte("pg18-aggregate-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func pg18AggregateCatalog(t *testing.T) api.CatalogSet {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("e", 64))
	reference, _ := api.NewCatalogReference(
		"forms", pg18AggregateResource, pg18AggregateVersion, hash,
	)
	idType, _ := api.NewUUIDFieldType()
	textProfile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	textType, _ := api.NewTextFieldType(textProfile)
	integerProfile, _ := api.NewIntegerProfile(math.MinInt64, math.MaxInt64)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	decimalProfile, _ := api.NewDecimalProfile(12, 6)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	fields := []api.FieldDefinition{
		fixtureField(t, pg18AggregateID, "id", "id", idType,
			api.FieldCapabilitiesInput{Selectable: true, Sortable: true,
				CursorCanonicalBytes: 36}),
		fixtureField(t, pg18AggregateStatus, "status", "status", textType,
			api.FieldCapabilitiesInput{Selectable: true, Groupable: true,
				Filters: []api.FilterCapabilityInput{{
					Operator: api.OperatorEq, CountSafe: true,
				}}}),
		fixtureField(t, pg18AggregateQuantity, "quantity", "quantity", integerType,
			api.FieldCapabilitiesInput{Selectable: true,
				Aggregates: []api.AggregateFunction{api.AggregateFunctionSum}}),
		fixtureField(t, pg18AggregateAmount, "amount", "amount", decimalType,
			api.FieldCapabilitiesInput{Selectable: true,
				Aggregates: []api.AggregateFunction{api.AggregateFunctionAvg}}),
	}
	table, _ := sqlidentifier.MintDynamic("frm_executor_aggregate")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, DefaultSelect: []uuid.UUID{uuid.MustParse(pg18AggregateID)},
		StableID: uuid.MustParse(pg18AggregateID),
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

func pg18StatusFilter(t *testing.T, status string) *api.RequestFilter {
	t.Helper()
	literal, _ := api.NewTextLiteral(status)
	filter, err := api.NewRequestPredicate(
		"status", api.OperatorEq, []api.Literal{literal},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &filter
}
