//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func verifyCompositeApplyContract(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	first *pgx.Conn,
	second *pgx.Conn,
) {
	t.Helper()
	verifyTargetApplySchemaBinding(t, ctx, database, first)
	assertDirectApplyProceduresDenied(t, ctx, first)
	verifyCompositeInputBounds(t, ctx, first, database.admin)

	identical := raceApply(t, ctx, first, second, 1, "same", "same")
	if identical[0].err != nil || identical[1].err != nil ||
		identical[0].result != identical[1].result {
		t.Fatalf("identical race = %#v", identical)
	}
	different := raceApply(t, ctx, first, second, 2, "left", "right")
	succeeded := 0
	for _, outcome := range different {
		if outcome.err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("different hash race successes = %d, outcomes=%#v", succeeded, different)
	}
	if _, err := callCompositeApply(ctx, first, 3, "complete", "complete"); err != nil {
		t.Fatalf("complete checkpoint apply: %v", err)
	}
	assertApplyEvidence(t, ctx, database.admin)
	verifySameXIDSplitAcrossCheckpoints(t, ctx, database, first)
}

type applyOutcome struct {
	result cutoverv1.ApplyResult
	err    error
}

func raceApply(
	t *testing.T,
	ctx context.Context,
	first *pgx.Conn,
	second *pgx.Conn,
	ordinal int,
	firstDigest string,
	secondDigest string,
) [2]applyOutcome {
	t.Helper()
	start := make(chan struct{})
	outcomes := [2]applyOutcome{}
	connections := []*pgx.Conn{first, second}
	variants := []string{firstDigest, secondDigest}
	var group sync.WaitGroup
	for index := range connections {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			outcomes[index].result, outcomes[index].err = callCompositeApply(
				ctx, connections[index], ordinal, variants[index], "delta",
			)
		}(index)
	}
	close(start)
	group.Wait()
	return outcomes
}

func callCompositeApply(
	ctx context.Context,
	connection *pgx.Conn,
	ordinal int,
	variant string,
	state string,
) (cutoverv1.ApplyResult, error) {
	commitLSN := fmt.Sprintf("0/%X", 0x20+ordinal)
	var finalLSN *string
	if state == "complete" {
		finalLSN = &commitLSN
	}
	payload, _ := json.Marshal(map[string]any{
		"operation_id": integrationOperationID, "run_id": integrationRunID,
		"scope_kind": "tenant", "scope_id": integrationScopeID,
		"relation_ordinal": 1, "checkpoint_key": "main", "checkpoint_state": state,
		"source_position": map[string]any{
			"snapshot_name": "snapshot-main", "snapshot_lsn": "0/10",
			"received_lsn": "0/FF", "commit_lsn": commitLSN,
			"final_lsn": finalLSN, "source_xid": fmt.Sprint(40 + ordinal),
			"batch_ordinal": ordinal,
		},
		"last_key": map[string]any{"id": ordinal},
		"changes": []any{map[string]any{
			"ordinal": 1, "kind": "insert", "key": map[string]any{"id": ordinal},
			"before": nil, "after": map[string]any{"id": ordinal, "variant": variant},
		}},
	})
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	return callCompositeDocument(
		ctx, connection, ordinal, state, payload, digest, int64(len(payload)), digestMarker('a'),
	)
}

func callCompositeDocument(
	ctx context.Context, connection *pgx.Conn, ordinal int, state string,
	payload []byte, digest string, declaredBytes int64, leaseSHA256 string,
) (cutoverv1.ApplyResult, error) {
	commitLSN := fmt.Sprintf("0/%X", 0x20+ordinal)
	var finalLSN *string
	if state == "complete" {
		finalLSN = &commitLSN
	}
	var decoded struct {
		LastKey json.RawMessage `json:"last_key"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return cutoverv1.ApplyResult{}, err
	}
	var result cutoverv1.ApplyResult
	err := connection.QueryRow(ctx, applyCutoverProbeV1SQL,
		integrationOperationID, integrationRunID, "tenant", integrationScopeID,
		1, "main", "snapshot-main", "0/10", "0/FF", commitLSN, finalLSN,
		fmt.Sprint(40+ordinal), ordinal, state, decoded.LastKey,
		payload, digest, declaredBytes, leaseSHA256,
	).Scan(&result.AppliedCount, &result.KeyStateSHA256, &result.StateSHA256)
	return result, err
}

func assertDirectApplyProceduresDenied(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	_, err := connection.Exec(ctx, `SELECT * FROM platform.cutover_probe_apply_v1(
		$1::uuid,$2::uuid,'tenant'::text,$3::uuid,1::smallint,'main'::text,
		'0/10'::pg_lsn,'0/FF'::pg_lsn,'0/20'::pg_lsn,'41'::xid8,1::integer,
		'{}'::jsonb,$4::text)`, integrationOperationID, integrationRunID,
		integrationScopeID, digestMarker('2'))
	assertPermissionDenied(t, err)
	_, err = connection.Exec(ctx, `SELECT platform.cutover_record_batch_v1(
		$1::uuid,$2::uuid,1::smallint,'main'::text,'snapshot-main'::text,
		'0/10'::pg_lsn,'0/FF'::pg_lsn,'0/20'::pg_lsn,NULL::pg_lsn,
		'41'::xid8,1::integer,'delta'::text,'{}'::jsonb,$3::text,
		10::bigint,1::bigint,$4::text,$5::text)`,
		integrationOperationID, integrationRunID, digestMarker('3'),
		digestMarker('4'), digestMarker('5'))
	assertPermissionDenied(t, err)
}

func assertPermissionDenied(t *testing.T, err error) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("permission denial = %v", err)
	}
}

func assertApplyEvidence(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	var receipts, probes, attempts int
	var state string
	err := admin.QueryRow(ctx, `SELECT
    (SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$1),
    (SELECT count(*) FROM platform.cutover_contract_probe_rows WHERE operation_id=$1),
    checkpoint.attempt, checkpoint.state
FROM platform.cutover_copy_checkpoints AS checkpoint
WHERE checkpoint.operation_id=$1 AND checkpoint.run_id=$2 AND checkpoint.checkpoint_key='main'`,
		integrationOperationID, integrationRunID).Scan(&receipts, &probes, &attempts, &state)
	if err != nil || receipts != 3 || probes != 3 || attempts != 3 || state != "complete" {
		t.Fatalf("apply evidence = %d %d %d %s, err=%v", receipts, probes, attempts, state, err)
	}
}

func digestMarker(marker byte) string { return strings.Repeat(string(marker), 64) }
