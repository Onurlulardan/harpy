//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

func verifyBatchIdentityBindingsDenied(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	tests := []struct {
		name  string
		field string
		value any
	}{
		{"operation", "operation_id", "20000000-0000-4000-8000-000000000001"},
		{"run", "run_id", "20000000-0000-4000-8000-000000000002"},
		{"scope kind null", "scope_kind", nil},
		{"scope id null", "scope_id", nil},
		{"relation", "relation_ordinal", 2},
		{"checkpoint", "checkpoint_key", "other"},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ordinal := 101 + index
			payload := mutateBatchRoot(t, buildBoundaryPayload(t, ordinal, 1, 0),
				test.field, test.value)
			digest := fmt.Sprintf("%x", sha256.Sum256(payload))
			_, err := callCompositeDocument(ctx, adopter, ordinal, "delta", payload,
				digest, int64(len(payload)), digestMarker('a'))
			assertFinalizationError(t, err, "22023", "batch evidence differs")
			assertNoApplyEffects(t, ctx, admin)
		})
	}
	verifyNullOwnerApplyScopeDenied(t, ctx, adopter, admin)
}

func mutateBatchRoot(t *testing.T, document []byte, field string, value any) []byte {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(document, &payload); err != nil {
		t.Fatal(err)
	}
	payload[field] = value
	mutated, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func verifyNullOwnerApplyScopeDenied(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `CREATE FUNCTION platform.integration_null_scope_apply(jsonb,text)
RETURNS TABLE(applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $body$
SELECT * FROM platform.cutover_probe_apply_v1(
'10000000-0000-4000-8000-000000000001'::uuid,'10000000-0000-4000-8000-000000000002'::uuid,
NULL::text,NULL::uuid,1::smallint,'main','0/10','0/FF','0/20','41'::xid8,1,$1,$2)
$body$; GRANT EXECUTE ON FUNCTION platform.integration_null_scope_apply(jsonb,text)
TO harpy_adopter;`)
	if err != nil {
		t.Fatal(err)
	}
	payload := mutateBatchRoot(t, buildBoundaryPayload(t, 110, 1, 0), "scope_kind", nil)
	payload = mutateBatchRoot(t, payload, "scope_id", nil)
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	_, err = adopter.Exec(ctx,
		"SELECT * FROM platform.integration_null_scope_apply($1::jsonb,$2)",
		json.RawMessage(payload), digest)
	assertFinalizationError(t, err, "22023", "probe apply binding differs")
	assertNoApplyEffects(t, ctx, admin)
}
