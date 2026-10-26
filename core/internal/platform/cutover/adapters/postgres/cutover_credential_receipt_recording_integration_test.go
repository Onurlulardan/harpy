//go:build integration

package postgres

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

const recordCredentialReceiptsSQL = `SELECT cutover_authority.record_credential_receipts_v1(
    $1::uuid,$2::uuid,$3::text,$4::bytea,$5::text,$6::bytea)`

func verifyProtectedCredentialReceiptRace(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
) {
	t.Helper()
	setClusterAdminLease(
		t, ctx, database.admin, time.Now().UTC().Add(110*time.Second),
	)
	first := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	second := connectClusterAdminSession(t, ctx, database.dsn)
	t.Cleanup(func() {
		_ = first.Close(context.Background())
		_ = second.Close(context.Background())
	})
	errors := raceCredentialReceiptRecording(
		ctx, first, second, adoption, seal, cleanup,
	)
	for _, err := range errors {
		if err != nil {
			t.Fatalf("idempotent credential receipt race failed: %v", errors)
		}
	}
	forged := replaceNestedJSONField(
		t, seal.document, "payload", "receipt_id",
		"39000000-0000-4000-8000-000000000099",
	)
	forgedSHA256, err := attestation.DigestHex(forged)
	if err != nil {
		t.Fatal(err)
	}
	if err := recordCredentialReceipts(
		ctx, first, adoption, builtCutoverArtifact{
			document: forged, digest: forgedSHA256,
		}, cleanup,
	); err == nil {
		t.Fatal("different credential receipt replay succeeded")
	}
	_ = first.Close(ctx)
	_ = second.Close(ctx)
}

func raceCredentialReceiptRecording(
	ctx context.Context, first, second *pgx.Conn,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
) [2]error {
	start := make(chan struct{})
	results := [2]error{}
	connections := []*pgx.Conn{first, second}
	var group sync.WaitGroup
	for index := range connections {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			results[index] = recordCredentialReceipts(
				ctx, connections[index], adoption, seal, cleanup,
			)
		}(index)
	}
	close(start)
	group.Wait()
	return results
}

func recordCredentialReceipts(
	ctx context.Context, connection *pgx.Conn,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
) error {
	_, err := connection.Exec(ctx, recordCredentialReceiptsSQL,
		adoption.identity.operationID, adoption.identity.runID,
		seal.digest, seal.document, cleanup.digest, cleanup.document,
	)
	return err
}

func connectClusterAdminSession(
	t *testing.T, ctx context.Context, dsn string,
) *pgx.Conn {
	t.Helper()
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Exec(
		ctx, "SET SESSION AUTHORIZATION harpy_cluster_admin",
	); err != nil {
		_ = connection.Close(ctx)
		t.Fatal(err)
	}
	return connection
}
