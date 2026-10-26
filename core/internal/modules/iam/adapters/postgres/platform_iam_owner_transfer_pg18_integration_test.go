//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	ownerTenantID = "10000000-0000-4000-8000-000000000001"
	ownerEnvID    = "30000000-0000-4000-8000-000000000001"
	ownerUserID   = "41000000-0000-4000-8000-000000000001"
	successorID   = "42000000-0000-4000-8000-000000000002"
	registryID    = "55000000-0000-4000-8000-000000000001"
)

type ownerTransferFixture struct {
	migrator *sql.Conn
	runtime  *sql.Conn
}

type ownerTransferAttempt struct {
	operationID, actorID, successorID                string
	successorSubjectID, successorBindingID           string
	actorRevision, actorToken, actorAuth             int64
	successorRevision, successorToken, successorAuth int64
	clock, policyRevision, headGeneration            int64
	request, head, state, proposal, signature        []byte
	ack                                              []byte
	authTime, expiresAt                              time.Time
}

type ownerTransferResult struct {
	operationID, registryBindingID, scopeID string
	generation                              int64
	previousHead, state, proposal           []byte
	status                                  string
	replayed                                bool
}

func TestPlatformIAMOwnerTransferPG18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "owner_transfer_"+origin)
			fixture := prepareOwnerTransferPG18(t, ctx, database, origin)
			assertOwnerProjection(t, ctx, fixture.runtime,
				"53000000-0000-4000-8000-000000000001", 1)
			first := firstOwnerTransferAttempt()
			exerciseOwnerTransferCommit(t, ctx, fixture, first)
			assertOwnerProjection(t, ctx, fixture.runtime,
				first.successorBindingID, 2)
			second := compensatedOwnerTransferAttempt(first.state)
			exerciseOwnerTransferCompensation(t, ctx, fixture, second)
		})
	}
}

func assertOwnerProjection(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	wantBinding string, wantRevision int64,
) {
	t.Helper()
	var environment, binding string
	var revision int64
	err := connection.QueryRowContext(ctx,
		`SELECT * FROM platform.iam_platform_current_owner_projection_v1($1)`,
		ownerEnvID).Scan(&environment, &binding, &revision)
	if err != nil {
		t.Fatal(err)
	}
	if environment != ownerEnvID || binding != wantBinding || revision != wantRevision {
		t.Fatalf("unexpected owner projection %s/%s/%d", environment, binding, revision)
	}
}

func firstOwnerTransferAttempt() ownerTransferAttempt {
	now := time.Now().UTC().Add(-time.Second)
	return ownerTransferAttempt{
		operationID: "71000000-0000-4000-8000-000000000001",
		actorID:     ownerUserID, successorID: successorID,
		successorSubjectID: "61000000-0000-4000-8000-000000000001",
		successorBindingID: "63000000-0000-4000-8000-000000000001",
		actorRevision:      1, actorToken: 1, actorAuth: 1,
		successorRevision: 1, successorToken: 1, successorAuth: 1,
		clock: 1, policyRevision: 1, headGeneration: 1,
		request: digestByte(0x31), head: digestByte(0x11), state: digestByte(0x12),
		proposal: digestByte(0x22), signature: digestByte(0x23), ack: digestByte(0x32),
		authTime: now, expiresAt: now.Add(5 * time.Minute),
	}
}

func compensatedOwnerTransferAttempt(head []byte) ownerTransferAttempt {
	now := time.Now().UTC().Add(-time.Second)
	return ownerTransferAttempt{
		operationID: "71000000-0000-4000-8000-000000000002",
		actorID:     successorID, successorID: ownerUserID,
		successorSubjectID: "51000000-0000-4000-8000-000000000001",
		successorBindingID: "63000000-0000-4000-8000-000000000002",
		actorRevision:      2, actorToken: 2, actorAuth: 2,
		successorRevision: 2, successorToken: 2, successorAuth: 2,
		clock: 2, policyRevision: 2, headGeneration: 2,
		request: digestByte(0x41), head: head, state: digestByte(0x13),
		proposal: digestByte(0x42), signature: digestByte(0x43), ack: digestByte(0x44),
		authTime: now, expiresAt: now.Add(5 * time.Minute),
	}
}

func digestByte(value byte) []byte {
	return []byte(strings.Repeat(string([]byte{value}), sha256.Size))
}

func transferPurpose(attempt ownerTransferAttempt) []byte {
	return sha256Text(fmt.Sprintf("platform.owner.transfer\n%s\n%s\n%d\n%s",
		attempt.operationID, attempt.successorID, attempt.headGeneration+1,
		hex.EncodeToString(attempt.state)))
}

func transferAckPurpose(operationID string, ack []byte) []byte {
	return sha256Text(fmt.Sprintf("platform.owner.transfer.ack\n%s\n%s",
		operationID, hex.EncodeToString(ack)))
}

func compensationPurpose(operationID, originalID string, generation int64, ack []byte) []byte {
	return sha256Text(fmt.Sprintf("platform.owner.transfer.compensate\n%s\n%s\n%d\n%s",
		operationID, originalID, generation, hex.EncodeToString(ack)))
}

func sha256Text(value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return digest[:]
}
