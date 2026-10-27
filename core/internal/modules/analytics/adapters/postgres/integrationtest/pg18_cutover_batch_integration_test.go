//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	analyticsadoption "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/adoption"
)

type analyticsCutoverEvidence struct {
	DependencySHA256     string `json:"dependency_sha256"`
	EvidenceSHA256       string `json:"evidence_sha256"`
	PublicApprovalSHA256 string `json:"public_approval_sha256"`
	RollbackSHA256       string `json:"rollback_sha256"`
	SourceCount          int64  `json:"source_count"`
	SourceRootSHA256     string `json:"source_root_sha256"`
	TargetCount          int64  `json:"target_count"`
	TargetRootSHA256     string `json:"target_root_sha256"`
	WatermarkSHA256      string `json:"watermark_sha256"`
	ZeroHitSHA256        string `json:"zero_hit_sha256"`
}

type analyticsCutoverBatch struct {
	ordinal, sourceXID            int
	state, commitLSN, receivedLSN string
	finalLSN, lastKey, digest     any
	bytes                         []byte
}

func loadAnalyticsCutoverEvidence(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority,
) analyticsCutoverEvidence {
	t.Helper()
	evidence := analyticsCutoverEvidence{}
	err := database.QueryRowContext(ctx, `SELECT
count(*) FILTER(WHERE NOT removed),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||source_sha256,
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(WHERE NOT removed),''::bytea)),'hex'),
count(*) FILTER(WHERE NOT removed AND disposition='adopted'),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||target_sha256,
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(
 WHERE NOT removed AND disposition='adopted'),''::bytea)),'hex'),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||source_sha256||
 decode(target_document->'published_version'->>'manifest_sha256','hex'),
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(
 WHERE NOT removed AND disposition='adopted'),''::bytea)),'hex'),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||
 uuid_send((target_document->'release'->>'id')::uuid)||
 decode(target_document->'approval'->>'evidence_sha256','hex')||
 decode(target_document->'public_pointer'->>'public_id','hex'),
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(
 WHERE NOT removed AND disposition='adopted' AND
 (source_document->>'is_public')::boolean),''::bytea)),'hex'),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||
 decode(target_document->'public_pointer'->>'public_id','hex')||
 decode(target_document->'legacy_route'->>'consumer_inventory_sha256','hex')||
 timestamptz_send((target_document->'legacy_route'->>'consumers_acked_at')::timestamptz)||
 timestamptz_send((target_document->'legacy_route'->>'remove_after')::timestamptz),
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(
 WHERE NOT removed AND disposition='adopted' AND
 (source_document->>'is_public')::boolean),''::bytea)),'hex'),
encode(sha256(COALESCE(string_agg(uuid_send(legacy_endpoint_id)||
 decode(target_document->'differential'->>'source_result_sha256','hex')||
 decode(target_document->'differential'->>'target_result_sha256','hex')||
 decode(target_document->'differential'->>'wire_contract_sha256','hex')||
 decode(target_document->'differential'->>'compatible_image_sha256','hex')||
 decode(target_document->'differential'->>'reverse_source_sha256','hex')||
 decode(target_document->'differential'->>'evidence_sha256','hex'),
 ''::bytea ORDER BY legacy_endpoint_id) FILTER(
 WHERE NOT removed AND disposition='adopted'),''::bytea)),'hex'),
encode(sha256(convert_to('analytics-v1-adoption-watermark.v1','UTF8')||
 pg_lsn_send($3::pg_lsn)||COALESCE(string_agg(uuid_send(legacy_endpoint_id)||
 xid8send(source_xid)||pg_lsn_send(commit_lsn)||int4send(batch_ordinal)||
 int2send(change_ordinal),''::bytea ORDER BY legacy_endpoint_id),''::bytea)),'hex')
FROM t_analytics_a.analytics_v1_adoption_stage
WHERE operation_id=$1 AND run_id=$2`, authority.operationID, authority.runID,
		authority.finalLSN).Scan(&evidence.SourceCount, &evidence.SourceRootSHA256,
		&evidence.TargetCount, &evidence.TargetRootSHA256, &evidence.DependencySHA256,
		&evidence.PublicApprovalSHA256, &evidence.ZeroHitSHA256,
		&evidence.RollbackSHA256, &evidence.WatermarkSHA256)
	if err != nil {
		t.Fatal(err)
	}
	fields := []string{fmt.Sprint(evidence.SourceCount), evidence.SourceRootSHA256,
		fmt.Sprint(evidence.TargetCount), evidence.TargetRootSHA256,
		evidence.DependencySHA256, evidence.PublicApprovalSHA256,
		evidence.ZeroHitSHA256, evidence.RollbackSHA256, evidence.WatermarkSHA256}
	buffer := bytes.NewBufferString("analytics-v1-adoption-evidence.v1")
	for _, field := range fields {
		buffer.WriteByte(0)
		buffer.WriteString(field)
	}
	digest := sha256.Sum256(buffer.Bytes())
	evidence.EvidenceSHA256 = hex.EncodeToString(digest[:])
	return evidence
}

func newAnalyticsCutoverBatch(authority analyticsCutoverAuthority, ordinal int,
	state string, before, after *analyticsCutoverValue,
	evidence *analyticsCutoverEvidence,
) analyticsCutoverBatch {
	changes := []any{}
	if before != nil || after != nil {
		kind := "update"
		if before == nil {
			kind = "insert"
		} else if after == nil {
			kind = "delete"
		}
		changes = append(changes, map[string]any{"kind": kind, "ordinal": 1,
			"key":    map[string]any{"legacy_endpoint_id": analyticsCutoverEndpoint},
			"before": analyticsCutoverPayload(before), "after": analyticsCutoverPayload(after)})
	}
	commit, received, final := "0/30", "0/40", any(nil)
	if state == "delta" {
		commit, received = "0/50", "0/60"
	} else if state == "complete" {
		commit, received, final = authority.finalLSN, authority.finalLSN, authority.finalLSN
	}
	lastKey := any(nil)
	if evidence != nil {
		lastKey = evidence
	}
	sourceXID := 8160 + ordinal
	document := map[string]any{"changes": changes, "checkpoint_key": "openapi_endpoints",
		"checkpoint_state": state, "last_key": lastKey,
		"operation_id": authority.operationID, "relation_ordinal": 1,
		"run_id": authority.runID, "scope_id": authority.tenantID,
		"scope_kind": "tenant", "source_position": map[string]any{
			"batch_ordinal": ordinal, "commit_lsn": commit, "final_lsn": final,
			"received_lsn": received, "snapshot_lsn": "0/20",
			"snapshot_name": "analytics-phase8-snapshot", "source_xid": sourceXID}}
	canonical, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(canonical)
	lastBytes, _ := json.Marshal(lastKey)
	return analyticsCutoverBatch{ordinal: ordinal, sourceXID: sourceXID, state: state,
		commitLSN: commit, receivedLSN: received, finalLSN: final,
		lastKey: string(lastBytes), bytes: canonical, digest: hex.EncodeToString(digest[:])}
}

func analyticsCutoverPayload(value *analyticsCutoverValue) any {
	if value == nil {
		return nil
	}
	return map[string]any{"source_document": json.RawMessage(value.sourceDocument),
		"source_sha256": value.sourceSHA256, "disposition": "adopted",
		"reason_code": nil, "target_document": json.RawMessage(value.targetDocument),
		"target_sha256": value.targetSHA256}
}

func executeAnalyticsCutoverBatch(ctx context.Context, database *sql.DB,
	authority analyticsCutoverAuthority, schema string, batch analyticsCutoverBatch,
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
		"SELECT set_config('search_path',$1,true)", schema+",pg_catalog"); err != nil {
		return rollback(err)
	}
	contract, err := analyticsadoption.NewCutoverContract()
	if err != nil {
		return rollback(err)
	}
	var count int64
	var keyDigest, stateDigest string
	err = transaction.QueryRowContext(ctx, contract.Query(), authority.operationID,
		authority.runID, "tenant", authority.tenantID, 1, "openapi_endpoints",
		"analytics-phase8-snapshot", "0/20", batch.receivedLSN,
		batch.commitLSN, batch.finalLSN, batch.sourceXID, batch.ordinal, batch.state,
		batch.lastKey, batch.bytes, batch.digest, int64(len(batch.bytes)),
		authority.leaseManifest).Scan(&count, &keyDigest, &stateDigest)
	if err != nil {
		return rollback(err)
	}
	if err = transaction.Commit(); err != nil {
		return 0, "", "", err
	}
	return count, keyDigest, stateDigest, nil
}
