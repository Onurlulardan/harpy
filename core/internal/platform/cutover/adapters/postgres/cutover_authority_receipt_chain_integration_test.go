//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestProtectedFinalizationReceiptSurvivesAuthorityGenerationUpgradePG18(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	authorityBinary := buildCutoverAuthorityBinary(t, ctx)
	migratorBinary := buildCutoverMigratorBinary(t, ctx)
	directory := t.TempDir()
	now := time.Now().UTC()
	signer := prepareBuiltCutoverSigner(t, directory, now)
	seedReleaseReadiness(t, ctx, database.admin)
	addLastTenantCheckpoint(t, ctx, database.admin)
	assertObserverRequiresAdoptedCatalog(t, ctx, database)
	authority := buildBuiltAuthorityArtifacts(t, ctx, database, signer, directory, now)
	authority = adoptBuiltCutoverAuthority(
		t, ctx, authorityBinary, database, signer, authority, directory,
	)
	evidence := verifyProtectedFinalizationState(
		t, ctx, database, authorityBinary, migratorBinary, signer,
		authority, 95, "cancelled", false, protectedFinalizationHooks{},
	)
	upgradeDirectory := t.TempDir()
	upgrade := buildBuiltAuthorityUpgradeArtifacts(
		t, ctx, database, signer, authority, upgradeDirectory, time.Now().UTC(),
	)
	_ = adoptBuiltCutoverAuthority(
		t, ctx, authorityBinary, database, signer, upgrade, upgradeDirectory,
	)
	leaseBuiltFinalizerRole(t, ctx, database, time.Now().UTC().Add(2*time.Minute))
	verifier, err := pgx.Connect(ctx, builtCutoverRoleDSN(
		t, database.dsn, "harpy_cutover_finalizer", builtCutoverPassword,
	))
	if err != nil {
		t.Fatalf("connect historical receipt verifier: %v", err)
	}
	defer verifier.Close(ctx)
	assertProtectedReceiptVerification(t, ctx, verifier, evidence)
	assertHistoricalReceiptChainTamperRejected(t, ctx, database, verifier, evidence)
	assertMissingHistoricalReceiptGenerationRejected(t, ctx, database, verifier, evidence)
}

func assertHistoricalReceiptChainTamperRejected(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	verifier *pgx.Conn, evidence finalizationEvidence,
) {
	t.Helper()
	var originalDocument []byte
	var originalSHA256 string
	if err := database.admin.QueryRow(ctx, `SELECT receipt_document,receipt_sha256
FROM cutover_authority.authority_adoption_inventory WHERE catalog_generation=2`).Scan(
		&originalDocument, &originalSHA256,
	); err != nil {
		t.Fatal(err)
	}
	_, err := database.admin.Exec(ctx, `WITH changed AS (
SELECT pg_catalog.convert_to((pg_catalog.convert_from(receipt_document,'UTF8')::jsonb ||
  pg_catalog.jsonb_build_object('previous_authority_receipt_sha256',
    pg_catalog.repeat('f',64)))::text,'UTF8') AS document
FROM cutover_authority.authority_adoption_inventory WHERE catalog_generation=2)
UPDATE cutover_authority.authority_adoption_inventory AS inventory
SET receipt_document=changed.document,
    receipt_sha256=pg_catalog.encode(pg_catalog.sha256(changed.document),'hex')
FROM changed WHERE inventory.catalog_generation=2`)
	if err != nil {
		t.Fatalf("tamper generation-two receipt chain: %v", err)
	}
	err = protectedReceiptVerificationError(ctx, verifier, evidence)
	assertFinalizationError(t, err, "55000", "authority receipt chain differs")
	_, err = database.admin.Exec(ctx, `UPDATE cutover_authority.authority_adoption_inventory
SET receipt_document=$1,receipt_sha256=$2 WHERE catalog_generation=2`,
		originalDocument, originalSHA256)
	if err != nil {
		t.Fatalf("restore generation-two receipt chain: %v", err)
	}
}

func assertMissingHistoricalReceiptGenerationRejected(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	verifier *pgx.Conn, evidence finalizationEvidence,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `CREATE TEMP TABLE saved_authority_generation_one
ON COMMIT PRESERVE ROWS AS SELECT * FROM cutover_authority.authority_adoption_inventory
WHERE catalog_generation=1;
DELETE FROM cutover_authority.authority_adoption_inventory WHERE catalog_generation=1`)
	if err != nil {
		t.Fatalf("remove historical receipt generation: %v", err)
	}
	err = protectedReceiptVerificationError(ctx, verifier, evidence)
	assertFinalizationError(t, err, "55000", "authority receipt chain differs")
	_, err = database.admin.Exec(ctx, `INSERT INTO cutover_authority.authority_adoption_inventory
SELECT * FROM saved_authority_generation_one;
DROP TABLE saved_authority_generation_one`)
	if err != nil {
		t.Fatalf("restore historical receipt generation: %v", err)
	}
}

func protectedReceiptVerificationError(
	ctx context.Context, connection *pgx.Conn, evidence finalizationEvidence,
) error {
	var verified bool
	return connection.QueryRow(ctx, `SELECT
cutover_authority.verify_finalization_receipt_v1($1::uuid,$2::uuid,$3::uuid,$4::text,$5::text)`,
		evidence.operationID, evidence.runID, evidence.leaseID,
		evidence.terminalState, evidence.terminalSHA256,
	).Scan(&verified)
}
