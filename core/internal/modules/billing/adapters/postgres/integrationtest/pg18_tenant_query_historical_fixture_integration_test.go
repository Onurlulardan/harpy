//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func retireTenantQueryOffering(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	authorityID, versionID uuid.UUID,
) {
	t.Helper()
	var offeringID uuid.UUID
	if err := connection.QueryRowContext(ctx, `
SELECT offering_id FROM platform.billing_offering_versions
WHERE authority_id=$1 AND id=$2
`, authorityID, versionID).Scan(&offeringID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	var count int
	if err := connection.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_retire_offering_v1(
 $1,$2,$3,$4,$5,$6,1,$7,$8
)
`, authorityID, billingV7(t), offeringID, uuid.New(), "retire-query-fixture",
		billingDigest("retire-query-fixture"), "historical fixture", now,
	).Scan(&count); err != nil || count != 1 {
		t.Fatalf("retire tenant query offering count=%d err=%v", count, err)
	}
}
