//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

func TestSupportAdoptionFreshOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startSupportDatabase(t, ctx)
	resetAndApplySupport(t, ctx, database)
	locators := createLegacySupportRelations(t, ctx, database)
	store, adopter := adopterStore(t, ctx, database)
	clock := &adoptionClock{value: time.Date(2026, 8, 1, 12, 0, 0, 123456000, time.UTC)}
	exerciseContactCrashResume(t, ctx, store, clock, locators.contact)
	exerciseBugTenantAdoption(t, ctx, store, clock, locators.bugA,
		uuid.MustParse("30000000-0000-4000-8000-000000000001"))
	exerciseBugTenantAdoption(t, ctx, store, clock, locators.bugB,
		uuid.MustParse("40000000-0000-4000-8000-000000000001"))
	assertAdoptionPartitions(t, ctx, database)
	exerciseTargetChecksumMismatch(t, ctx, database, store, clock, locators.contact)
	exerciseChecksumMismatch(t, ctx, store, clock, locators.contact)
	exerciseLocatorMismatch(t, ctx, store, clock, locators.contact)
	assertAdopterPrivileges(t, ctx, adopter)
}

type legacyLocators struct {
	contact, bugA, bugB application.VerifiedLocator
}

func createLegacySupportRelations(t *testing.T, ctx context.Context,
	database *sql.DB,
) legacyLocators {
	t.Helper()
	if _, err := database.ExecContext(ctx, `
        DROP TABLE IF EXISTS public.contact_messages;
        DROP SCHEMA IF EXISTS tenant_support_a CASCADE;
        DROP SCHEMA IF EXISTS tenant_support_b CASCADE;
        CREATE TABLE public.contact_messages(id uuid PRIMARY KEY);
        CREATE SCHEMA tenant_support_a;
        CREATE SCHEMA tenant_support_b;
        CREATE TABLE tenant_support_a.bug_reports(id uuid PRIMARY KEY);
        CREATE TABLE tenant_support_b.bug_reports(id uuid PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	tenantA := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	tenantB := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	return legacyLocators{
		contact: verifiedLocator(t, database, ctx, application.ContactGlobal,
			uuid.Nil, "public", "contact_messages", "", ""),
		bugA: verifiedLocator(t, database, ctx, application.BugTenant,
			tenantA, "tenant_support_a", "bug_reports", "tenant-a", "Tenant A"),
		bugB: verifiedLocator(t, database, ctx, application.BugTenant,
			tenantB, "tenant_support_b", "bug_reports", "tenant-b", "Tenant B")}
}

func verifiedLocator(t *testing.T, database *sql.DB, ctx context.Context,
	kind application.SourceKind, tenantID uuid.UUID, schema, relation, slug, name string,
) application.VerifiedLocator {
	t.Helper()
	var oid uint32
	if err := database.QueryRowContext(ctx, `SELECT relation.oid::bigint
        FROM pg_catalog.pg_class relation JOIN pg_catalog.pg_namespace namespace
        ON namespace.oid=relation.relnamespace
        WHERE namespace.nspname=$1 AND relation.relname=$2`, schema, relation).Scan(&oid); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(schema + "." + relation))
	locator, err := application.NewVerifiedLocator(kind, tenantID, schema, relation,
		oid, slug, name, digest)
	if err != nil {
		t.Fatal(err)
	}
	return locator
}

func exerciseContactCrashResume(t *testing.T, ctx context.Context,
	store application.Target, clock *adoptionClock, locator application.VerifiedLocator,
) {
	t.Helper()
	rows := []application.ContactSourceRow{
		contactRow("01000000-0000-4000-8000-000000000001", "Ada", clock.value),
		contactRow("02000000-0000-4000-8000-000000000001", "Lin", clock.value.Add(time.Microsecond))}
	source := contactSourceFixture(t, locator, rows, "contact-crash-resume")
	worker, err := application.NewWorker(source, store, clock)
	if err != nil {
		t.Fatal(err)
	}
	operationID, runID, leaseOwner := uuid.New(), uuid.Must(uuid.NewV7()), uuid.New()
	run, replayed, err := worker.Start(ctx, application.StartRequest{
		OperationID: operationID, RunID: runID, Locator: locator})
	if err != nil || replayed {
		t.Fatalf("start contact adoption replay=%v err=%v", replayed, err)
	}
	crashing, _ := application.NewWorker(source, &crashBeforeAdvance{Target: store}, clock)
	_, err = crashing.Execute(ctx, application.ProcessRequest{OperationID: operationID,
		RunID: runID, LeaseOwner: leaseOwner, ExpectedRevision: run.Revision})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("crash point error=%v", err)
	}
	completed, err := worker.Execute(ctx, application.ProcessRequest{OperationID: operationID,
		RunID: runID, LeaseOwner: leaseOwner, ExpectedRevision: run.Revision})
	if err != nil || completed.Status != application.RunCompleted || completed.CopiedCount != 2 {
		t.Fatalf("resume result=%s copied=%d err=%v", completed.Status, completed.CopiedCount, err)
	}
	replayedRun, replayed, err := worker.Start(ctx, application.StartRequest{
		OperationID: operationID, RunID: runID, Locator: locator})
	if err != nil || !replayed || replayedRun.Status != application.RunCompleted {
		t.Fatalf("start replay status=%s replay=%v err=%v", replayedRun.Status, replayed, err)
	}
}

func exerciseBugTenantAdoption(t *testing.T, ctx context.Context,
	store application.Target, clock *adoptionClock, locator application.VerifiedLocator,
	sourceID uuid.UUID,
) {
	t.Helper()
	source := bugSourceFixture(t, locator,
		[]application.BugSourceRow{bugRow(sourceID, locator.TenantID(), clock.value)},
		"bug-"+locator.TenantSlug())
	worker, _ := application.NewWorker(source, store, clock)
	run, _, err := worker.Start(ctx, application.StartRequest{
		OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()), Locator: locator})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := worker.Execute(ctx, application.ProcessRequest{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: uuid.New(), ExpectedRevision: run.Revision})
	if err != nil || completed.Status != application.RunCompleted || completed.TenantID != locator.TenantID() {
		t.Fatalf("bug tenant adoption status=%s tenant=%s err=%v",
			completed.Status, completed.TenantID, err)
	}
}

func exerciseChecksumMismatch(t *testing.T, ctx context.Context,
	store application.Target, clock *adoptionClock, locator application.VerifiedLocator,
) {
	t.Helper()
	row := contactRow("05000000-0000-4000-8000-000000000001", "Drift", clock.value)
	source := contactSourceFixture(t, locator, []application.ContactSourceRow{row}, "contact-drift")
	source.manifest.RootSHA256 = sha256.Sum256([]byte("wrong-root"))
	worker, _ := application.NewWorker(source, store, clock)
	run, _, err := worker.Start(ctx, application.StartRequest{
		OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()), Locator: locator})
	if err != nil {
		t.Fatal(err)
	}
	leaseOwner := uuid.New()
	_, err = worker.Execute(ctx, application.ProcessRequest{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, ExpectedRevision: run.Revision})
	if !errors.Is(err, application.ErrSourceDrift) {
		t.Fatalf("worker checksum error=%v", err)
	}
	claimed, err := store.AdoptionClaim(ctx, application.ClaimCommand{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, ExpectedRevision: 3, ObservedAt: clock.value})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.AdoptionReconcile(ctx, application.LeaseCommand{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: leaseOwner, LeaseGeneration: claimed.LeaseGeneration,
		ObservedAt: clock.value})
	if !errors.Is(err, application.ErrSourceDrift) {
		t.Fatalf("database checksum error=%v", err)
	}
}
