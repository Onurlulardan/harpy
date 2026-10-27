package executor

import (
	"crypto/sha256"
	"math"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	aggregateStatus   = "21000000-0000-4000-8000-000000000001"
	aggregateAmount   = "21000000-0000-4000-8000-000000000002"
	aggregateQuantity = "21000000-0000-4000-8000-000000000003"
)

type aggregateFixture struct {
	batch compiler.AggregateBatch
	plan  resolved.AuthorizedPlan
	codec cursor.Codec
}

func newAggregateFixture(t *testing.T, grouped bool) aggregateFixture {
	t.Helper()
	catalogs := aggregateCatalogSet(t)
	amount := "amount"
	count, _ := api.NewRequestMeasure("row_count", api.AggregateFunctionCount, nil)
	sum, _ := api.NewRequestMeasure("revenue", api.AggregateFunctionSum, &amount)
	groups := []string{}
	if grouped {
		groups = []string{"status"}
	}
	aggregate, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy: groups, Measures: []api.RequestMeasure{count, sum},
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Aggregate: aggregate,
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
	visibility, _ := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("aggregate-executor-requirement")), reference, path,
	)
	plan, err := resolved.MintAuthorizedPlan(
		spec, fixtureTenant, fixtureActor,
		sha256.Sum256([]byte("aggregate-executor-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	codec := fixtureCodec(t)
	postgresCompiler, _ := compiler.New(codec)
	batch, err := postgresCompiler.CompileAggregateBatch(plan)
	if err != nil {
		t.Fatal(err)
	}
	return aggregateFixture{batch: batch, plan: plan, codec: codec}
}

func aggregateCatalogSet(t *testing.T) api.CatalogSet {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("b", 64))
	reference, _ := api.NewCatalogReference(
		"forms", fixtureResource, fixtureVersion, hash,
	)
	idType, _ := api.NewUUIDFieldType()
	textProfile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	textType, _ := api.NewTextFieldType(textProfile)
	decimalProfile, _ := api.NewDecimalProfile(12, 2)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	integerProfile, _ := api.NewIntegerProfile(math.MinInt64, math.MaxInt64)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	fields := []api.FieldDefinition{
		fixtureField(t, fixtureID, "id", "id", idType, api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
		}),
		fixtureField(t, aggregateStatus, "status", "status", textType,
			api.FieldCapabilitiesInput{Selectable: true, Groupable: true}),
		fixtureField(t, aggregateAmount, "amount", "amount", decimalType,
			api.FieldCapabilitiesInput{Selectable: true, Aggregates: []api.AggregateFunction{
				api.AggregateFunctionSum, api.AggregateFunctionAvg,
			}}),
		fixtureField(t, aggregateQuantity, "quantity", "quantity", integerType,
			api.FieldCapabilitiesInput{Selectable: true, Aggregates: []api.AggregateFunction{
				api.AggregateFunctionSum, api.AggregateFunctionAvg,
			}}),
	}
	table, _ := sqlidentifier.MintDynamic("frm_executor_aggregate")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, DefaultSelect: []uuid.UUID{uuid.MustParse(fixtureID)},
		StableID: uuid.MustParse(fixtureID),
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

func consumedAggregateFixture(t *testing.T, fixture aggregateFixture) consumedAggregate {
	t.Helper()
	consumed, err := consumeAggregateBatch(fixture.batch, fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	return consumed
}
