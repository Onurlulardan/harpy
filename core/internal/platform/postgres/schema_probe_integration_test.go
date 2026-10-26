//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertSchemaContractProbe(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	ready, err := pool.ProbeSchemaContract(ctx)
	if err != nil || !ready.Ready() || ready.Reason() != SchemaReady {
		t.Fatalf("ready schema contract rejected: %+v %v", ready, err)
	}
	if _, err := database.admin.Exec(
		ctx, "UPDATE platform.schema_contracts SET status='fenced', revision=2 WHERE component='core'",
	); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = database.admin.Exec(
			context.Background(),
			"UPDATE platform.schema_contracts SET status='ready', revision=3 WHERE component='core'",
		)
	})
	fenced, err := pool.ProbeSchemaContract(ctx)
	if err != nil || fenced.Ready() || fenced.Reason() != SchemaFenced {
		t.Fatalf("fenced schema contract accepted: %+v %v", fenced, err)
	}
	observation := pool.ProbeReadiness(ctx)
	if !observation.DatabaseReady() || !observation.DatabaseRoleReady() ||
		observation.SchemaContractReady() || observation.SchemaReason() != SchemaFenced {
		t.Fatalf("schema failure was not separated: %+v", observation)
	}
	if _, err := database.admin.Exec(
		ctx, "UPDATE platform.schema_contracts SET status='ready', revision=3 WHERE component='core'",
	); err != nil {
		t.Fatal(err)
	}
}
