//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertTenantAuthorityCannotWriteGlobal(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	globalAuthorityID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
) {
	t.Helper()
	tenantAuthorityID := billingV7(t)
	if _, err := connection.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,$3,$4,$5,$6)
`, tenantAuthorityID, uuid.MustParse(billingTenantID), int64(1), int64(1),
		int64(1), now); err != nil {
		t.Fatalf("register tenant Billing authority: %v", err)
	}
	if tenantAuthorityID == globalAuthorityID {
		t.Fatal("tenant and global authority identities collided")
	}
	call := callCreateOffering(ctx, connection, tenantAuthorityID, actorID,
		"tenant-global-denial", billingDigest("tenant-global-denial"), now)
	if billingPGCode(call.err) != "P5721" {
		t.Fatalf("tenant-to-global denial code=%q err=%v",
			billingPGCode(call.err), call.err)
	}
}
