package postgres

import (
	"strings"
	"testing"
)

func TestLayoutStaticSQLCallsOnlyOwnedRoutines(t *testing.T) {
	for name, statement := range map[string]string{
		"create": createSQL, "get": getSQL, "list": listSQL,
		"update": updateSQL, "delete": deleteSQL,
		"publish": publishSQL, "resolve": resolveSQL,
	} {
		version := "_v1"
		if name == "update" {
			version = "_v2"
		}
		if name == "create" {
			version = "_v3"
		}
		if name == "publish" {
			version = "_v5"
		}
		if name == "resolve" {
			version = "_v2"
		}
		if !strings.Contains(statement, "system_resource_layout_"+name+version) ||
			!strings.Contains(statement,
				"(pg_catalog.current_schemas(false))[2]::name") ||
			strings.Contains(statement, "pg_catalog.current_schema()") ||
			strings.Contains(statement, "iam_") || strings.Contains(statement, " JOIN ") {
			t.Fatalf("%s static SQL crossed its owner boundary: %s", name, statement)
		}
	}
}
