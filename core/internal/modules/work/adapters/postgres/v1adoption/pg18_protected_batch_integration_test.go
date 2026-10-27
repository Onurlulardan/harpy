//go:build integration

package v1adoption

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type protectedWorkBatch struct {
	checkpointKey string
	batchOrdinal  int
	sourceXID     string
	finalLSN      string
	bytes         []byte
	digest        string
}

func newProtectedBoardBatch(a protectedWorkAuthority, name string) protectedWorkBatch {
	sourceID := "41000000-0000-4000-8000-000000000101"
	targetID := "41000000-0000-4000-8000-000000000102"
	key := map[string]any{"source_id": sourceID, "source_kind": "task_boards",
		"target_id": targetID, "target_kind": "BOARD"}
	record := map[string]any{
		"ArchivedAt": nil, "CreatedAt": "2026-07-21T09:00:00Z",
		"CreatorID":   "41000000-0000-4000-8000-000000000103",
		"Description": "Protected PostgreSQL 18 adoption", "Lifecycle": "ACTIVE",
		"Name": name, "UpdatedAt": "2026-07-21T09:00:00Z",
	}
	after := map[string]any{"source_id": sourceID, "source_kind": "task_boards",
		"source_sha256": fmt.Sprintf("%064x", 44), "target_id": targetID,
		"target_kind": "BOARD", "target_record": record}
	changes := []any{map[string]any{"after": after, "before": nil, "key": key,
		"kind": "insert", "ordinal": 1}}
	return newProtectedWorkBatch(a, "task_boards", 1, "4101", a.finalLSN, changes)
}

func newProtectedQuietBatch(a protectedWorkAuthority, checkpoint string,
	ordinal int, finalLSN string,
) protectedWorkBatch {
	return newProtectedWorkBatch(a, checkpoint, ordinal,
		fmt.Sprintf("%d", 4100+ordinal), finalLSN, []any{})
}

func newProtectedWorkBatch(a protectedWorkAuthority, checkpoint string,
	ordinal int, sourceXID, finalLSN string, changes []any,
) protectedWorkBatch {
	sourcePosition := map[string]any{
		"batch_ordinal": ordinal, "commit_lsn": a.finalLSN,
		"final_lsn": finalLSN, "received_lsn": "0/A0", "snapshot_lsn": "0/20",
		"snapshot_name": "work-phase10-snapshot", "source_xid": sourceXID,
	}
	document := map[string]any{
		"changes": changes, "checkpoint_key": checkpoint,
		"checkpoint_state": "complete", "last_key": nil,
		"operation_id": a.operationID, "relation_ordinal": 1, "run_id": a.runID,
		"scope_id": a.tenantID, "scope_kind": "tenant",
		"source_position": sourcePosition,
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(canonical)
	return protectedWorkBatch{checkpointKey: checkpoint, batchOrdinal: ordinal,
		sourceXID: sourceXID, finalLSN: finalLSN, bytes: canonical,
		digest: hex.EncodeToString(digest[:])}
}

func executeProtectedWorkBatch(ctx context.Context, database *sql.DB,
	a protectedWorkAuthority, schema string, batch protectedWorkBatch, commit bool,
) (int64, string, string, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", "", err
	}
	rollback := func(cause error) (int64, string, string, error) {
		_ = transaction.Rollback()
		return 0, "", "", cause
	}
	if _, err = transaction.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION harpy_adopter"); err != nil {
		return rollback(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", schema+",pg_catalog"); err != nil {
		return rollback(err)
	}
	contract, err := NewCutoverContract()
	if err != nil {
		return rollback(err)
	}
	var applied int64
	var keyDigest, stateDigest string
	err = transaction.QueryRowContext(ctx, contract.Query(),
		a.operationID, a.runID, "tenant", a.tenantID, 1, batch.checkpointKey,
		"work-phase10-snapshot", "0/20", "0/A0", a.finalLSN, batch.finalLSN,
		batch.sourceXID, batch.batchOrdinal, "complete", "null", batch.bytes,
		batch.digest, int64(len(batch.bytes)), a.leaseManifest,
	).Scan(&applied, &keyDigest, &stateDigest)
	if err != nil {
		return rollback(err)
	}
	if !commit {
		return applied, keyDigest, stateDigest, transaction.Rollback()
	}
	if err = transaction.Commit(); err != nil {
		return 0, "", "", err
	}
	return applied, keyDigest, stateDigest, nil
}
