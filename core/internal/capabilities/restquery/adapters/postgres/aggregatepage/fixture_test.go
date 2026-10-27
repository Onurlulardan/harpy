package aggregatepage

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	pageResource = "41000000-0000-4000-8000-000000000001"
	pageVersion  = "41000000-0000-4000-8000-000000000002"
	pageID       = "41000000-0000-4000-8000-000000000003"
	pageStatus   = "41000000-0000-4000-8000-000000000004"
	pageAmount   = "41000000-0000-4000-8000-000000000005"
)

type captureAggregate struct{ statement compiler.AggregateStatement }

func (capture *captureAggregate) ConsumeAggregate(
	statement compiler.AggregateStatement,
) error {
	capture.statement = statement
	return nil
}

func fixtureAggregateStatement(t testing.TB) compiler.AggregateStatement {
	t.Helper()
	catalogs := fixtureAggregateCatalog(t)
	amount := "amount"
	count, _ := api.NewRequestMeasure("row_count", api.AggregateFunctionCount, nil)
	sum, _ := api.NewRequestMeasure("revenue", api.AggregateFunctionSum, &amount)
	limit := 2
	aggregate, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy: []string{"status"}, Measures: []api.RequestMeasure{count, sum},
		Limit: &limit,
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
		sha256.Sum256([]byte("aggregate-page-requirement")), reference, path,
	)
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse(pageResource), uuid.MustParse(pageVersion),
		sha256.Sum256([]byte("aggregate-page-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	postgresCompiler, err := compiler.New(fixtureAggregateCodec(t))
	if err != nil {
		t.Fatal(err)
	}
	batch, err := postgresCompiler.CompileAggregateBatch(plan)
	if err != nil {
		t.Fatal(err)
	}
	capture := &captureAggregate{}
	if err := compiler.ConsumeAggregateBatch(batch, plan, capture); err != nil {
		t.Fatal(err)
	}
	return capture.statement
}

func fixtureAggregateCatalog(t testing.TB) api.CatalogSet {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("c", 64))
	reference, _ := api.NewCatalogReference("forms", pageResource, pageVersion, hash)
	idType, _ := api.NewUUIDFieldType()
	textProfile, _ := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	textType, _ := api.NewTextFieldType(textProfile)
	decimalProfile, _ := api.NewDecimalProfile(12, 2)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	fields := []api.FieldDefinition{
		fixtureAggregateField(t, pageID, "id", idType, api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
		}),
		fixtureAggregateField(t, pageStatus, "status", textType,
			api.FieldCapabilitiesInput{Selectable: true, Groupable: true}),
		fixtureAggregateField(t, pageAmount, "amount", decimalType,
			api.FieldCapabilitiesInput{Selectable: true,
				Aggregates: []api.AggregateFunction{api.AggregateFunctionSum}}),
	}
	table, _ := sqlidentifier.MintDynamic("frm_aggregate_page")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, DefaultSelect: []uuid.UUID{uuid.MustParse(pageID)},
		StableID: uuid.MustParse(pageID),
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

func fixtureAggregateField(
	t testing.TB, id, name string, fieldType api.FieldType,
	input api.FieldCapabilitiesInput,
) api.FieldDefinition {
	t.Helper()
	column, err := sqlidentifier.MintDynamic("fld_" + name)
	if err != nil {
		t.Fatalf("column %s: %v", name, err)
	}
	storage, err := api.NewFieldStorage(column, nil, nil)
	if err != nil {
		t.Fatalf("storage %s: %v", name, err)
	}
	capabilities, err := api.NewFieldCapabilities(fieldType, input)
	if err != nil {
		t.Fatalf("capabilities %s: %v", name, err)
	}
	field, err := api.NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func fixtureAggregateCodec(t testing.TB) cursor.Codec {
	t.Helper()
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	}
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},`+
			`"previous":{"kid":"previous-1","key":"%s"}}`, key(1), key(2),
	)
	t.Setenv("HARPY_ENVIRONMENT", "development")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN",
		"postgresql://user@example.invalid/db?sslmode=disable")
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", document)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.NewCodec(keyring)
	if err != nil {
		t.Fatal(err)
	}
	return codec
}
