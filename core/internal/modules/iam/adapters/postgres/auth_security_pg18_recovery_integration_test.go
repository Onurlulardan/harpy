//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type iamRecoveryCredentialEvidence struct {
	codeID, authenticatorID, userID string
	codeRevision, authenticatorRev  int64
	userRevision, tokenVersion      int64
	codeState, authenticatorState   string
	userStatus, identityKind, keyID string
	codeHMAC                        []byte
}

func assertIAMRecoveryTenantSubstitutionRejected(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	factors []iamAuthSecurityFactor,
) {
	t.Helper()
	if len(factors) < 2 {
		t.Fatal("recovery tenant substitution requires two tenants")
	}
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	bindIAMRuntimeSchema(t, ctx, transaction, factors[1].tenant.schema)
	_, err = readIAMRecoveryCredential(ctx, transaction, factors[0].codeHMAC)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-tenant recovery lookup err=%v", err)
	}
}

func assertIAMRecoveryConsumption(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	factor iamAuthSecurityFactor,
) {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	bindIAMRuntimeSchema(t, ctx, transaction, factor.tenant.schema)
	evidence, err := readIAMRecoveryCredential(
		ctx, transaction, factor.codeHMAC,
	)
	if err != nil || evidence.codeState != "active" ||
		evidence.authenticatorState != "active" ||
		evidence.userStatus != "active" || evidence.identityKind != "human" {
		t.Fatalf("recovery credential evidence=%+v err=%v", evidence, err)
	}
	version, err := consumeIAMRecoveryCredential(
		ctx, transaction, factor.codeHMAC, evidence,
	)
	if err != nil || version != evidence.tokenVersion+1 {
		t.Fatalf("recovery token version=%d err=%v", version, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMRecoveryConsumedState(t, ctx, database, factor, evidence)
	assertIAMRecoveryReplayRejected(t, ctx, database, factor, evidence)
}

func readIAMRecoveryCredential(
	ctx context.Context, transaction *sql.Tx, codeHMAC []byte,
) (iamRecoveryCredentialEvidence, error) {
	var evidence iamRecoveryCredentialEvidence
	err := transaction.QueryRowContext(ctx, `
SELECT code_id,authenticator_id,user_id,code_revision,
       authenticator_revision,user_revision,token_version,code_state,
       authenticator_state,user_status,identity_kind,hmac_key_id,code_hmac
FROM iam_recovery_credential_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,'iam-recovery-key-v1',$2)`,
		codeHMAC, codeHMAC,
	).Scan(
		&evidence.codeID, &evidence.authenticatorID, &evidence.userID,
		&evidence.codeRevision, &evidence.authenticatorRev,
		&evidence.userRevision, &evidence.tokenVersion, &evidence.codeState,
		&evidence.authenticatorState, &evidence.userStatus,
		&evidence.identityKind, &evidence.keyID, &evidence.codeHMAC,
	)
	return evidence, err
}

func consumeIAMRecoveryCredential(
	ctx context.Context, transaction *sql.Tx, codeHMAC []byte,
	evidence iamRecoveryCredentialEvidence,
) (int64, error) {
	var version int64
	err := transaction.QueryRowContext(ctx, `
SELECT iam_recovery_consume_v2(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		codeHMAC, evidence.keyID, codeHMAC, evidence.codeID,
		evidence.codeRevision, evidence.authenticatorID,
		evidence.authenticatorRev, evidence.userID, evidence.userRevision,
		evidence.tokenVersion,
	).Scan(&version)
	return version, err
}

func assertIAMRecoveryConsumedState(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	factor iamAuthSecurityFactor, evidence iamRecoveryCredentialEvidence,
) {
	t.Helper()
	var codeState, authenticatorState string
	var codeRevision, authenticatorRevision, userRevision, tokenVersion int64
	queryIAMAuthState(t, ctx, database.migrator, factor.tenant.schema, `
SELECT code.state,code.revision,authenticator.state,authenticator.revision,
       principal.revision,principal.token_version
FROM iam_recovery_codes AS code
JOIN iam_authenticators AS authenticator
  ON authenticator.id=code.authenticator_id
JOIN iam_users AS principal ON principal.id=code.user_id
WHERE code.id='86100000-0000-4000-8000-000000000001'::uuid`,
		&codeState, &codeRevision, &authenticatorState,
		&authenticatorRevision, &userRevision, &tokenVersion)
	if codeState != "consumed" || codeRevision != evidence.codeRevision+1 ||
		authenticatorState != "renewal_required" ||
		authenticatorRevision != evidence.authenticatorRev+1 ||
		userRevision != evidence.userRevision+1 ||
		tokenVersion != evidence.tokenVersion+1 {
		t.Fatal("recovery consumption state contract drifted")
	}
}

func assertIAMRecoveryReplayRejected(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	factor iamAuthSecurityFactor, evidence iamRecoveryCredentialEvidence,
) {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	bindIAMRuntimeSchema(t, ctx, transaction, factor.tenant.schema)
	if _, err := consumeIAMRecoveryCredential(
		ctx, transaction, factor.codeHMAC, evidence,
	); err == nil {
		t.Fatal("consumed recovery credential was replayed")
	}
}
