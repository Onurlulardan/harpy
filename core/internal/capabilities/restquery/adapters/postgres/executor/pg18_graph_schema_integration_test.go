//go:build integration

package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func pg18PrepareGraphSchema(t *testing.T, gate *pg18Gate) {
	t.Helper()
	statements := []string{"SET ROLE forgecrud_owner"}
	for _, schema := range []string{"tenant_a", "tenant_b"} {
		statements = append(statements, pg18GraphRecordDDL(schema)...)
		statements = append(statements,
			pg18GraphRelationDDL(schema, graphRelationA, "a", false),
			pg18GraphRelationDDL(schema, graphRelationB, "b", true),
			pg18GraphRelationDDL(schema, graphRelationC, "c", false),
		)
	}
	statements = append(statements, "RESET ROLE",
		"GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA tenant_a,tenant_b TO forgecrud_runtime",
	)
	pg18ExecuteSetup(t, gate.ctx, gate.admin, statements)
	for _, tenant := range []string{pg18TenantA, pg18TenantB} {
		if _, err := gate.admin.Exec(gate.ctx, `INSERT INTO platform.domain_cutovers
VALUES('tenant',$1,'iam','v2','v2',1)`, tenant); err != nil {
			t.Fatal(err)
		}
	}
}

func pg18GraphRecordDDL(schema string) []string {
	definitions := []struct{ table, id, label string }{
		{"frm_graph_root", "fld_graph_root_id", "fld_graph_root_label"},
		{"frm_graph_a", "fld_graph_a_id", "fld_graph_a_label"},
		{"frm_graph_b", "fld_graph_b_id", "fld_graph_b_label"},
		{"frm_graph_c", "fld_graph_c_id", "fld_graph_c_label"},
	}
	result := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, fmt.Sprintf(
			"CREATE TABLE %s(%s uuid PRIMARY KEY,%s text NOT NULL)",
			pg18GraphQualified(schema, definition.table),
			pgx.Identifier{definition.id}.Sanitize(),
			pgx.Identifier{definition.label}.Sanitize(),
		))
	}
	return result
}

func pg18GraphRelationDDL(
	schema string, relationID string, suffix string, toOne bool,
) string {
	table := pg18GraphQualified(schema, pg18GraphRelationTable(relationID))
	check := "edge_slot BETWEEN 1 AND 100"
	unique := "UNIQUE(source_id,edge_slot)"
	if toOne {
		check, unique = "edge_slot = 1", "UNIQUE(source_id)"
	}
	return fmt.Sprintf(`CREATE TABLE %s(
source_id uuid NOT NULL,target_id uuid NOT NULL,edge_slot smallint NOT NULL,
CONSTRAINT graph_rel_%s_pk PRIMARY KEY(source_id,target_id),
CONSTRAINT graph_rel_%s_slot CHECK(%s),
CONSTRAINT graph_rel_%s_source %s)`,
		table, suffix, suffix, check, suffix, unique,
	)
}

func pg18GraphRelationTable(relationID string) string {
	name, err := physicalid.Name("rel_", uuid.MustParse(relationID))
	if err != nil {
		panic(err)
	}
	return name
}

func pg18GraphQualified(schema string, table string) string {
	return pgx.Identifier{schema, table}.Sanitize()
}

func pg18ResetGraphTenant(
	t *testing.T, ctx context.Context, admin *pgx.Conn, schema string,
) {
	t.Helper()
	tables := []string{
		"frm_graph_root", "frm_graph_a", "frm_graph_b", "frm_graph_c",
		pg18GraphRelationTable(graphRelationA),
		pg18GraphRelationTable(graphRelationB),
		pg18GraphRelationTable(graphRelationC),
	}
	qualified := make([]string, len(tables))
	for index, table := range tables {
		qualified[index] = pg18GraphQualified(schema, table)
	}
	statement := "TRUNCATE TABLE "
	for index, table := range qualified {
		if index > 0 {
			statement += ","
		}
		statement += table
	}
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatal(err)
	}
}
