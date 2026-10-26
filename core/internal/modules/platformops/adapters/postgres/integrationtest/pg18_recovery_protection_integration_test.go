//go:build integration

package postgres_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"
)

func exerciseRecoveryProtection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	now := time.Date(2026, 7, 23, 11, 30, 0, 0, time.UTC)
	assertProtectionState(t, ctx, database, "unprotected", 0)
	operationID := "12121212-1212-4212-8212-121212121212"
	requestDigest := bytes.Repeat([]byte{0x41}, 32)
	publicKey := bytes.Repeat([]byte{0x42}, 32)
	request := `SELECT * FROM platform.platformops_request_recovery_control_v1(
		$1,$2,$3,$4,$5,'enable',$6,$7,$8)`
	for replay := 0; replay < 2; replay++ {
		var operation, state, claim string
		err := database.QueryRowContext(ctx, request, operationID, environmentID, ownerID,
			"recovery-enable-0001", requestDigest,
			"HPKE-v1-X25519-HKDF-SHA256-AES-256-GCM", publicKey, now).
			Scan(&operation, &state, &claim)
		if err != nil || operation != operationID || state != "accepted" || claim != "pending" {
			t.Fatalf("request recovery protection replay=%d: %v", replay, err)
		}
	}
	artifact := []byte("hpke-sealed-environment-recovery-kit")
	proof := []byte("owner-possession-proof")
	var published bool
	err := database.QueryRowContext(ctx,
		`SELECT platform.platformops_publish_recovery_kit_v1(
		$1,$2,sha256($2),sha256($3),$4,$5)`, operationID, artifact, proof,
		now.Add(5*time.Minute), now.Add(time.Minute)).Scan(&published)
	if err != nil || !published {
		t.Fatalf("publish recovery kit: %v", err)
	}
	sessionID := "13131313-1313-4313-8313-131313131313"
	var claimed []byte
	var digest string
	var expires time.Time
	err = database.QueryRowContext(ctx,
		`SELECT * FROM platform.platformops_claim_recovery_kit_v1($1,$2,$3,$4)`,
		operationID, ownerID, sessionID, now.Add(2*time.Minute)).
		Scan(&claimed, &digest, &expires)
	if err != nil || !bytes.Equal(claimed, artifact) || digest != sha256Hex(artifact) {
		t.Fatalf("claim recovery kit: %v", err)
	}
	var acknowledged bool
	err = database.QueryRowContext(ctx,
		`SELECT platform.platformops_ack_recovery_kit_v1(
		$1,$2,$3,decode($4,'hex'),$5,$6)`, operationID, ownerID, sessionID,
		digest, proof, now.Add(3*time.Minute)).Scan(&acknowledged)
	if err != nil || !acknowledged {
		t.Fatalf("ack recovery kit: %v", err)
	}
	assertProtectionState(t, ctx, database, "protected", 1)
}

func assertProtectionState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	wantState string,
	wantGeneration int64,
) {
	t.Helper()
	var environment, state string
	var generation int64
	var acknowledged bool
	var updated time.Time
	err := database.QueryRowContext(ctx,
		`SELECT * FROM platform.platformops_get_recovery_protection_v1($1)`,
		environmentID).Scan(&environment, &state, &generation, &acknowledged, &updated)
	if err != nil || environment != environmentID || state != wantState ||
		generation != wantGeneration || acknowledged != (wantGeneration > 0) {
		t.Fatalf("protection=%s/%d/%t: %v", state, generation, acknowledged, err)
	}
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
