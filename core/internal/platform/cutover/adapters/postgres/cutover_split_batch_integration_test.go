//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func verifySameXIDSplitAcrossCheckpoints(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adopter *pgx.Conn,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.cutover_operation_checkpoint_keys
    (operation_id,run_id,relation_ordinal,ordinal,checkpoint_key)
VALUES ($1,$2,1,2,'bulk-a'),($1,$2,1,3,'bulk-b')`,
		integrationOperationID, integrationRunID)
	if err != nil {
		t.Fatalf("seed split checkpoint keys: %v", err)
	}
	if result, err := applySplitBatch(ctx, adopter, "bulk-a", 1, 1000); err != nil ||
		result.AppliedCount != 1000 {
		t.Fatalf("apply first split batch: result=%+v err=%v", result, err)
	}
	if _, err := applySplitBatch(ctx, adopter, "bulk-b", 1, 1); err == nil {
		t.Fatal("same-XID relation-global batch ordinal was reusable across checkpoints")
	}
	if result, err := applySplitBatch(ctx, adopter, "bulk-b", 2, 1); err != nil ||
		result.AppliedCount != 1 {
		t.Fatalf("apply second split batch: result=%+v err=%v", result, err)
	}
	var receiptCount, distinctOrdinals, minimumOrdinal, maximumOrdinal, batchCount int
	err = database.admin.QueryRow(ctx, `SELECT
    count(*)::integer,count(DISTINCT change_ordinal)::integer,
    min(change_ordinal)::integer,max(change_ordinal)::integer,
    (SELECT count(*)::integer FROM platform.cutover_batch_receipts
      WHERE operation_id=$1 AND run_id=$2 AND source_xid='90'::xid8)
FROM platform.applied_change_receipts
WHERE operation_id=$1 AND run_id=$2 AND source_xid='90'::xid8`,
		integrationOperationID, integrationRunID).Scan(
		&receiptCount, &distinctOrdinals, &minimumOrdinal, &maximumOrdinal, &batchCount,
	)
	if err != nil || receiptCount != 1001 || distinctOrdinals != 1001 ||
		minimumOrdinal != 1 || maximumOrdinal != 1001 || batchCount != 2 {
		t.Fatalf("split receipt identity = %d %d %d %d %d, err=%v",
			receiptCount, distinctOrdinals, minimumOrdinal, maximumOrdinal, batchCount, err)
	}
}

func applySplitBatch(
	ctx context.Context, adopter *pgx.Conn, checkpoint string, batchOrdinal, count int,
) (cutoverv1.ApplyResult, error) {
	changes := make([]any, count)
	for index := range changes {
		ordinal := index + 1
		changes[index] = map[string]any{
			"ordinal": ordinal, "kind": "insert", "key": map[string]any{"id": ordinal},
			"before": nil, "after": map[string]any{"id": ordinal},
		}
	}
	payload, _ := json.Marshal(map[string]any{
		"operation_id": integrationOperationID, "run_id": integrationRunID,
		"scope_kind": "tenant", "scope_id": integrationScopeID,
		"relation_ordinal": 1, "checkpoint_key": checkpoint, "checkpoint_state": "delta",
		"source_position": map[string]any{
			"snapshot_name": "snapshot-main", "snapshot_lsn": "0/10",
			"received_lsn": "0/90", "commit_lsn": "0/90", "final_lsn": nil,
			"source_xid": "90", "batch_ordinal": batchOrdinal,
		},
		"last_key": map[string]any{"id": fmt.Sprintf("%s-last", checkpoint)},
		"changes":  changes,
	})
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	var decoded struct {
		LastKey json.RawMessage `json:"last_key"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return cutoverv1.ApplyResult{}, err
	}
	var result cutoverv1.ApplyResult
	err := adopter.QueryRow(ctx, applyCutoverProbeV1SQL,
		integrationOperationID, integrationRunID, "tenant", integrationScopeID,
		1, checkpoint, "snapshot-main", "0/10", "0/90", "0/90", nil,
		"90", batchOrdinal, "delta", decoded.LastKey, payload,
		digest, len(payload), digestMarker('a'),
	).Scan(&result.AppliedCount, &result.KeyStateSHA256, &result.StateSHA256)
	return result, err
}
