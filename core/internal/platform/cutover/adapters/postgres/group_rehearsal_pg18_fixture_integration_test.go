//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func seedBillingFormsRehearsalGroup(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	value grouprehearsal.PersistedRehearsal, complete bool,
) {
	t.Helper()
	members := []string{"billing"}
	if complete {
		members = append(members, "forms")
	}
	for _, domain := range members {
		_, err := admin.Exec(ctx, `INSERT INTO platform.domain_cutovers (
scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,$2,'shadow','v1',$3,$4,'billing Forms PG18','PHASE6-BF-1')`,
			value.Evidence.TenantID, domain, value.Evidence.Generation, integrationOwnerID)
		if err != nil {
			t.Fatalf("seed %s domain: %v", domain, err)
		}
	}
	_, err := admin.Exec(ctx, `INSERT INTO platform.cutover_groups (
scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,'billing_forms','shadow','v1',$2,$3,
'billing Forms PG18','PHASE6-BF-1')`, value.Evidence.TenantID,
		value.Evidence.Generation, integrationOwnerID)
	if err != nil {
		t.Fatalf("seed billing Forms group: %v", err)
	}
	for ordinal, domain := range members {
		_, err = admin.Exec(ctx, `INSERT INTO platform.cutover_group_members (
scope_kind,scope_id,group_id,ordinal,domain
) VALUES ('tenant',$1,'billing_forms',$2,$3)`,
			value.Evidence.TenantID, ordinal+1, domain)
		if err != nil {
			t.Fatalf("seed billing Forms member %s: %v", domain, err)
		}
	}
}

func leaseGroupRehearsalAdopter(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) {
	t.Helper()
	expiresAt := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339Nano)
	statement := fmt.Sprintf(`ALTER ROLE harpy_adopter LOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
CONNECTION LIMIT 1 PASSWORD 'integration-leased-secret' VALID UNTIL '%s'`, expiresAt)
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("lease billing Forms adopter: %v", err)
	}
}

func groupRehearsalWithIdentity(
	tenantID, receiptID string,
) grouprehearsal.PersistedRehearsal {
	value := groupRehearsalFixture()
	value.Evidence.TenantID = uuid.MustParse(tenantID)
	value.ReceiptID = uuid.MustParse(receiptID)
	return value
}

func requireGroupRehearsalSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("billing Forms rehearsal error = %v, want SQLSTATE %s", err, want)
	}
}
