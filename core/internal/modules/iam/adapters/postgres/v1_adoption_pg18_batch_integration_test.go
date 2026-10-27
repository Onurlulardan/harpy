//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func applyIAMV1AdoptionBatches(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, operationID, runID string,
) string {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM V1 adoption copier: %v", err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin IAM V1 adoption copier: %v", err)
	}
	defer transaction.Rollback()
	bindIAMV1Adopter(t, ctx, transaction, tenant.schema)
	organizationID := adoptedOrganizationID(tenant)
	for index, category := range iamV1AdoptionCategories[:14] {
		batch, lastKey := adoptionBatch(
			tenant, operationID, runID, int16(index+1), category,
		)
		want := int64(len(adoptedTargets(tenant, category)))
		got := applyIAMV1AdoptionBatch(
			t, ctx, transaction, tenant, operationID, runID,
			int16(index+1), category, batch, lastKey,
		)
		if got != want {
			t.Fatalf("%s applied count = %d, want %d", category, got, want)
		}
		if index == 0 {
			replayed := applyIAMV1AdoptionBatch(
				t, ctx, transaction, tenant, operationID, runID,
				int16(index+1), category, batch, lastKey,
			)
			if replayed != got {
				t.Fatalf("idempotent IAM V1 adoption replay = %d, want %d", replayed, got)
			}
		}
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM V1 adoption batches: %v", err)
	}
	return organizationID
}

func applyIAMV1AdoptionBatch(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	tenant iamFoundationTenant, operationID, runID string,
	relation int16, category string, batch []byte, lastKey []byte,
) int64 {
	t.Helper()
	digest := sha256.Sum256(batch)
	query := fmt.Sprintf(`SELECT applied_count, key_state_sha256, state_sha256
FROM %s.iam_v1_adoption_apply_and_record_v1(
 $1::uuid,$2::uuid,'tenant',$3::uuid,$4::smallint,$5::text,$6::text,
 $7::pg_lsn,$8::pg_lsn,$9::pg_lsn,$10::pg_lsn,$11::xid8,1,'complete',
 $12::jsonb,$13::bytea,$14::text,$15::bigint,$16::text)`, tenant.schema)
	var count int64
	var keySHA, stateSHA string
	err := transaction.QueryRowContext(
		ctx, query, operationID, runID, tenant.id, relation, category,
		"iam-v1-adoption-"+tenant.schema, "0/10", "0/30", "0/20", "0/20",
		fmt.Sprintf("%d", 500+relation), string(lastKey), batch,
		hex.EncodeToString(digest[:]), int64(len(batch)), strings.Repeat("4", 64),
	).Scan(&count, &keySHA, &stateSHA)
	if err != nil {
		t.Fatalf("apply IAM V1 adoption %s in %s: %v", category, tenant.schema, err)
	}
	if len(keySHA) != 64 || len(stateSHA) != 64 {
		t.Fatalf("IAM V1 adoption digest is incomplete: %s %s", keySHA, stateSHA)
	}
	return count
}

func adoptionBatch(
	tenant iamFoundationTenant, operationID, runID string,
	relation int16, category string,
) ([]byte, []byte) {
	targets := adoptedTargets(tenant, category)
	changes := make([]any, 0, len(targets))
	var lastKey any
	for index, target := range targets {
		key := map[string]any{
			"source_category": category, "source_key": target.sourceKey,
			"target_table": target.table, "target_key": target.targetKey,
		}
		after := map[string]any{
			"source_category": category, "source_key": target.sourceKey,
			"target_table": target.table, "target_key": target.targetKey,
			"target_record": target.record,
		}
		changes = append(changes, map[string]any{
			"ordinal": index + 1, "kind": "insert", "key": key,
			"before": nil, "after": after,
		})
		lastKey = map[string]any{"id": target.sourceKey}
	}
	payload := map[string]any{
		"operation_id": operationID, "run_id": runID, "scope_kind": "tenant",
		"scope_id": tenant.id, "relation_ordinal": relation,
		"checkpoint_key": category, "checkpoint_state": "complete",
		"last_key": lastKey, "changes": changes,
		"source_position": map[string]any{
			"snapshot_name": "iam-v1-adoption-" + tenant.schema,
			"snapshot_lsn":  "0/10", "received_lsn": "0/30", "commit_lsn": "0/20",
			"final_lsn": "0/20", "source_xid": fmt.Sprintf("%d", 500+relation),
			"batch_ordinal": 1,
		},
	}
	document, _ := json.Marshal(payload)
	keyDocument, _ := json.Marshal(lastKey)
	return document, keyDocument
}
