//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func verifyCompositeInputBounds(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	tests := []struct {
		name          string
		payload       []byte
		digest        string
		declaredBytes int64
		code          string
		message       string
	}{
		{name: "too many changes", payload: buildBoundaryPayload(t, 90, 1001, 0),
			code: "22023", message: "batch evidence differs"},
		{name: "oversized payload", payload: buildBoundaryPayload(
			t, 91, 1, cutoverv1.MaximumChangeBatchBytes),
			code: "22023", message: "batch size invalid"},
		{name: "false byte count", payload: buildBoundaryPayload(t, 92, 1, 0),
			code: "22023", message: "batch size invalid"},
		{name: "false digest", payload: buildBoundaryPayload(t, 93, 1, 0),
			digest: digestMarker('f'), code: "22000", message: "batch digest differs"},
	}
	for index := range tests {
		test := &tests[index]
		if test.digest == "" {
			test.digest = fmt.Sprintf("%x", sha256.Sum256(test.payload))
		}
		test.declaredBytes = int64(len(test.payload))
		if test.name == "false byte count" {
			test.declaredBytes++
		}
		t.Run(test.name, func(t *testing.T) {
			_, err := callCompositeDocument(
				ctx, adopter, 90+index, "delta", test.payload,
				test.digest, test.declaredBytes, digestMarker('a'),
			)
			assertFinalizationError(t, err, test.code, test.message)
			assertNoApplyEffects(t, ctx, admin)
		})
	}
	verifyLeaseManifestDigestDenied(t, ctx, adopter, admin)
	verifyBatchIdentityBindingsDenied(t, ctx, adopter, admin)
	verifyUnknownBatchMembersDenied(t, ctx, adopter, admin)
	verifyNullChangeKindDenied(t, ctx, adopter, admin)
	verifyOwnerApplyBounds(t, ctx, adopter, admin)
}

func verifyUnknownBatchMembersDenied(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	tests := []struct {
		name    string
		member  string
		message string
	}{
		{"root", "root", "batch evidence differs"},
		{"position", "position", "batch evidence differs"},
		{"change", "change", "batch shape differs"},
	}
	for index, test := range tests {
		t.Run("unknown "+test.name+" member", func(t *testing.T) {
			ordinal := 94 + index
			payload := addUnknownBatchMember(t,
				buildBoundaryPayload(t, ordinal, 1, 0), test.member)
			digest := fmt.Sprintf("%x", sha256.Sum256(payload))
			_, err := callCompositeDocument(ctx, adopter, ordinal, "delta",
				payload, digest, int64(len(payload)), digestMarker('a'))
			assertFinalizationError(t, err, "22023", test.message)
			assertNoApplyEffects(t, ctx, admin)
		})
	}
}

func addUnknownBatchMember(t *testing.T, document []byte, location string) []byte {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(document, &payload); err != nil {
		t.Fatal(err)
	}
	switch location {
	case "root":
		payload["unknown"] = true
	case "position":
		payload["source_position"].(map[string]any)["unknown"] = true
	case "change":
		payload["changes"].([]any)[0].(map[string]any)["unknown"] = true
	default:
		t.Fatalf("unknown batch mutation location %s", location)
	}
	mutated, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func buildBoundaryPayload(
	t *testing.T, ordinal, changeCount, paddingBytes int,
) []byte {
	t.Helper()
	changes := make([]any, changeCount)
	for index := range changes {
		changes[index] = map[string]any{
			"ordinal": index + 1, "kind": "insert", "key": map[string]any{"id": index + 1},
			"before": nil, "after": map[string]any{"id": index + 1},
		}
	}
	commitLSN := fmt.Sprintf("0/%X", 0x20+ordinal)
	payload := map[string]any{
		"operation_id": integrationOperationID, "run_id": integrationRunID,
		"scope_kind": "tenant", "scope_id": integrationScopeID,
		"relation_ordinal": 1, "checkpoint_key": "main", "checkpoint_state": "delta",
		"source_position": map[string]any{
			"snapshot_name": "snapshot-main", "snapshot_lsn": "0/10",
			"received_lsn": "0/FF", "commit_lsn": commitLSN, "final_lsn": nil,
			"source_xid": fmt.Sprint(40 + ordinal), "batch_ordinal": ordinal,
		},
		"last_key": map[string]any{"id": ordinal}, "changes": changes,
	}
	if paddingBytes > 0 {
		payload["padding"] = strings.Repeat("x", paddingBytes)
	}
	document, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func verifyOwnerApplyBounds(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `CREATE FUNCTION platform.integration_owner_apply(jsonb,text)
RETURNS TABLE(applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $body$
SELECT * FROM platform.cutover_probe_apply_v1(
'10000000-0000-4000-8000-000000000001'::uuid,'10000000-0000-4000-8000-000000000002'::uuid,
'tenant','10000000-0000-4000-8000-000000000006'::uuid,1::smallint,'main',
'0/10'::pg_lsn,'0/FF'::pg_lsn,'0/7A'::pg_lsn,'130'::xid8,90,$1,$2)
$body$; GRANT EXECUTE ON FUNCTION platform.integration_owner_apply(jsonb,text) TO harpy_adopter;`)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{
		buildBoundaryPayload(t, 90, 1001, 0),
		buildBoundaryPayload(t, 90, 1, cutoverv1.MaximumChangeBatchBytes),
		withNullChangeKind(t, buildBoundaryPayload(t, 90, 1, 0)),
	} {
		digest := fmt.Sprintf("%x", sha256.Sum256(payload))
		_, err := adopter.Exec(ctx,
			"SELECT * FROM platform.integration_owner_apply($1::jsonb,$2::text)",
			json.RawMessage(payload), digest)
		assertFinalizationError(t, err, "22023", "probe apply binding differs")
		assertNoApplyEffects(t, ctx, admin)
	}
}

func assertNoApplyEffects(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	var effects int
	err := admin.QueryRow(ctx, `SELECT
    (SELECT count(*) FROM platform.cutover_batch_receipts) +
    (SELECT count(*) FROM platform.cutover_contract_probe_rows) +
    (SELECT count(*) FROM platform.cutover_copy_checkpoints)`).Scan(&effects)
	if err != nil || effects != 0 {
		t.Fatalf("rejected batch effects = %d, err=%v", effects, err)
	}
}
