//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingLicenseEvidenceImmutable(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	statements := []string{
		`UPDATE platform.billing_license_snapshots SET status='active'
WHERE snapshot_id=(SELECT snapshot_id FROM platform.billing_license_heads LIMIT 1)`,
		`UPDATE platform.billing_license_snapshot_projections SET boolean_value=false
WHERE value_type='boolean'`,
		`UPDATE platform.billing_license_verification_receipts SET outcome='noop'
WHERE outcome='accepted'`,
		`UPDATE platform.billing_license_restore_receipts SET restore_receipt_sha256=
repeat('f',64)`,
		`UPDATE platform.billing_license_offering_publications SET offering_sha256=
repeat('f',64)`,
	}
	for _, statement := range statements {
		if _, err := connection.ExecContext(ctx, statement); billingPGCode(err) != "55000" {
			t.Fatalf("mutable billing license evidence code=%q err=%v",
				billingPGCode(err), err)
		}
	}
}
