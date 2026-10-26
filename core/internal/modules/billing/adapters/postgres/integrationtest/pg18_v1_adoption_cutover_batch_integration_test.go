//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	billingv1adoptiondb "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/v1adoption"
	"github.com/google/uuid"
)

type protectedBillingBatch struct {
	bytes, lastKey             []byte
	digest, sourceXID          string
	commitLSN, checkpointState string
	finalLSN                   any
	batchOrdinal               int
}

func newProtectedBillingPlanBatch(a protectedBillingAuthority,
	checkpointState, commitLSN, sealSeed string,
) protectedBillingBatch {
	chunkDigest := sha256.Sum256(a.planBytes)
	after := map[string]any{
		"chunk_base64": base64.StdEncoding.EncodeToString(a.planBytes),
		"chunk_count":  1, "chunk_ordinal": 1,
		"chunk_sha256": hex.EncodeToString(chunkDigest[:]),
	}
	changes := []any{map[string]any{
		"after": after, "before": nil,
		"key":  map[string]any{"chunk_ordinal": 1},
		"kind": "insert", "ordinal": 1,
	}}
	var finalLSN any
	if checkpointState == "complete" {
		finalLSN = commitLSN
	}
	return newProtectedBillingBatch(a, changes, 1, a.sourceXID,
		checkpointState, commitLSN, finalLSN, sealSeed)
}

func newProtectedBillingQuietBatch(a protectedBillingAuthority,
	sealSeed string,
) protectedBillingBatch {
	return newProtectedBillingBatch(a, []any{}, 2, "8102", "complete",
		a.finalLSN, a.finalLSN, sealSeed)
}

func newProtectedBillingBatch(a protectedBillingAuthority, changes []any,
	batchOrdinal int, sourceXID, state, commitLSN string, finalLSN any,
	sealSeed string,
) protectedBillingBatch {
	sealDigest := sha256.Sum256([]byte(sealSeed))
	lastKey := map[string]any{
		"session_seal_receipt_id": uuid.NewString(),
		"session_seal_sha256":     hex.EncodeToString(sealDigest[:]),
	}
	sourcePosition := map[string]any{
		"batch_ordinal": batchOrdinal, "commit_lsn": commitLSN,
		"final_lsn": finalLSN, "received_lsn": "0/A0",
		"snapshot_lsn": "0/20", "snapshot_name": "billing-" + a.scopeKind,
		"source_xid": sourceXID,
	}
	document := map[string]any{
		"changes": changes, "checkpoint_key": a.checkpoint,
		"checkpoint_state": state, "last_key": lastKey,
		"operation_id": a.operationID, "relation_ordinal": 1,
		"run_id": a.runID, "scope_id": a.scopeID,
		"scope_kind": a.scopeKind, "source_position": sourcePosition,
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	lastKeyBytes, err := json.Marshal(lastKey)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(canonical)
	return protectedBillingBatch{
		bytes: canonical, lastKey: lastKeyBytes,
		digest: hex.EncodeToString(digest[:]), sourceXID: sourceXID,
		commitLSN: commitLSN, checkpointState: state, finalLSN: finalLSN,
		batchOrdinal: batchOrdinal,
	}
}

func executeProtectedBillingBatch(ctx context.Context, database *sql.DB,
	a protectedBillingAuthority, batch protectedBillingBatch,
	asAdopter, commit bool,
) (int64, string, string, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", "", err
	}
	rollback := func(cause error) (int64, string, string, error) {
		_ = transaction.Rollback()
		return 0, "", "", cause
	}
	if asAdopter {
		if _, err = transaction.ExecContext(ctx,
			"SET LOCAL SESSION AUTHORIZATION harpy_adopter"); err != nil {
			return rollback(err)
		}
	}
	contract, err := billingCutoverContract(a.scopeKind)
	if err != nil {
		return rollback(err)
	}
	var applied int64
	var keyDigest, stateDigest string
	err = transaction.QueryRowContext(ctx, contract.Query(),
		a.operationID, a.runID, a.scopeKind, a.scopeID, 1, a.checkpoint,
		"billing-"+a.scopeKind, "0/20", "0/A0", batch.commitLSN,
		batch.finalLSN, batch.sourceXID, batch.batchOrdinal,
		batch.checkpointState, string(batch.lastKey), batch.bytes,
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

func billingCutoverContract(scope string) (
	billingv1adoptiondb.CutoverContract, error,
) {
	if scope == "global" {
		return billingv1adoptiondb.NewGlobalCutoverContract()
	}
	if scope == "tenant" {
		return billingv1adoptiondb.NewTenantCutoverContract()
	}
	return billingv1adoptiondb.CutoverContract{}, fmt.Errorf(
		"unsupported Billing cutover scope %q", scope)
}
