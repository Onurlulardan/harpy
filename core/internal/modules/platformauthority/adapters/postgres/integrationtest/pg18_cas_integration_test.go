//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func exerciseAuthorityCAS(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	binding := "22222222-2222-4222-8222-222222222222"
	environment := "11111111-1111-4111-8111-111111111111"
	operation := "33333333-3333-4333-8333-333333333333"
	digestA := strings.Repeat("a", 64)
	digestB := strings.Repeat("b", 64)
	digestC := strings.Repeat("c", 64)
	if _, err := database.ExecContext(ctx, bindSQL,
		binding, environment, "onpremise", "development", binding,
		digestA, digestB, int64(1), "proposal.v1", digestC,
		digestA, []byte("proof"), now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, prepareSQL,
		operation, binding, "initial_environment_authority", "environment", environment,
		int64(1), nil, digestA, []byte("state"), digestB,
		[]byte(`{"signer_kid":"proposal.v1"}`), nil,
		now, now.Add(10*time.Minute), "proposal.v1", now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, recordAckSQL,
		operation, binding, "environment", environment, int64(1), nil,
		digestA, digestB, digestC, []byte("ack"), int64(1), now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, commitSQL, operation, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, commitSQL, operation, now); err != nil {
		t.Fatalf("idempotent crash replay failed: %v", err)
	}
	var generation int64
	var ackSHA string
	if err := database.QueryRowContext(
		ctx, getHeadSQL, binding, "environment", environment,
	).Scan(new(string), new(string), new(string), &generation, new(string), &ackSHA, new(int64)); err != nil {
		t.Fatal(err)
	}
	if generation != 1 || ackSHA != digestC {
		t.Fatalf("head generation/ack = %d/%s", generation, ackSHA)
	}
	staleOperation := "44444444-4444-4444-8444-444444444444"
	if _, err := database.ExecContext(ctx, prepareSQL,
		staleOperation, binding, "protected_admin_change", "environment", environment,
		int64(1), nil, digestA, []byte("state-2"), digestB, []byte("proposal-2"), nil,
		now, now.Add(10*time.Minute), "proposal.v1", now); err == nil {
		t.Fatal("stale/ABA generation was accepted")
	}
	compensation := "55555555-5555-4555-8555-555555555555"
	if _, err := database.ExecContext(ctx, prepareSQL,
		compensation, binding, "compensate_uncommitted_ack", "environment", environment,
		int64(2), digestC, digestB, []byte("old-state"), digestA,
		[]byte("compensation"), digestC, now, now.Add(10*time.Minute),
		"proposal.v1", now); err != nil {
		t.Fatalf("prepare higher-generation compensation: %v", err)
	}
	exerciseSignerLifecycle(t, ctx, database, binding, environment, now,
		digestA, digestB, digestC)
}

func exerciseSignerLifecycle(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	binding, environment string,
	now time.Time,
	digestA, digestB, previousAck string,
) {
	t.Helper()
	rotation := "66666666-6666-4666-8666-666666666666"
	rotationAck := strings.Repeat("d", 64)
	newPublicDigest := strings.Repeat("e", 64)
	newTrustDigest := strings.Repeat("9", 64)
	if _, err := database.ExecContext(ctx, prepareSQL, rotation, binding,
		"proposal_signer_rotation", "environment", environment, int64(2), previousAck,
		digestB, []byte("rotated-state"), digestA,
		[]byte(`{"signer_kid":"proposal.v1"}`), nil,
		now, now.Add(10*time.Minute), "proposal.v1", now); err != nil {
		t.Fatalf("prepare signer rotation: %v", err)
	}
	if _, err := database.ExecContext(ctx, recordAckSQL, rotation, binding,
		"environment", environment, int64(2), previousAck, digestB, digestA,
		rotationAck, []byte("rotation-ack"), int64(2), now); err != nil {
		t.Fatalf("record signer rotation ack: %v", err)
	}
	if _, err := database.ExecContext(ctx, commitSQL, rotation, now); err != nil {
		t.Fatalf("commit signer rotation head: %v", err)
	}
	if _, err := database.ExecContext(ctx, rotateSignerSQL, rotation, binding,
		int64(1), "proposal.v1", int64(2), "proposal.v2", newPublicDigest,
		newTrustDigest, now); err != nil {
		t.Fatalf("activate signer rotation: %v", err)
	}
	if _, err := database.ExecContext(ctx, rotateSignerSQL, rotation, binding,
		int64(1), "proposal.v1", int64(2), "proposal.v2", newPublicDigest,
		newTrustDigest, now); err != nil {
		t.Fatalf("idempotent signer activation: %v", err)
	}
	var loadedGeneration int64
	var loadedKID, loadedPublic, loadedTrust, loadedStatus string
	var loadedLicense, loadedFinalization string
	if err := database.QueryRowContext(ctx, getBindingSQL, binding).Scan(
		new(string), new(string), new(string), new(string), &loadedLicense,
		&loadedFinalization, &loadedTrust, &loadedGeneration, &loadedKID,
		&loadedPublic, &loadedStatus,
	); err != nil || loadedGeneration != 2 || loadedKID != "proposal.v2" ||
		loadedPublic != newPublicDigest || loadedTrust != newTrustDigest ||
		loadedStatus != "active" {
		t.Fatalf("rotated binding = generation %d, kid %q, trust %q, status %q, err=%v",
			loadedGeneration, loadedKID, loadedTrust, loadedStatus, err)
	}
	if _, err := database.ExecContext(ctx, prepareSQL,
		"77777777-7777-4777-8777-777777777777", binding, "protected_admin_change",
		"environment", environment, int64(3), rotationAck, digestA, []byte("state-3"),
		digestB, []byte(`{"signer_kid":"proposal.v1"}`), nil,
		now, now.Add(10*time.Minute), "proposal.v1", now); err == nil {
		t.Fatal("retired signer prepared a mutation")
	}
	pending := "88888888-8888-4888-8888-888888888888"
	if _, err := database.ExecContext(ctx, prepareSQL, pending, binding,
		"protected_admin_change", "environment", environment, int64(3), rotationAck,
		digestA, []byte("state-3"), digestB,
		[]byte(`{"signer_kid":"proposal.v2"}`), nil,
		now, now.Add(10*time.Minute), "proposal.v2", now); err != nil {
		t.Fatalf("prepare new-signer mutation: %v", err)
	}
	pendingAck := strings.Repeat("f", 64)
	if _, err := database.ExecContext(ctx, recordAckSQL, pending, binding,
		"environment", environment, int64(3), rotationAck, digestA, digestB,
		pendingAck, []byte("pending-ack"), int64(3), now); err != nil {
		t.Fatalf("record pending ack: %v", err)
	}
	if _, err := database.ExecContext(ctx, fenceSignerSQL, binding, int64(2),
		"proposal.v2", digestA, now); err != nil {
		t.Fatalf("fence signer: %v", err)
	}
	if _, err := database.ExecContext(ctx, commitSQL, pending, now); err == nil {
		t.Fatal("fenced signer committed a pending mutation")
	}
}

const bindSQL = `SELECT platform.platformauthority_bind_v2(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

const prepareSQL = `SELECT platform.platformauthority_prepare_v2(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`

const recordAckSQL = `SELECT platform.platformauthority_record_ack_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

const commitSQL = `SELECT platform.platformauthority_commit_v1($1,$2)`

const getHeadSQL = `SELECT * FROM platform.platformauthority_get_head_v1($1,$2,$3)`

const getBindingSQL = `SELECT * FROM platform.platformauthority_get_binding_v2($1)`

const rotateSignerSQL = `SELECT platform.platformauthority_rotate_signer_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9)`

const fenceSignerSQL = `SELECT platform.platformauthority_fence_signer_v1($1,$2,$3,$4,$5)`
