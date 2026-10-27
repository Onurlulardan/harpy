//go:build integration

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIssueAndAuthorizeAllTerminalStatesPG184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkflowDatabase(t, ctx)
	directory := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	authority := adoptWorkflowAuthority(t, ctx, database, directory, now)
	assertClusterAdminCannotObserveAuthIDVerifiers(t, ctx, database)
	var failedFixture workflowFinalizationFixture
	for _, state := range []string{"completed", "failed", "cancelled"} {
		t.Run(state, func(t *testing.T) {
			fixture := prepareWorkflowFinalizationFixture(
				t, database, authority, directory, state, now,
			)
			if state == "failed" {
				failedFixture = fixture
			}
			issueResult, err := executeFinalizerIssuance(ctx, fixture.issueInvocation())
			if err != nil {
				t.Fatalf("issue %s finalizer lease: %v", state, err)
			}
			assertCredentialReceiptsPersisted(t, ctx, database, fixture)
			leaseDocument, signatureDocument := readIssuedWorkflowLease(t, fixture)
			assertLeaseTenantAuthority(t, leaseDocument, state)
			if state == "failed" {
				if err := os.Remove(fixture.signaturePath); err != nil {
					t.Fatal(err)
				}
				recovered, err := executeFinalizerIssuance(ctx, fixture.issueInvocation())
				if err != nil || recovered.Replayed ||
					recovered.EvidenceSHA256 != issueResult.EvidenceSHA256 {
					t.Fatalf("recover partial %s publication: result=%+v err=%v",
						state, recovered, err)
				}
			}
			replayResult, err := executeFinalizerIssuance(ctx, fixture.issueInvocation())
			if err != nil || replayResult.EvidenceSHA256 != issueResult.EvidenceSHA256 {
				t.Fatalf("replay %s issuance: result=%+v err=%v", state, replayResult, err)
			}
			if issueResult.Replayed || !replayResult.Replayed {
				t.Fatalf("%s issuance replay flags differ: first=%v replay=%v",
					state, issueResult.Replayed, replayResult.Replayed)
			}
			assertCredentialReceiptsPersisted(t, ctx, database, fixture)
			replayedLease, replayedSignature := readIssuedWorkflowLease(t, fixture)
			if !bytes.Equal(leaseDocument, replayedLease) ||
				!bytes.Equal(signatureDocument, replayedSignature) {
				t.Fatalf("%s issuance replay bytes differ", state)
			}
			authorization := fixture.authorizeInvocation()
			if _, err := executeFinalizationAuthorization(ctx, authorization); err != nil {
				t.Fatalf("authorize %s finalization: %v", state, err)
			}
			if _, err := executeFinalizationAuthorization(ctx, authorization); err != nil {
				t.Fatalf("replay %s authorization: %v", state, err)
			}
		})
	}
	mismatch := prepareMismatchedTerminal(t, directory, failedFixture)
	if _, err := executeFinalizerIssuance(ctx, mismatch.issueInvocation()); err != nil {
		t.Fatalf("issue valid mismatched replay lease: %v", err)
	}
	if _, err := executeFinalizationAuthorization(
		ctx, mismatch.authorizeInvocation(),
	); err == nil {
		t.Fatal("protected authorization accepted a different replay payload")
	}
	var count int
	if err := database.admin.QueryRow(ctx,
		"SELECT count(*)::integer FROM cutover_authority.finalization_authorizations",
	).Scan(&count); err != nil || count != 3 {
		t.Fatalf("unexpected authorization inventory: count=%d err=%v", count, err)
	}
	assertCredentialReceiptInventoryCount(t, ctx, database, 3)
}

func assertClusterAdminCannotObserveAuthIDVerifiers(
	t *testing.T, ctx context.Context, database workflowDatabase,
) {
	t.Helper()
	connection := connectWorkflowRole(t, ctx, database, "harpy_cluster_admin")
	defer connection.Close(context.Background())
	_, err := connection.Exec(ctx, "SELECT count(*) FROM pg_catalog.pg_authid")
	assertPGAuthIDDenied(t, err)
	rows, queryErr := connection.Query(
		ctx, "SELECT rolpassword FROM pg_catalog.pg_authid LIMIT 1",
	)
	if queryErr != nil {
		assertPGAuthIDDenied(t, queryErr)
		return
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("cluster admin observed a pg_authid verifier row")
	}
	assertPGAuthIDDenied(t, rows.Err())
}

func assertPGAuthIDDenied(t *testing.T, err error) {
	t.Helper()
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "42501" {
		t.Fatalf("cluster admin pg_authid read was not denied: %v", err)
	}
}

func readIssuedWorkflowLease(
	t *testing.T, fixture workflowFinalizationFixture,
) ([]byte, []byte) {
	t.Helper()
	manifest, err := os.ReadFile(fixture.leasePath)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := os.ReadFile(fixture.signaturePath)
	if err != nil {
		t.Fatal(err)
	}
	return manifest, signature
}

func assertLeaseTenantAuthority(
	t *testing.T, document []byte, state string,
) {
	t.Helper()
	var lease cutoverv1.FinalizerLease
	if err := attestation.DecodeStrict(document, &lease); err != nil {
		t.Fatal(err)
	}
	if lease.TerminalState != state {
		t.Fatalf("finalizer state differs: %+v", lease)
	}
	if state == "completed" {
		if lease.TenantRequirementCatalogSHA256 == nil ||
			lease.TenantRequirementCount == nil || *lease.TenantRequirementCount != 1 {
			t.Fatalf("completed lease omitted tenant authority: %+v", lease)
		}
		return
	}
	if lease.TenantRequirementCatalogSHA256 != nil || lease.TenantRequirementCount != nil {
		t.Fatalf("abort lease retained tenant authority: %+v", lease)
	}
}

func prepareMismatchedTerminal(
	t *testing.T, directory string, fixture workflowFinalizationFixture,
) workflowFinalizationFixture {
	t.Helper()
	document, err := os.ReadFile(fixture.prerequisites.TerminalEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var terminal cutoverv1.TerminalEvidence
	if err := attestation.DecodeStrict(document, &terminal); err != nil {
		t.Fatal(err)
	}
	code := "target_failed"
	terminal.OutcomeCode = &code
	path, _ := writeWorkflowArtifact(t, directory, "mismatch-terminal.json", terminal)
	fixture.prerequisites.TerminalEvidence = path
	fixture.leasePath = filepath.Join(directory, "mismatch-finalizer-lease.json")
	fixture.signaturePath = filepath.Join(directory, "mismatch-finalizer-lease.sig")
	return fixture
}
