//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type ownerCompensationAttempt struct {
	operationID, originalID, actorID          string
	token, authorization, clock, generation   int64
	request, originalAck, previousHead, state []byte
	proposal, signature, ack                  []byte
	authTime, expiresAt                       time.Time
}

type ownerCompensationResult struct {
	operationID, originalID string
	generation              int64
	state, ack              []byte
	replayed                bool
}

func newOwnerCompensationAttempt(original ownerTransferAttempt) ownerCompensationAttempt {
	now := time.Now().UTC().Add(-time.Second)
	return ownerCompensationAttempt{
		operationID: "72000000-0000-4000-8000-000000000001",
		originalID:  original.operationID, actorID: original.actorID,
		token: original.actorToken, authorization: original.actorAuth,
		clock: original.clock, generation: original.headGeneration + 2,
		request: digestByte(0x51), originalAck: original.ack,
		previousHead: original.state, state: original.head,
		proposal: digestByte(0x52), signature: digestByte(0x53), ack: digestByte(0x54),
		authTime: now, expiresAt: now.Add(5 * time.Minute),
	}
}

func callOwnerCompensationPrepare(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	attempt ownerCompensationAttempt,
) ownerTransferResult {
	t.Helper()
	row := connection.QueryRowContext(ctx, `SELECT * FROM iam_owner_transfer_compensate_prepare_v1(
        $1::name,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
        $8::bytea,$9::timestamptz,$10::text,$11::boolean,$12::boolean,$13::boolean,
        $14::uuid,$15::bytea,$16::bigint,$17::bytea,$18::text,$19::bytea,$20::timestamptz)`,
		"t_owner", attempt.operationID, attempt.request, attempt.actorID,
		attempt.token, attempt.authorization, attempt.clock,
		compensationPurpose(attempt.operationID, attempt.originalID,
			attempt.generation, attempt.originalAck),
		attempt.authTime, "urn:harpy:acr:step-up", true, false, false,
		attempt.originalID, attempt.originalAck, attempt.generation-1,
		attempt.proposal, "registry-key-1", attempt.signature, attempt.expiresAt)
	return scanOwnerTransferResult(t, row)
}

func callOwnerCompensationApply(
	ctx context.Context, connection *sql.Conn, attempt ownerCompensationAttempt,
) (ownerCompensationResult, error) {
	row := connection.QueryRowContext(ctx, `SELECT * FROM iam_owner_transfer_compensate_apply_v1(
        $1::name,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
        $8::bytea,$9::timestamptz,$10::text,$11::boolean,$12::boolean,$13::boolean,
        $14::uuid,$15::uuid,$16::bigint,$17::bigint,$18::bytea,$19::bytea,$20::bytea,
        $21::bytea,$22::timestamptz,$23::text)`,
		"t_owner", attempt.operationID, attempt.request, attempt.actorID,
		attempt.token, attempt.authorization, attempt.clock,
		transferAckPurpose(attempt.operationID, attempt.ack), attempt.authTime,
		"urn:harpy:acr:step-up", true, false, false, registryID, ownerEnvID,
		int64(3), attempt.generation, attempt.proposal, attempt.previousHead,
		attempt.state, attempt.ack, attempt.authTime, "registry-key-1")
	var result ownerCompensationResult
	err := row.Scan(&result.operationID, &result.originalID, &result.generation,
		&result.state, &result.ack, &result.replayed)
	return result, err
}
