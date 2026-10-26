package compiler

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestQualifiedStaticStorageRendersSchemaAndTable(t *testing.T) {
	schema, _ := sqlidentifier.MintStatic("platform")
	table, _ := sqlidentifier.MintStatic("billing_offering_query_v1")
	storage, err := api.NewQualifiedResourceStorage(schema, table)
	if err != nil {
		t.Fatal(err)
	}
	var builder strings.Builder
	err = appendResourceStorage(&builder, storage)
	if err != nil || builder.String() != `"platform"."billing_offering_query_v1"` {
		t.Fatalf("qualified SQL = %q, %v", builder.String(), err)
	}
}
