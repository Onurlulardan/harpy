//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func prepareBillingOrigin(
	t *testing.T,
	ctx context.Context,
	origin string,
) *sql.DB {
	t.Helper()
	return prepareBillingOriginThrough(t, ctx, origin, 0)
}

func prepareBillingOriginThrough(
	t *testing.T,
	ctx context.Context,
	origin string,
	maxVersion int64,
) *sql.DB {
	t.Helper()
	database := startBillingPG18(t, ctx, origin)
	prepareBillingPrerequisites(t, ctx, database, origin)
	documents := loadBillingDocuments(t, origin)
	if maxVersion > 0 {
		documents = billingDocumentsThrough(t, documents, maxVersion)
	}
	applyBillingDocuments(t, ctx, database, documents)
	return database
}

func billingRuntimeConnection(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		t.Fatalf("set Billing runtime session: %v", err)
	}
	if _, err = connection.ExecContext(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatalf("set Billing owner role: %v", err)
	}
	return connection
}

func billingGlobalAuthorityID(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) uuid.UUID {
	t.Helper()
	var authorityID uuid.UUID
	if err := database.QueryRowContext(ctx, `
SELECT authority_id FROM platform.billing_global_authorities
`).Scan(&authorityID); err != nil {
		t.Fatal(err)
	}
	return authorityID
}

func billingV7(t *testing.T) uuid.UUID {
	t.Helper()
	identifier, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return identifier
}

func billingDigest(value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return digest[:]
}

func billingPGCode(err error) string {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code
	}
	return ""
}

func billingTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 5*time.Minute)
}
