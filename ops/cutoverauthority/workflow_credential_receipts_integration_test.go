//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func assertCredentialReceiptsPersisted(
	t *testing.T, ctx context.Context, database workflowDatabase,
	fixture workflowFinalizationFixture,
) {
	t.Helper()
	sealDocument := readWorkflowTestFile(t, fixture.prerequisites.SessionSealReceipt)
	cleanupDocument := readWorkflowTestFile(
		t, fixture.prerequisites.CredentialCleanupReceipt,
	)
	sealSHA, err := attestation.DigestHex(sealDocument)
	if err != nil {
		t.Fatal(err)
	}
	cleanupSHA, err := attestation.DigestHex(cleanupDocument)
	if err != nil {
		t.Fatal(err)
	}
	var storedSealSHA, storedCleanupSHA string
	var storedSeal, storedCleanup []byte
	err = database.admin.QueryRow(ctx, `SELECT seal_receipt_sha256,
seal_receipt_document,cleanup_receipt_sha256,cleanup_receipt_document
FROM platform.cutover_credential_receipts WHERE operation_id=$1 AND run_id=$2`,
		fixture.ids.operation, fixture.ids.run,
	).Scan(&storedSealSHA, &storedSeal, &storedCleanupSHA, &storedCleanup)
	if err != nil || storedSealSHA != sealSHA || storedCleanupSHA != cleanupSHA ||
		!bytes.Equal(storedSeal, sealDocument) || !bytes.Equal(storedCleanup, cleanupDocument) {
		t.Fatalf("credential receipt inventory differs: err=%v", err)
	}
}

func assertCredentialReceiptInventoryCount(
	t *testing.T, ctx context.Context, database workflowDatabase, expected int,
) {
	t.Helper()
	var count int
	if err := database.admin.QueryRow(ctx,
		"SELECT count(*)::integer FROM platform.cutover_credential_receipts",
	).Scan(&count); err != nil || count != expected {
		t.Fatalf("credential receipt count differs: got=%d want=%d err=%v",
			count, expected, err)
	}
}

func readWorkflowTestFile(t *testing.T, path string) []byte {
	t.Helper()
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
