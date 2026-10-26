//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func callOwnerTransferPrepare(
	t *testing.T, ctx context.Context, connection *sql.Conn, attempt ownerTransferAttempt,
) ownerTransferResult {
	t.Helper()
	row := connection.QueryRowContext(ctx, `SELECT * FROM iam_owner_transfer_prepare_v1(
        $1::name,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
        $8::bigint,$9::bytea,$10::timestamptz,$11::text,$12::boolean,$13::boolean,$14::boolean,
        $15::uuid,$16::bigint,$17::bigint,$18::bigint,$19::uuid,$20::uuid,$21::uuid,
        $22::bigint,$23::bigint,$24::bytea,$25::bytea,$26::bytea,$27::text,$28::bytea,
        $29::timestamptz)`,
		"t_owner", attempt.operationID, attempt.request, attempt.actorID,
		attempt.actorRevision, attempt.actorToken, attempt.actorAuth, attempt.clock,
		transferPurpose(attempt), attempt.authTime, "urn:harpy:acr:step-up",
		true, false, false, attempt.successorID, attempt.successorRevision,
		attempt.successorToken, attempt.successorAuth, attempt.successorSubjectID,
		attempt.successorBindingID, registryID, attempt.policyRevision,
		attempt.headGeneration, attempt.head, attempt.state, attempt.proposal,
		"registry-key-1", attempt.signature, attempt.expiresAt)
	return scanOwnerTransferResult(t, row)
}

func callOwnerTransferAck(
	t *testing.T, ctx context.Context, connection *sql.Conn, attempt ownerTransferAttempt,
	sequence int64,
) ownerTransferResult {
	t.Helper()
	row := connection.QueryRowContext(ctx, `SELECT * FROM iam_owner_transfer_ack_v1(
        $1::name,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
        $8::bytea,$9::timestamptz,$10::text,$11::boolean,$12::boolean,$13::boolean,
        $14::uuid,$15::uuid,$16::bigint,$17::bigint,$18::bytea,$19::bytea,$20::bytea,
        $21::bytea,$22::timestamptz,$23::text)`,
		"t_owner", attempt.operationID, attempt.request, attempt.actorID,
		attempt.actorToken, attempt.actorAuth, attempt.clock,
		transferAckPurpose(attempt.operationID, attempt.ack), attempt.authTime,
		"urn:harpy:acr:step-up", true, false, false, registryID, ownerEnvID,
		sequence, attempt.headGeneration+1, attempt.proposal, attempt.head,
		attempt.state, attempt.ack, attempt.authTime, "registry-key-1")
	return scanOwnerTransferResult(t, row)
}

func callOwnerTransferCommit(
	ctx context.Context, connection *sql.Conn, attempt ownerTransferAttempt,
	eventID string, eventDigest []byte,
) (ownerTransferCommitResult, error) {
	row := connection.QueryRowContext(ctx, `SELECT * FROM iam_owner_transfer_commit_v1(
        $1::name,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
        $8::uuid,$9::bytea)`,
		"t_owner", attempt.operationID, attempt.request, attempt.actorID,
		attempt.actorToken, attempt.actorAuth, attempt.clock, eventID, eventDigest)
	var result ownerTransferCommitResult
	err := row.Scan(&result.operationID, &result.predecessorID, &result.successorID,
		&result.predecessorToken, &result.predecessorAuth,
		&result.successorToken, &result.successorAuth,
		&result.policyRevision, &result.clock, &result.generation, &result.replayed)
	return result, err
}

type ownerTransferCommitResult struct {
	operationID, predecessorID, successorID string
	predecessorToken, predecessorAuth       int64
	successorToken, successorAuth           int64
	policyRevision, clock, generation       int64
	replayed                                bool
}

func scanOwnerTransferResult(t *testing.T, row *sql.Row) ownerTransferResult {
	t.Helper()
	var result ownerTransferResult
	if err := row.Scan(&result.operationID, &result.registryBindingID, &result.scopeID,
		&result.generation, &result.previousHead, &result.state, &result.proposal,
		&result.status, &result.replayed); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertTransferResult(
	t *testing.T, result ownerTransferResult, attempt ownerTransferAttempt, status string,
) {
	t.Helper()
	if result.operationID != attempt.operationID || result.scopeID != ownerEnvID ||
		result.registryBindingID != registryID || result.generation != attempt.headGeneration+1 ||
		result.status != status || result.replayed {
		t.Fatalf("owner transfer result = %+v", result)
	}
}
