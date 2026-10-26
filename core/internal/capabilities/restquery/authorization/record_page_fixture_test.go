package authorization

import (
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/google/uuid"
)

func recordPageStatement(
	t *testing.T,
	explicitNameSort bool,
) (compiler.RecordStatement, cursor.Codec) {
	t.Helper()
	base := snapshotRecordSpec(t, 1)
	catalogs, err := base.Catalogs()
	if err != nil {
		t.Fatal(err)
	}
	page, err := api.NewRequestPage(2, nil)
	if err != nil {
		t.Fatal(err)
	}
	var sorts []api.RequestSort
	if explicitNameSort {
		sortValue, sortErr := api.NewRequestSort("name", api.DirectionAscending)
		if sortErr != nil {
			t.Fatal(sortErr)
		}
		sorts = []api.RequestSort{sortValue}
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Sort: sorts, Page: page, Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	codec := compilerBoundaryCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	statement, err := postgresCompiler.CompileRecord(compilerBoundaryPlan(t, spec))
	if err != nil {
		t.Fatal(err)
	}
	return statement, codec
}

func recordPageRow(
	t *testing.T,
	statement compiler.RecordStatement,
	name any,
	identifier any,
) []rowdecode.Value {
	t.Helper()
	columns, err := statement.Columns()
	if err != nil {
		t.Fatal(err)
	}
	row := make([]rowdecode.Value, len(columns))
	for index, column := range columns {
		fieldType, typeErr := column.FieldType()
		kind, kindErr := column.LogicalType()
		if typeErr != nil || kindErr != nil {
			t.Fatal("column metadata unavailable")
		}
		driverValue := identifier
		if kind == api.LogicalTypeText {
			driverValue = name
		} else if kind != api.LogicalTypeUUID {
			t.Fatalf("unexpected row kind %s", kind)
		}
		row[index], err = rowdecode.Decode(fieldType, driverValue)
		if err != nil {
			t.Fatal(err)
		}
	}
	return row
}

func recordPageUUID(index int) [16]byte {
	identifier := uuid.MustParse(fmt.Sprintf(
		"40000000-0000-4000-8000-%012x", index,
	))
	return [16]byte(identifier)
}
