//go:build integration

package executor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pg18GraphCapSource = "83000000-0000-4000-8000-000000000001"

func pg18VerifyGraphStorage(t *testing.T, gate *pg18Gate) {
	for _, schema := range []string{"tenant_a", "tenant_b"} {
		pg18AssertGraphRelation(t, gate, schema, graphRelationA, false)
		pg18AssertGraphRelation(t, gate, schema, graphRelationB, true)
		pg18AssertGraphRelation(t, gate, schema, graphRelationC, false)
	}
	pg18VerifyToManyCap(t, gate)
	pg18VerifyToOneCap(t, gate)
}

func pg18AssertGraphRelation(
	t *testing.T, gate *pg18Gate, schema, relationID string, toOne bool,
) {
	t.Helper()
	regclass := schema + "." + pg18GraphRelationTable(relationID)
	var names, types, definitions []string
	var allRequired bool
	err := gate.admin.QueryRow(gate.ctx, `SELECT
array_agg(attribute.attname ORDER BY attribute.attnum),
array_agg(pg_catalog.format_type(attribute.atttypid,attribute.atttypmod)
ORDER BY attribute.attnum),pg_catalog.bool_and(attribute.attnotnull),
(SELECT array_agg(pg_catalog.pg_get_constraintdef(constraint_row.oid)
ORDER BY constraint_row.contype,constraint_row.conname)
FROM pg_catalog.pg_constraint AS constraint_row
WHERE constraint_row.conrelid=pg_catalog.to_regclass($1))
FROM pg_catalog.pg_attribute AS attribute
WHERE attribute.attrelid=pg_catalog.to_regclass($1)
AND attribute.attnum>0 AND NOT attribute.attisdropped`, regclass).Scan(
		&names, &types, &allRequired, &definitions,
	)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(definitions, "|")
	wantCheck, wantUnique := "edge_slot >= 1", "UNIQUE (source_id, edge_slot)"
	if toOne {
		wantCheck, wantUnique = "edge_slot = 1", "UNIQUE (source_id)"
	}
	if fmt.Sprint(names) != "[source_id target_id edge_slot]" ||
		fmt.Sprint(types) != "[uuid uuid smallint]" || !allRequired ||
		!strings.Contains(joined, "PRIMARY KEY (source_id, target_id)") ||
		!strings.Contains(joined, wantCheck) || !strings.Contains(joined, wantUnique) {
		t.Fatalf("relation=%s names=%v types=%v required=%t constraints=%v",
			regclass, names, types, allRequired, definitions)
	}
}

func pg18VerifyToManyCap(t *testing.T, gate *pg18Gate) {
	table := pg18GraphQualified("tenant_a", pg18GraphRelationTable(graphRelationA))
	_, err := gate.admin.Exec(gate.ctx, `INSERT INTO `+table+`
SELECT $1,pg_catalog.md5('cap-'||series.value::text)::uuid,
series.value::smallint FROM pg_catalog.generate_series(1,99) AS series(value)`,
		pg18GraphCapSource)
	if err != nil {
		t.Fatal(err)
	}
	connections := make([]*pgx.Conn, 2)
	transactions := make([]pgx.Tx, 2)
	for index := range connections {
		connections[index], err = pgx.Connect(gate.ctx, gate.admin.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		defer connections[index].Close(context.Background())
		transactions[index], err = connections[index].Begin(gate.ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
	start, results := make(chan struct{}), make(chan error, 2)
	for index, target := range []string{
		"83000000-0000-4000-8000-000000000100",
		"83000000-0000-4000-8000-000000000101",
	} {
		tx := transactions[index]
		go func() {
			<-start
			_, runErr := tx.Exec(gate.ctx, "INSERT INTO "+table+" VALUES($1,$2,100)",
				pg18GraphCapSource, target)
			if runErr == nil {
				runErr = tx.Commit(gate.ctx)
			} else {
				_ = tx.Rollback(context.Background())
			}
			results <- runErr
		}()
	}
	close(start)
	first, second := <-results, <-results
	if !pg18OneSuccessOneCode(first, second, "23505") {
		t.Fatalf("concurrent slot 100 errors=%v/%v", first, second)
	}
	var count int
	if err := gate.admin.QueryRow(gate.ctx,
		"SELECT count(*) FROM "+table+" WHERE source_id=$1", pg18GraphCapSource,
	).Scan(&count); err != nil || count != 100 {
		t.Fatalf("edge count=%d error=%v", count, err)
	}
	_, err = gate.admin.Exec(gate.ctx, "INSERT INTO "+table+" VALUES($1,$2,101)",
		pg18GraphCapSource, "83000000-0000-4000-8000-000000000102")
	if !pg18HasSQLState(err, "23514") {
		t.Fatalf("slot 101 error=%v", err)
	}
}

func pg18VerifyToOneCap(t *testing.T, gate *pg18Gate) {
	table := pg18GraphQualified("tenant_a", pg18GraphRelationTable(graphRelationB))
	source := "84000000-0000-4000-8000-000000000001"
	if _, err := gate.admin.Exec(gate.ctx, "INSERT INTO "+table+" VALUES($1,$2,1)",
		source, "84000000-0000-4000-8000-000000000002"); err != nil {
		t.Fatal(err)
	}
	_, duplicate := gate.admin.Exec(gate.ctx, "INSERT INTO "+table+" VALUES($1,$2,1)",
		source, "84000000-0000-4000-8000-000000000003")
	_, wrongSlot := gate.admin.Exec(gate.ctx, "INSERT INTO "+table+" VALUES($1,$2,2)",
		"84000000-0000-4000-8000-000000000004",
		"84000000-0000-4000-8000-000000000005")
	if !pg18HasSQLState(duplicate, "23505") || !pg18HasSQLState(wrongSlot, "23514") {
		t.Fatalf("to-one duplicate=%v wrong-slot=%v", duplicate, wrongSlot)
	}
}

func pg18OneSuccessOneCode(left, right error, code string) bool {
	return left == nil && pg18HasSQLState(right, code) ||
		right == nil && pg18HasSQLState(left, code)
}

func pg18HasSQLState(err error, code string) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == code
}
