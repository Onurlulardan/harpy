//go:build integration

package executor

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	pg18GraphRootOne   = "81000000-0000-4000-8000-000000000001"
	pg18GraphRootTwo   = "81000000-0000-4000-8000-000000000002"
	pg18GraphAOne      = "81000000-0000-4000-8000-000000000011"
	pg18GraphATwo      = "81000000-0000-4000-8000-000000000012"
	pg18GraphAMissing  = "81000000-0000-4000-8000-000000000019"
	pg18GraphASnapshot = "81000000-0000-4000-8000-000000000013"
	pg18GraphBOne      = "81000000-0000-4000-8000-000000000021"
	pg18GraphBMissing  = "81000000-0000-4000-8000-000000000029"
	pg18GraphCOne      = "81000000-0000-4000-8000-000000000031"
	pg18GraphCTwo      = "81000000-0000-4000-8000-000000000032"
	pg18GraphBRoot     = "82000000-0000-4000-8000-000000000001"
	pg18GraphBA        = "82000000-0000-4000-8000-000000000011"
	pg18GraphBB        = "82000000-0000-4000-8000-000000000021"
	pg18GraphBC        = "82000000-0000-4000-8000-000000000031"
)

type pg18GraphSeedStatement struct {
	sql  string
	args []any
}

func pg18SeedGraphHappyA(t *testing.T, gate *pg18Gate) {
	t.Helper()
	schema := "tenant_a"
	pg18ExecGraphSeed(t, gate.ctx, gate.admin, []pg18GraphSeedStatement{
		{`INSERT INTO tenant_a.frm_graph_root VALUES
($1,'Root One'),($2,'Root Two')`, []any{pg18GraphRootOne, pg18GraphRootTwo}},
		{`INSERT INTO tenant_a.frm_graph_a VALUES
($1,'Zulu child'),($2,'Alpha child')`, []any{pg18GraphAOne, pg18GraphATwo}},
		{`INSERT INTO tenant_a.frm_graph_b VALUES($1,'Visible owner')`, []any{pg18GraphBOne}},
		{`INSERT INTO tenant_a.frm_graph_c VALUES
($1,'Zulu nested'),($2,'Alpha nested')`, []any{pg18GraphCOne, pg18GraphCTwo}},
		{`INSERT INTO ` + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationA)) + ` VALUES
($1,$2,1),($1,$3,2),($1,$4,3),($5,$4,1)`,
			[]any{pg18GraphRootOne, pg18GraphAOne, pg18GraphATwo,
				pg18GraphAMissing, pg18GraphRootTwo}},
		{`INSERT INTO ` + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationB)) + ` VALUES
($1,$2,1),($3,$4,1)`, []any{
			pg18GraphRootOne, pg18GraphBOne, pg18GraphRootTwo, pg18GraphBMissing,
		}},
		{`INSERT INTO ` + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationC)) + ` VALUES
($1,$2,1),($1,$3,2)`, []any{pg18GraphATwo, pg18GraphCOne, pg18GraphCTwo}},
	})
}

func pg18SeedGraphHappyB(t *testing.T, gate *pg18Gate) {
	t.Helper()
	schema := "tenant_b"
	pg18ExecGraphSeed(t, gate.ctx, gate.admin, []pg18GraphSeedStatement{
		{"INSERT INTO tenant_b.frm_graph_root VALUES($1,'Tenant B Root')", []any{pg18GraphBRoot}},
		{"INSERT INTO tenant_b.frm_graph_a VALUES($1,'Tenant B Child')", []any{pg18GraphBA}},
		{"INSERT INTO tenant_b.frm_graph_b VALUES($1,'Tenant B Owner')", []any{pg18GraphBB}},
		{"INSERT INTO tenant_b.frm_graph_c VALUES($1,'Tenant B Nested')", []any{pg18GraphBC}},
		{"INSERT INTO " + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationA)) + " VALUES($1,$2,1)", []any{pg18GraphBRoot, pg18GraphBA}},
		{"INSERT INTO " + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationB)) + " VALUES($1,$2,1)", []any{pg18GraphBRoot, pg18GraphBB}},
		{"INSERT INTO " + pg18GraphQualified(schema, pg18GraphRelationTable(graphRelationC)) + " VALUES($1,$2,1)", []any{pg18GraphBA, pg18GraphBC}},
	})
}

func pg18ExecGraphSeed(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	statements []pg18GraphSeedStatement,
) {
	t.Helper()
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
}

func pg18InsertSnapshotChild(
	ctx context.Context, admin *pgx.Conn,
) error {
	tx, err := admin.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, `INSERT INTO tenant_a.frm_graph_a
VALUES($1,'Aardvark Snapshot')`, pg18GraphASnapshot); err != nil {
		return err
	}
	relation := pg18GraphQualified("tenant_a", pg18GraphRelationTable(graphRelationA))
	if _, err = tx.Exec(ctx, "INSERT INTO "+relation+" VALUES($1,$2,4)",
		pg18GraphRootOne, pg18GraphASnapshot); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
